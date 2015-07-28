package kinesumeriface

type Provisioner interface {
	Begin(handlers KinesumerHandlers) error
	End()
	TryAcquire(shardID *string) error
	Release(shardID *string) error
	HeartbeatC() chan string
}
