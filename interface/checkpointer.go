package kinesumeriface

type Checkpointer interface {
	DoneC() chan<- *KinesisRecord
	Begin(chan<- *KinesisRecord) error
	End()
	GetStartSequence(shardID *string) *string
	Sync()
	TryAcquire(shardID *string) error
	Release(shardID *string) error
}
