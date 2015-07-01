package kinesumer

import (
	"sync"
)

type EmptyStateSync struct {
	c  chan *KinesisRecord
	wg sync.WaitGroup
}

func NewEmptyStateSync() *EmptyStateSync {
	return &EmptyStateSync{
		c: make(chan *KinesisRecord),
	}
}

func (e *EmptyStateSync) DoneC() chan *KinesisRecord {
	return e.c
}

func (e *EmptyStateSync) Begin() error {
	e.wg.Add(1)
	go func() {
		for range e.c {
		}
		e.wg.Done()
	}()
	return nil
}

func (e *EmptyStateSync) End() {
	close(e.c)
}

func (e *EmptyStateSync) GetStartSequence(*string) *string {
	return nil
}

func (e *EmptyStateSync) Sync() {
}
