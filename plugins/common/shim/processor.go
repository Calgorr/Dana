package shim

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"Dana"
	"Dana/agent"
	"Dana/models"
	"Dana/plugins/parsers/influx"
	"Dana/plugins/processors"
)

// AddProcessor adds the processor to the shim. Later calls to Run() will run this.
func (s *Shim) AddProcessor(processor Dana.Processor) error {
	models.SetLoggerOnPlugin(processor, s.Log())
	p := processors.NewStreamingProcessorFromProcessor(processor)
	return s.AddStreamingProcessor(p)
}

// AddStreamingProcessor adds the processor to the shim. Later calls to Run() will run this.
func (s *Shim) AddStreamingProcessor(processor Dana.StreamingProcessor) error {
	models.SetLoggerOnPlugin(processor, s.Log())
	if p, ok := processor.(Dana.Initializer); ok {
		err := p.Init()
		if err != nil {
			return fmt.Errorf("failed to init input: %w", err)
		}
	}

	s.Processor = processor
	return nil
}

func (s *Shim) RunProcessor() error {
	acc := agent.NewAccumulator(s, s.metricCh)
	acc.SetPrecision(time.Nanosecond)

	err := s.Processor.Start(acc)
	if err != nil {
		return fmt.Errorf("failed to start processor: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := s.writeProcessedMetrics()
		if err != nil {
			s.log.Warn(err.Error())
		}
		wg.Done()
	}()

	parser := influx.NewStreamParser(s.stdin)
	for {
		m, err := parser.Next()
		if err != nil {
			if errors.Is(err, influx.EOF) {
				break // stream ended
			}
			var parseErr *influx.ParseError
			if errors.As(err, &parseErr) {
				fmt.Fprintf(s.stderr, "Failed to parse metric: %s\b", parseErr)
				continue
			}
			fmt.Fprintf(s.stderr, "Failure during reading stdin: %s\b", err)
			continue
		}

		if err = s.Processor.Add(m, acc); err != nil {
			fmt.Fprintf(s.stderr, "Failure during processing metric by processor: %v\b", err)
		}
	}

	close(s.metricCh)
	s.Processor.Stop()
	wg.Wait()
	return nil
}
