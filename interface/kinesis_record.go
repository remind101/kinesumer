package kinesumeriface

type Record interface {
	Data() []byte
	PartitionKey() string
	SequenceNumber() string
	ShardID() string
	MillisBehindLatest() int64
	Done()
}
