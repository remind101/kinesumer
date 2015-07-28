package kinesumeriface

type Lifechecker interface {
	Begin()
	End()
	TryAcquire(shardID *string) error
	Release(shardID *string) error
}
