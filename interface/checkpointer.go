package kinesumeriface

type Checkpointer interface {
	DoneC() chan<- Record
	Begin(handlers Handlers) error
	End()
	GetStartSequence(shardID *string) *string
	Sync()
}
