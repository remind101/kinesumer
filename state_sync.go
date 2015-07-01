package kinesumer

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/remind101/pkg/logger"
)

type KinesisRecord struct {
	Record  *kinesis.Record
	ShardID *string
	sync    chan *KinesisRecord
}

func (s *KinesisRecord) Done() {
	s.sync <- s
}

type ShardStateSync interface {
	DoneC() chan *KinesisRecord
	Begin() error
	End()
	GetStartSequence(shardID *string) *string
	Sync()
}

type ShardStateSyncOptions struct {
	Logger logger.Logger
	Ticker <-chan time.Time
}
