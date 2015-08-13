package kinesumer

import (
	k "github.com/remind101/kinesumer/interface"
)

type Record struct {
	data               []byte
	partitionKey       string
	sequenceNumber     string
	shardID            string
	millisBehindLatest int64
	checkpointC        chan<- k.Record
}

func (r *Record) Data() []byte {
	return r.data
}

func (r *Record) PartitionKey() string {
	return r.partitionKey
}

func (r *Record) SequenceNumber() string {
	return r.sequenceNumber
}

func (r *Record) ShardID() string {
	return r.shardID
}

func (r *Record) MillisBehindLatest() int64 {
	return r.millisBehindLatest
}

func (r *Record) Done() {
	r.checkpointC <- r
}
