package emptyprovisioner

type Provisioner struct {
}

func (l *Provisioner) Begin() {
}

func (l *Provisioner) End() {
}

func (l *Provisioner) TryAcquire(shardID *string) error {
	return nil
}

func (l *Provisioner) Release(shardID *string) error {
	return nil
}
