package kinesumeriface

type Checkpointer interface {
	DoneC() chan<- Record
	Begin(handlers KinesumerHandlers) error
	End()
	GetStartSequence(shardID *string) *string
	Sync()
}
