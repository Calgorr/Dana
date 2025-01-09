package opentelemetry

import (
	"strings"

	"Dana"
)

type otelLogger struct {
	Dana.Logger
}

func (l otelLogger) Debug(msg string, kv ...interface{}) {
	format := msg + strings.Repeat(" %s=%q", len(kv)/2)
	l.Logger.Debugf(format, kv...)
}
