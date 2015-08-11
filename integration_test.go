// +build integration

package kinesumer

import (
	"fmt"
	"testing"
)

type nSync struct {
	In  chan struct{}
	Out chan struct{}
}

func NewNSync(n int) nSync {
	sync := nSync{
		In:  make(chan struct{}),
		Out: make(chan struct{}),
	}
	go func() {
		for i := 0; i < n; i++ {
			<-sync.In
		}
		close(sync.Out)
	}()
	return sync
}

func (n nSync) Sync() {
	n.In <- struct{}{}
	<-n.Out
}

func TestEverything(t *testing.T) {
	fmt.Println("wat wat")
	fmt.Println("wat wat")
}
