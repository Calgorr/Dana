package shim

import (
	"bufio"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"Dana"
	"Dana/logger"
)

func TestShimSetsUpLogger(t *testing.T) {
	stderrReader, stderrWriter := io.Pipe()
	stdinReader, stdinWriter := io.Pipe()

	runErroringInputPlugin(t, 40*time.Second, stdinReader, nil, stderrWriter)

	_, err := stdinWriter.Write([]byte("\n"))
	require.NoError(t, err)

	r := bufio.NewReader(stderrReader)
	out, err := r.ReadString('\n')
	require.NoError(t, err)
	require.Contains(t, out, "Error in plugin: intentional")

	err = stdinWriter.Close()
	require.NoError(t, err)
}

func runErroringInputPlugin(t *testing.T, interval time.Duration, stdin io.Reader, stdout, stderr io.Writer) (chan bool, chan bool) {
	metricProcessed := make(chan bool, 1)
	exited := make(chan bool, 1)
	inp := &erroringInput{}

	shim := New()
	if stdin != nil {
		shim.stdin = stdin
	}
	if stdout != nil {
		shim.stdout = stdout
	}
	if stderr != nil {
		shim.stderr = stderr
		logger.RedirectLogging(stderr)
	}
	err := shim.AddInput(inp)
	require.NoError(t, err)
	go func() {
		if err := shim.Run(interval); err != nil {
			t.Error(err)
		}
		exited <- true
	}()
	return metricProcessed, exited
}

type erroringInput struct {
}

func (i *erroringInput) SampleConfig() string {
	return ""
}

func (i *erroringInput) Gather(acc Dana.Accumulator) error {
	acc.AddError(errors.New("intentional"))
	return nil
}

func (i *erroringInput) Start(_ Dana.Accumulator) error {
	return nil
}

func (i *erroringInput) Stop() {
}
