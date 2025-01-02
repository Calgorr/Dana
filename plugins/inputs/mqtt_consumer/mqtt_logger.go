package mqtt_consumer

import (
	"Dana"
)

type mqttLogger struct {
	telegraf.Logger
}

// Printf implements mqtt.Logger
func (l mqttLogger) Printf(fmt string, args ...interface{}) {
	l.Logger.Debugf(fmt, args...)
}

// Println implements mqtt.Logger
func (l mqttLogger) Println(args ...interface{}) {
	l.Logger.Debug(args...)
}
