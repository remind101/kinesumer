package emptycheckpointer

import (
	k "github.com/remind101/kinesumer/interface"
)

type Checkpointer struct {
}

func (e *Checkpointer) DoneC() chan<- *k.KinesisRecord {
	return nil
}

func (e *Checkpointer) Begin(chan<- *k.KinesisRecord) error {
	return nil
}

func (e *Checkpointer) End() {
}

func (e *Checkpointer) GetStartSequence(*string) *string {
	return nil
}

func (e *Checkpointer) Sync() {
}

func (e *Checkpointer) TryAcquire(shardID *string) error {
	return nil
}

func (e *Checkpointer) Release(shardID *string) error {
	return nil
}
