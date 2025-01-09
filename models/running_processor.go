package models

import (
	"sync"

	"Dana"
	logging "Dana/logger"
	"Dana/selfstat"
)

type RunningProcessor struct {
	sync.Mutex
	log       Dana.Logger
	Processor Dana.StreamingProcessor
	Config    *ProcessorConfig
}

type RunningProcessors []*RunningProcessor

func (rp RunningProcessors) Len() int           { return len(rp) }
func (rp RunningProcessors) Swap(i, j int)      { rp[i], rp[j] = rp[j], rp[i] }
func (rp RunningProcessors) Less(i, j int) bool { return rp[i].Config.Order < rp[j].Config.Order }

// ProcessorConfig containing a name and filter
type ProcessorConfig struct {
	Name     string
	Alias    string
	ID       string
	Order    int64
	Filter   Filter
	LogLevel string
}

func NewRunningProcessor(processor Dana.StreamingProcessor, config *ProcessorConfig) *RunningProcessor {
	tags := map[string]string{"processor": config.Name}
	if config.Alias != "" {
		tags["alias"] = config.Alias
	}

	processErrorsRegister := selfstat.Register("process", "errors", tags)
	logger := logging.New("processors", config.Name, config.Alias)
	logger.RegisterErrorCallback(func() {
		processErrorsRegister.Incr(1)
	})
	if err := logger.SetLogLevel(config.LogLevel); err != nil {
		logger.Error(err)
	}
	SetLoggerOnPlugin(processor, logger)

	return &RunningProcessor{
		Processor: processor,
		Config:    config,
		log:       logger,
	}
}

func (rp *RunningProcessor) metricFiltered(metric Dana.Metric) {
	metric.Drop()
}

func (rp *RunningProcessor) Init() error {
	if p, ok := rp.Processor.(Dana.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (rp *RunningProcessor) ID() string {
	if p, ok := rp.Processor.(Dana.PluginWithID); ok {
		return p.ID()
	}
	return rp.Config.ID
}

func (rp *RunningProcessor) Log() Dana.Logger {
	return rp.log
}

func (rp *RunningProcessor) LogName() string {
	return logName("processors", rp.Config.Name, rp.Config.Alias)
}

func (rp *RunningProcessor) MakeMetric(metric Dana.Metric) Dana.Metric {
	return metric
}

func (rp *RunningProcessor) Start(acc Dana.Accumulator) error {
	return rp.Processor.Start(acc)
}

func (rp *RunningProcessor) Add(m Dana.Metric, acc Dana.Accumulator) error {
	ok, err := rp.Config.Filter.Select(m)
	if err != nil {
		rp.log.Errorf("filtering failed: %v", err)
	} else if !ok {
		// pass downstream
		acc.AddMetric(m)
		return nil
	}

	rp.Config.Filter.Modify(m)
	if len(m.FieldList()) == 0 {
		// drop metric
		rp.metricFiltered(m)
		return nil
	}

	return rp.Processor.Add(m, acc)
}

func (rp *RunningProcessor) Stop() {
	rp.Processor.Stop()
}
