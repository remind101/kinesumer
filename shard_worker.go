package kinesumer

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesis"
	k "github.com/remind101/kinesumer/interface"
)

type ShardWorker struct {
	kinesis         k.Kinesis
	shard           *kinesis.Shard
	checkpointer    k.Checkpointer
	stream          *string
	pollTime        int
	sequence        *string
	stop            <-chan Unit
	stopped         chan<- Unit
	c               chan *k.KinesisRecord
	provisioner     k.Provisioner
	handlers        k.KinesumerHandlers
	GetRecordsLimit int64
}

func (s *ShardWorker) GetShardIterator(iteratorType string, sequence *string) (*string, error) {
	iter, err := s.kinesis.GetShardIterator(&kinesis.GetShardIteratorInput{
		ShardID:                s.shard.ShardID,
		ShardIteratorType:      &iteratorType,
		StartingSequenceNumber: sequence,
		StreamName:             s.stream,
	})
	if err != nil {
		return nil, err
	}
	return iter.ShardIterator, nil
}

func (s *ShardWorker) TryGetShardIterator(iteratorType string, sequence *string) *string {
	it, err := s.GetShardIterator(iteratorType, sequence)
	if err != nil {
		panic(err)
	}
	return it
}

func (s *ShardWorker) GetRecords(it *string) ([]*kinesis.Record, *string, int64, error) {
	resp, err := s.kinesis.GetRecords(&kinesis.GetRecordsInput{
		Limit:         &s.GetRecordsLimit,
		ShardIterator: it,
	})
	if err != nil {
		return nil, nil, 0, err
	}
	return resp.Records, resp.NextShardIterator, *resp.MillisBehindLatest, nil
}

func (s *ShardWorker) GetRecordsAndProcess(it, sequence *string) (cont bool, nextIt *string, nextSeq *string) {
	records, nextIt, lag, err := s.GetRecords(it)
	if err != nil || len(records) == 0 {
		if err != nil {
			s.c <- &k.KinesisRecord{
				ShardID:            s.shard.ShardID,
				MillisBehindLatest: lag,
				Err:                err,
			}
			nextIt = s.TryGetShardIterator("AFTER_SEQUENCE_NUMBER", sequence)
		}

		if err := s.provisioner.Heartbeat(*s.shard.ShardID); err != nil {
			s.handlers.Err(k.NewKinesumerError(k.KinesumerEError, "Heartbeat failed", err))
			return true, nil, sequence
		}
		// GetRecords is not guaranteed to return records even if there are records to be read.
		// However, if our lag time behind the shard head is <= 3 seconds then there's probably
		// no records.
		if lag <= 3000 /* milliseconds */ {
			select {
			case <-time.NewTimer(time.Duration(s.pollTime) * time.Millisecond).C:
			case <-s.stop:
				return true, nil, sequence
			}
		}
	} else {
		for _, rec := range records {
			s.c <- &k.KinesisRecord{
				Record:             *rec,
				ShardID:            s.shard.ShardID,
				CheckpointC:        s.checkpointer.DoneC(),
				MillisBehindLatest: lag,
			}

			if err := s.provisioner.Heartbeat(*s.shard.ShardID); err != nil {
				s.handlers.Err(k.NewKinesumerError(k.KinesumerEError, "Heartbeat failed", err))
				return true, nil, sequence
			}
		}
		sequence = records[len(records)-1].SequenceNumber
	}
	return false, nextIt, sequence
}

func (s *ShardWorker) RunWorker() {
	defer func() {
		s.stopped <- Unit{}
	}()

	sequence := s.checkpointer.GetStartSequence(s.shard.ShardID)
	end := s.shard.SequenceNumberRange.EndingSequenceNumber
	var it *string
	if sequence == nil || len(*sequence) == 0 {
		sequence = s.shard.SequenceNumberRange.StartingSequenceNumber

		s.c <- &k.KinesisRecord{
			ShardID: s.shard.ShardID,
			Err:     k.NewKinesumerError(k.KinesumerEInfo, "Using TRIM_HORIZON", nil),
		}
		it = s.TryGetShardIterator("TRIM_HORIZON", nil)
	} else {
		it = s.TryGetShardIterator("AFTER_SEQUENCE_NUMBER", sequence)
	}

loop:
	for {
		if err := s.provisioner.Heartbeat(*s.shard.ShardID); err != nil {
			s.handlers.Err(k.NewKinesumerError(k.KinesumerEError, "Heartbeat failed", err))
			break loop
		}

		if end != nil && *sequence == *end {
			s.c <- &k.KinesisRecord{
				ShardID: s.shard.ShardID,
				Err:     k.NewKinesumerError(k.KinesumerEInfo, "Shard has reached its end", nil),
			}
			break loop
		}

		select {
		case <-s.stop:
			break loop
		default:
			if brk, nextIt, seq := s.GetRecordsAndProcess(it, sequence); brk {
				break loop
			} else {
				it = nextIt
				sequence = seq
			}
		}
	}
}
