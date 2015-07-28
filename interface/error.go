package kinesumeriface

type Error interface {
	Severity() string
	Origin() error
	Error() string
}
