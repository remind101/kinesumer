package emptycheckpointer

import (
	k "github.com/remind101/kinesumer/interface"
)

type Checkpointer struct {
}

func (p *Checkpointer) DoneC() chan<- *k.KinesisRecord {
	return nil
}

func (p *Checkpointer) Begin(k.KinesumerHandlers) error {
	return nil
}

func (p *Checkpointer) End() {
}

func (p *Checkpointer) GetStartSequence(*string) *string {
	return nil
}

func (p *Checkpointer) Sync() {
}

func (p *Checkpointer) TryAcquire(shardID *string) error {
	return nil
}

func (p *Checkpointer) Release(shardID *string) error {
	return nil
}
