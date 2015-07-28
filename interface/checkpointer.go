package kinesumeriface

type Checkpointer interface {
	DoneC() chan<- *KinesisRecord
	Begin(handlers KinesumerHandlers) error
	End()
	GetStartSequence(shardID *string) *string
	Sync()
}
