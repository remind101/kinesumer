package kinesumeriface

const (
	KinesumerECrit  = "crit"
	KinesumerEError = "error"
	KinesumerEWarn  = "warn"
	KinesumerEInfo  = "info"
	KinesumerEDebug = "debug"
)

type KinesumerError struct {
	// One of "crit", "error", "warn", "info", "debug"
	Severity string
	message  string
	Origin   error
}

func NewKinesumerError(severity, message string, origin error) *KinesumerError {
	return &KinesumerError{
		Severity: severity,
		message:  message,
		Origin:   origin,
	}
}

func (e *KinesumerError) Error() string {
	if e.Origin == nil {
		return e.message
	} else {
		return e.message + " from " + e.Origin.Error()
	}
}
