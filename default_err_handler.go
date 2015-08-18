package kinesumer

import (
	"fmt"
	"os"

	k "github.com/remind101/kinesumer/interface"
)

// This is the default error handler. It prints the error to stderr and panics if its level is worse
// than or equal to EError.
func DefaultErrHandler(err k.Error) {
	fmt.Fprintf(os.Stderr, err.Severity()+":", err.Error())

	severity := err.Severity()
	if severity == ECrit || severity == EError {
		panic(err)
	}
}
