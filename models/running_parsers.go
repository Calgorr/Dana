package models

import (
	"time"

	"Dana"
	logging "Dana/logger"
	"Dana/selfstat"
)

type RunningParser struct {
	Parser Dana.Parser
	Config *ParserConfig
	log    Dana.Logger

	MetricsParsed selfstat.Stat
	ParseTime     selfstat.Stat
}

func NewRunningParser(parser Dana.Parser, config *ParserConfig) *RunningParser {
	tags := map[string]string{"type": config.DataFormat}
	if config.Alias != "" {
		tags["alias"] = config.Alias
	}

	parserErrorsRegister := selfstat.Register("parser", "errors", tags)
	logger := logging.New("parsers", config.DataFormat+"::"+config.Parent, config.Alias)
	logger.RegisterErrorCallback(func() {
		parserErrorsRegister.Incr(1)
	})
	if err := logger.SetLogLevel(config.LogLevel); err != nil {
		logger.Error(err)
	}
	SetLoggerOnPlugin(parser, logger)

	return &RunningParser{
		Parser: parser,
		Config: config,
		MetricsParsed: selfstat.Register(
			"parser",
			"metrics_parsed",
			tags,
		),
		ParseTime: selfstat.Register(
			"parser",
			"parse_time_ns",
			tags,
		),
		log: logger,
	}
}

// ParserConfig is the common config for all parsers.
type ParserConfig struct {
	Parent      string
	Alias       string
	DataFormat  string
	DefaultTags map[string]string
	LogLevel    string
}

func (r *RunningParser) LogName() string {
	return logName("parsers", r.Config.DataFormat+"::"+r.Config.Parent, r.Config.Alias)
}

func (r *RunningParser) Init() error {
	if p, ok := r.Parser.(Dana.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RunningParser) Parse(buf []byte) ([]Dana.Metric, error) {
	start := time.Now()
	m, err := r.Parser.Parse(buf)
	elapsed := time.Since(start)
	r.ParseTime.Incr(elapsed.Nanoseconds())
	r.MetricsParsed.Incr(int64(len(m)))

	return m, err
}

func (r *RunningParser) ParseLine(line string) (Dana.Metric, error) {
	start := time.Now()
	m, err := r.Parser.ParseLine(line)
	elapsed := time.Since(start)
	r.ParseTime.Incr(elapsed.Nanoseconds())
	r.MetricsParsed.Incr(1)

	return m, err
}

func (r *RunningParser) SetDefaultTags(tags map[string]string) {
	r.Parser.SetDefaultTags(tags)
}

func (r *RunningParser) Log() Dana.Logger {
	return r.log
}
