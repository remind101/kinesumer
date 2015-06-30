package kinesumer

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/remind101/pkg/logger"
)

type ShardWorker struct {
	kinesis         *kinesis.Kinesis
	logger          logger.Logger
	shard           *kinesis.Shard
	stateSync       ShardStateSync
	stream          *string
	sequence        *string
	stop            *chan struct{}
	stopped         *chan struct{}
	c               *chan *ShardState
	GetRecordsLimit int64
}

func (s *ShardWorker) GetShardIterator() *string {
	shardIteratorType := "AFTER_SEQUENCE_NUMBER"
	iter, err := s.kinesis.GetShardIterator(&kinesis.GetShardIteratorInput{
		ShardID:                s.shard.ShardID,
		ShardIteratorType:      &shardIteratorType,
		StartingSequenceNumber: s.sequence,
		StreamName:             s.stream,
	})
	if err != nil {
		panic(err)
	}
	return iter.ShardIterator
}

func (s *ShardWorker) GetRecords(it *string) ([]*kinesis.Record, *string) {
	resp, err := s.kinesis.GetRecords(&kinesis.GetRecordsInput{
		Limit:         &s.GetRecordsLimit,
		ShardIterator: it,
	})
	if err != nil {
		panic(err)
	}
	return resp.Records, resp.NextShardIterator
}

func (s *ShardWorker) RunWorker() {
	defer func() {
		(*s.stopped) <- struct{}{}
	}()

	s.sequence = s.stateSync.getStartSequence(s.shard.ShardID)
	end := s.shard.SequenceNumberRange.EndingSequenceNumber
	if s.sequence == nil || len(*s.sequence) == 0 {
		s.sequence = s.shard.SequenceNumberRange.StartingSequenceNumber
		s.logger.Info("Using shard sequence head", "shard", *s.shard.ShardID, "sequence", *s.sequence)
	}

	it := s.GetShardIterator()

loop:
	for {
		if end != nil && *s.sequence == *end {
			s.logger.Info("Shard has reached its end", "shard", *s.shard.ShardID)
			break loop
		}

		select {
		case <-*s.stop:
			break loop
		default:
			records, nextIt := s.GetRecords(it)
			if len(records) == 0 {
				s.logger.Info("No new records. Waiting.", "shard", *s.shard.ShardID, "sequence", *s.sequence)
				select {
				case <-time.NewTimer(30 * time.Second).C:
					break
				case <-*s.stop:
					break loop
				}
			} else {
				for _, rec := range records {
					*s.c <- &ShardState{
						Record:  rec,
						ShardID: s.shard.ShardID,
						sync:    s.stateSync,
					}
				}
				s.sequence = records[len(records)-1].SequenceNumber
			}
			it = nextIt
		}
	}

	s.logger.Info("Shard worker stopping", "shard", *s.shard.ShardID)
}
