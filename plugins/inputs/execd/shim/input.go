package shim

import "Dana"

// inputShim implements the MetricMaker interface.
type inputShim struct {
	Input Dana.Input
}

// LogName satisfies the MetricMaker interface
func (inputShim) LogName() string {
	return ""
}

// MakeMetric satisfies the MetricMaker interface
func (inputShim) MakeMetric(m Dana.Metric) Dana.Metric {
	return m // don't need to do anything to it.
}

// Log satisfies the MetricMaker interface
func (inputShim) Log() Dana.Logger {
	return nil
}
