package kinesumeriface

import (
	"github.com/aws/aws-sdk-go/service/kinesis"
)

// If Err == nil then everything else is set
// Otherwise, Kinesis encountered a problem
type KinesisRecord struct {
	// This contains:
	//   Data []byte
	//   PartitionKey *string
	//   SequenceNumber *string
	kinesis.Record
	ShardID            *string
	CheckpointC        chan<- *KinesisRecord
	MillisBehindLatest int64
	// May or may not be a KinesumerError
	Err error
}

func (s *KinesisRecord) Done() {
	if s.Err == nil {
		s.CheckpointC <- s
	}
}
