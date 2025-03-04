package kafka

import (
	"sync"

	"github.com/IBM/sarama"

	"Dana"
	"Dana/logger"
)

var (
	log  = logger.New("sarama", "", "")
	once sync.Once
)

type debugLogger struct{}

func (l *debugLogger) Print(v ...interface{}) {
	log.Trace(v...)
}

func (l *debugLogger) Printf(format string, v ...interface{}) {
	log.Tracef(format, v...)
}

func (l *debugLogger) Println(v ...interface{}) {
	l.Print(v...)
}

// SetLogger configures a debug logger for kafka (sarama)
func SetLogger(level Dana.LogLevel) {
	// Set-up the sarama logger only once
	once.Do(func() {
		sarama.Logger = &debugLogger{}
	})
	// Increase the log-level if needed.
	if !log.Level().Includes(level) {
		log.SetLevel(level)
	}
}
