package kinesumeriface

type Handlers interface {
	Go(f func())
	Err(err Error)
}

type DefaultHandlers struct {
}

func (h DefaultHandlers) Go(f func()) {
	go f()
}

func (h DefaultHandlers) Err(err Error) {
	panic(err)
}
