package kinesumer

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/remind101/pkg/logger"
)

type ShardWorker struct {
	kinesis         KinesisAPI
	logger          logger.Logger
	shard           *kinesis.Shard
	stateSync       ShardStateSync
	stream          *string
	sequence        *string
	stop            <-chan Unit
	stopped         chan<- Unit
	c               chan *KinesisRecord
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
		if awsErr, ok := err.(awserr.Error); ok {
			s.logger.Crit("Failed to get shard iterator", "shard", *s.shard.ShardID,
				"sequence", *sequence, "error", awsErr)
		} else {
			panic(err)
		}
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

func (s *ShardWorker) GetRecordsAndProcess(it, sequence *string) (bool, *string) {
	records, nextIt, lag, err := s.GetRecords(it)
	if err != nil || len(records) == 0 {
		if err == nil {
			s.logger.Info("No new records. Waiting.", "shard", *s.shard.ShardID, "sequence", *sequence)
		} else {
			s.logger.Error("GetRecords encountered an error", "shard", *s.shard.ShardID, "error", err)
			nextIt = s.TryGetShardIterator("AFTER_SEQUENCE_NUMBER", sequence)
		}
		if lag < 30000 /* milliseconds */ {
			select {
			case <-time.NewTimer(30 * time.Second).C:
			case <-s.stop:
				return true, sequence
			}
		}
	} else {
		for _, rec := range records {
			s.c <- &KinesisRecord{
				Record:  rec,
				ShardID: s.shard.ShardID,
				sync:    s.stateSync.DoneC(),
			}
		}
		sequence = records[len(records)-1].SequenceNumber
	}
	it = nextIt
	return false, sequence
}

func (s *ShardWorker) RunWorker() {
	defer func() {
		s.stopped <- Unit{}
	}()

	sequence := s.stateSync.GetStartSequence(s.shard.ShardID)
	end := s.shard.SequenceNumberRange.EndingSequenceNumber
	var it *string
	if sequence == nil || len(*sequence) == 0 {
		sequence = s.shard.SequenceNumberRange.StartingSequenceNumber
		s.logger.Info("Using shard sequence head", "shard", *s.shard.ShardID, "sequence", *sequence)
		it = s.TryGetShardIterator("TRIM_HORIZON", nil)
	} else {
		it = s.TryGetShardIterator("AFTER_SEQUENCE_NUMBER", sequence)
	}

loop:
	for {
		if end != nil && *sequence == *end {
			s.logger.Info("Shard has reached its end", "shard", *s.shard.ShardID)
			break loop
		}

		select {
		case <-s.stop:
			break loop
		default:
			if brk, seq := s.GetRecordsAndProcess(it, sequence); brk {
				break loop
			} else {
				sequence = seq
			}
		}
	}

	s.logger.Info("Shard worker stopping", "shard", *s.shard.ShardID)
}
