package kinesumer

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/remind101/pkg/logger"
)

type ShardState struct {
	Record  *kinesis.Record
	ShardID *string
	sync    ShardStateSync
}

func (s *ShardState) Done() {
	*s.sync.doneC() <- s
}

type ShardStateSync interface {
	doneC() *chan *ShardState
	begin()
	end()
	getStartSequence(shardID *string) *string
}

type ShardStateSyncOptions struct {
	Logger logger.Logger
	Ticker *<-chan time.Time
}
