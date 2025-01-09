package mqtt

import (
	"Dana"
)

type mqttLogger struct {
	Dana.Logger
}

func (l mqttLogger) Printf(fmt string, args ...interface{}) {
	l.Logger.Debugf(fmt, args...)
}

func (l mqttLogger) Println(args ...interface{}) {
	l.Logger.Debug(args...)
}
