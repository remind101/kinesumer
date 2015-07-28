package kinesumeriface

type KinesumerHandlers interface {
	Go(f func())
	Err(err *KinesumerError)
}

type DefaultKinesumerHandlers struct {
}

func (h DefaultKinesumerHandlers) Go(f func()) {
	go f()
}

func (h DefaultKinesumerHandlers) Err(err *KinesumerError) {
	panic(err)
}
