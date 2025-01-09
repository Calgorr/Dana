package opcua

import (
	"Dana"
)

// DebugLogger logs messages from opcua at the debug level.
type DebugLogger struct {
	Log Dana.Logger
}

func (l *DebugLogger) Write(p []byte) (n int, err error) {
	l.Log.Debug(string(p))
	return len(p), nil
}
