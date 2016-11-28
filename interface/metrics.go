package kinesumeriface

type MetricsReporter interface {
	Count(name string, value int64, tags map[string]string, rate float64) error
	Gauge(name string, value float64, tags map[string]string, rate float64) error
	Histogram(name string, value float64, tags map[string]string, rate float64) error
	Set(name string, value string, tags map[string]string, rate float64) error
	TimeInMilliseconds(name string, value float64, tags map[string]string, rate float64) error
}
