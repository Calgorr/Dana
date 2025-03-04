package influx

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"Dana"
	"Dana/config"
	"Dana/plugins/parsers"
)

const (
	maxErrorBufferSize = 1024
)

var (
	ErrNoMetric = errors.New("no metric in line")
)

type TimeFunc func() time.Time

// ParseError indicates a error in the parsing of the text.
type ParseError struct {
	Offset     int
	LineOffset int
	LineNumber int
	Column     int
	msg        string
	buf        string
}

func (e *ParseError) Error() string {
	buffer := e.buf[e.LineOffset:]
	eol := strings.IndexAny(buffer, "\n")
	if eol >= 0 {
		buffer = strings.TrimSuffix(buffer[:eol], "\r")
	}
	if len(buffer) > maxErrorBufferSize {
		startEllipsis := true
		offset := e.Offset - e.LineOffset
		start := offset - maxErrorBufferSize
		if start < 0 {
			startEllipsis = false
			start = 0
		}
		// if we trimmed it the column won't line up. it'll always be the last character,
		// because the parser doesn't continue past it, but point it out anyway so
		// it's obvious where the issue is.
		buffer = buffer[start:offset] + "<-- here"
		if startEllipsis {
			buffer = "..." + buffer
		}
	}
	return fmt.Sprintf("metric parse error: %s at %d:%d: %q", e.msg, e.LineNumber, e.Column, buffer)
}

// Parser is an InfluxDB Line Protocol parser that implements the
// parsers.Parser interface.
type Parser struct {
	InfluxTimestampPrecision config.Duration   `toml:"influx_timestamp_precision"`
	DefaultTags              map[string]string `toml:"-"`
	// If set to "series" a series machine will be initialized, defaults to regular machine
	Type string `toml:"-"`

	sync.Mutex
	*machine
	handler *MetricHandler
}

func (p *Parser) SetTimeFunc(f TimeFunc) {
	p.handler.SetTimeFunc(f)
}

func (p *Parser) SetTimePrecision(u time.Duration) {
	p.handler.SetTimePrecision(u)
}

func (p *Parser) Parse(input []byte) ([]Dana.Metric, error) {
	p.Lock()
	defer p.Unlock()
	metrics := make([]Dana.Metric, 0)
	p.machine.SetData(input)

	for {
		err := p.machine.Next()
		if errors.Is(err, EOF) {
			break
		}

		if err != nil {
			return nil, &ParseError{
				Offset:     p.machine.Position(),
				LineOffset: p.machine.LineOffset(),
				LineNumber: p.machine.LineNumber(),
				Column:     p.machine.Column(),
				msg:        err.Error(),
				buf:        string(input),
			}
		}

		metric := p.handler.Metric()
		if metric == nil {
			continue
		}

		metrics = append(metrics, metric)
	}

	p.applyDefaultTags(metrics)
	return metrics, nil
}

func (p *Parser) ParseLine(line string) (Dana.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, ErrNoMetric
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) applyDefaultTags(metrics []Dana.Metric) {
	if len(p.DefaultTags) == 0 {
		return
	}

	for _, m := range metrics {
		p.applyDefaultTagsSingle(m)
	}
}

func (p *Parser) applyDefaultTagsSingle(metric Dana.Metric) {
	for k, v := range p.DefaultTags {
		if !metric.HasTag(k) {
			metric.AddTag(k, v)
		}
	}
}

func (p *Parser) Init() error {
	p.handler = NewMetricHandler()
	if p.Type == "series" {
		p.machine = NewSeriesMachine(p.handler)
	} else {
		p.machine = NewMachine(p.handler)
	}

	timeDuration := time.Duration(p.InfluxTimestampPrecision)
	switch timeDuration {
	case 0:
	case time.Nanosecond, time.Microsecond, time.Millisecond, time.Second:
		p.SetTimePrecision(timeDuration)
	default:
		return fmt.Errorf("invalid time precision: %d", p.InfluxTimestampPrecision)
	}

	return nil
}

func init() {
	parsers.Add("influx",
		func(string) Dana.Parser {
			return &Parser{}
		},
	)
}

// StreamParser is an InfluxDB Line Protocol parser.  It is not safe for
// concurrent use in multiple goroutines.
type StreamParser struct {
	machine *streamMachine
	handler *MetricHandler
}

func NewStreamParser(r io.Reader) *StreamParser {
	handler := NewMetricHandler()
	return &StreamParser{
		machine: NewStreamMachine(r, handler),
		handler: handler,
	}
}

// SetTimeFunc changes the function used to determine the time of metrics
// without a timestamp.  The default TimeFunc is time.Now.  Useful mostly for
// testing, or perhaps if you want all metrics to have the same timestamp.
func (sp *StreamParser) SetTimeFunc(f TimeFunc) {
	sp.handler.SetTimeFunc(f)
}

func (sp *StreamParser) SetTimePrecision(u time.Duration) {
	sp.handler.SetTimePrecision(u)
}

// Next parses the next item from the stream.  You can repeat calls to this
// function if it returns ParseError to get the next metric or error.
func (sp *StreamParser) Next() (Dana.Metric, error) {
	err := sp.machine.Next()
	if errors.Is(err, EOF) {
		return nil, err
	}

	var e *readErr
	if errors.As(err, &e) {
		return nil, e.Err
	}

	if err != nil {
		return nil, &ParseError{
			Offset:     sp.machine.Position(),
			LineOffset: sp.machine.LineOffset(),
			LineNumber: sp.machine.LineNumber(),
			Column:     sp.machine.Column(),
			msg:        err.Error(),
			buf:        sp.machine.LineText(),
		}
	}

	return sp.handler.Metric(), nil
}

// Position returns the current byte offset into the data.
func (sp *StreamParser) Position() int {
	return sp.machine.Position()
}

// LineOffset returns the byte offset of the current line.
func (sp *StreamParser) LineOffset() int {
	return sp.machine.LineOffset()
}

// LineNumber returns the current line number.  Lines are counted based on the
// regular expression `\r?\n`.
func (sp *StreamParser) LineNumber() int {
	return sp.machine.LineNumber()
}

// Column returns the current column.
func (sp *StreamParser) Column() int {
	return sp.machine.Column()
}

// LineText returns the text of the current line that has been parsed so far.
func (sp *StreamParser) LineText() string {
	return sp.machine.LineText()
}
