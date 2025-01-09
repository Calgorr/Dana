package kinesis_consumer

import (
	"github.com/aws/smithy-go/logging"

	"Dana"
)

type Dana2LoggerWrapper struct {
	Dana.Logger
}

func (t *Dana2LoggerWrapper) Log(args ...interface{}) {
	t.Trace(args...)
}

func (t *Dana2LoggerWrapper) Logf(classification logging.Classification, format string, v ...interface{}) {
	switch classification {
	case logging.Debug:
		format = "DEBUG " + format
	case logging.Warn:
		format = "WARN" + format
	default:
		format = "INFO " + format
	}
	t.Logger.Tracef(format, v...)
}
