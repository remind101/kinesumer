package kinesumer

import (
	k "github.com/remind101/kinesumer/interface"
)

type testHandlers struct{}

var errs = make([]k.Error, 0)
var toRun = make([]func(), 0)

func resetTestHandlers() {
	errs = make([]k.Error, 0)
	toRun = make([]func(), 0)
}

func (t testHandlers) Go(f func()) {
	toRun = append(toRun, f)
}

func (t testHandlers) Err(e k.Error) {
	errs = append(errs, e)
}
