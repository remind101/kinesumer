package kinesumeriface

type Handlers interface {
	Go(f func())
	Err(err Error)
}
