package kinesumeriface

type Provisioner interface {
	Begin()
	End()
	TryAcquire(shardID *string) error
	Release(shardID *string) error
}
