package kinesumeriface

type KinesumerHandlers interface {
	Go(f func())
	Err(err Error)
}

type DefaultKinesumerHandlers struct {
}

func (h DefaultKinesumerHandlers) Go(f func()) {
	go f()
}

func (h DefaultKinesumerHandlers) Err(err Error) {
	panic(err)
}
