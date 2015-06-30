package kinesumer

import (
	"sync"
)

type EmptyStateSync struct {
	c  *chan *ShardState
	wg sync.WaitGroup
}

func NewEmptyStateSync() *EmptyStateSync {
	c := make(chan *ShardState)
	return &EmptyStateSync{
		c: &c,
	}
}

func (e *EmptyStateSync) doneC() *chan *ShardState {
	return e.c
}

func (e *EmptyStateSync) begin() {
	e.wg.Add(1)
	go func() {
		for range *e.c {
		}
		e.wg.Done()
	}()
}

func (e *EmptyStateSync) end() {
	close(*e.c)
}

func (e *EmptyStateSync) getStartSequence(*string) *string {
	return nil
}
