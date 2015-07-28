package kinesumer

import (
	"fmt"

	k "github.com/remind101/kinesumer/interface"
)

type DefaultHandlers struct {
}

func (h DefaultHandlers) Go(f func()) {
	go f()
}

func (h DefaultHandlers) Err(err k.Error) {
	fmt.Println(err.Severity()+":", err.Error())

	severity := err.Severity()
	if severity == ECrit || severity == EError {
		panic(err)
	}
}
