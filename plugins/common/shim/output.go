package shim

import (
	"bufio"
	"fmt"

	"Dana"
	"Dana/models"
	"Dana/plugins/parsers/influx"
)

// AddOutput adds the input to the shim. Later calls to Run() will run this.
func (s *Shim) AddOutput(output Dana.Output) error {
	models.SetLoggerOnPlugin(output, s.Log())
	if p, ok := output.(Dana.Initializer); ok {
		err := p.Init()
		if err != nil {
			return fmt.Errorf("failed to init input: %w", err)
		}
	}

	s.Output = output
	return nil
}

func (s *Shim) RunOutput() error {
	parser := influx.Parser{}
	err := parser.Init()
	if err != nil {
		return fmt.Errorf("failed to create new parser: %w", err)
	}

	err = s.Output.Connect()
	if err != nil {
		return fmt.Errorf("failed to start processor: %w", err)
	}
	defer s.Output.Close()

	var m Dana.Metric

	scanner := bufio.NewScanner(s.stdin)
	for scanner.Scan() {
		m, err = parser.ParseLine(scanner.Text())
		if err != nil {
			fmt.Fprintf(s.stderr, "Failed to parse metric: %s\n", err)
			continue
		}
		if err = s.Output.Write([]Dana.Metric{m}); err != nil {
			fmt.Fprintf(s.stderr, "Failed to write metric: %s\n", err)
		}
	}

	return nil
}
