package emptyprovisioner

import (
	k "github.com/remind101/kinesumer/interface"
)

type Provisioner struct {
}

func (l *Provisioner) Begin(k.KinesumerHandlers) error {
	return nil
}

func (l *Provisioner) End() {
}

func (l *Provisioner) TryAcquire(shardID *string) error {
	return nil
}

func (l *Provisioner) Release(shardID *string) error {
	return nil
}
