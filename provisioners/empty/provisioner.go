package emptyprovisioner

import (
	k "github.com/remind101/kinesumer/interface"
)

type Provisioner struct {
}

func (p *Provisioner) Begin(k.KinesumerHandlers) error {
	return nil
}

func (p *Provisioner) End() {
}

func (p *Provisioner) TryAcquire(shardID *string) error {
	return nil
}

func (p *Provisioner) Release(shardID *string) error {
	return nil
}

func (p *Provisioner) HeartbeatC() chan string {
	return nil
}
