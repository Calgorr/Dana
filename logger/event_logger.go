//go:build windows

package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"Dana"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	eidInfo    = 1
	eidWarning = 2
	eidError   = 3
)

type eventLogger struct {
	eventlog *eventlog.Log
	errlog   *log.Logger
}

func (l *eventLogger) Close() error {
	return l.eventlog.Close()
}

func (l *eventLogger) Print(level Dana.LogLevel, _ time.Time, prefix string, _ map[string]interface{}, args ...interface{}) {
	// Skip debug and beyond as they cannot be logged
	if level >= Dana.Debug {
		return
	}

	msg := level.Indicator() + " " + prefix + fmt.Sprint(args...)

	var err error
	switch level {
	case Dana.Error:
		err = l.eventlog.Error(eidError, msg)
	case Dana.Warn:
		err = l.eventlog.Warning(eidWarning, msg)
	case Dana.Info:
		err = l.eventlog.Info(eidInfo, msg)
	}
	if err != nil {
		l.errlog.Printf("E! Writing log message failed: %v", err)
	}

	// TODO attributes...
}

func createEventLogger(cfg *Config) (sink, error) {
	eventLog, err := eventlog.Open(cfg.InstanceName)
	if err != nil {
		return nil, err
	}

	l := &eventLogger{
		eventlog: eventLog,
		errlog:   log.New(os.Stderr, "", 0),
	}

	return l, nil
}

func init() {
	add("eventlog", createEventLogger)
}
