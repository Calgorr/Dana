package shim

// this package is deprecated. use plugins/common/shim instead

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"

	"Dana"
	"Dana/agent"
	"Dana/plugins/inputs"
	"Dana/plugins/serializers/influx"
)

var (
	envVarEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
	oldpkg = "Dana/plugins/inputs/execd/shim"
	newpkg = "Dana/plugins/common/shim"
)

const (
	// PollIntervalDisabled is used to indicate that you want to disable polling,
	// as opposed to duration 0 meaning poll constantly.
	PollIntervalDisabled = time.Duration(0)
)

// Shim allows you to wrap your inputs and run them as if they were part of Dana2,
// except built externally.
type Shim struct {
	Inputs            []Dana.Input
	gatherPromptChans []chan empty
	metricCh          chan Dana.Metric

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

type empty struct{}

// New creates a new shim interface
func New() *Shim {
	fmt.Fprintf(os.Stderr, "%s is deprecated; please change your import to %s\n", oldpkg, newpkg)
	return &Shim{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// AddInput adds the input to the shim. Later calls to Run() will run this input.
func (s *Shim) AddInput(input Dana.Input) error {
	if p, ok := input.(Dana.Initializer); ok {
		err := p.Init()
		if err != nil {
			return fmt.Errorf("failed to init input: %w", err)
		}
	}

	s.Inputs = append(s.Inputs, input)
	return nil
}

// AddInputs adds multiple inputs to the shim. Later calls to Run() will run these.
func (s *Shim) AddInputs(newInputs []Dana.Input) error {
	for _, inp := range newInputs {
		if err := s.AddInput(inp); err != nil {
			return err
		}
	}
	return nil
}

// Run the input plugins..
func (s *Shim) Run(pollInterval time.Duration) error {
	// context is used only to close the stdin reader. everything else cascades
	// from that point and closes cleanly when it's done.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.metricCh = make(chan Dana.Metric, 1)

	wg := sync.WaitGroup{}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	collectMetricsPrompt := make(chan os.Signal, 1)
	listenForCollectMetricsSignals(ctx, collectMetricsPrompt)

	serializer := &influx.Serializer{}
	if err := serializer.Init(); err != nil {
		return fmt.Errorf("creating serializer failed: %w", err)
	}

	for _, input := range s.Inputs {
		wrappedInput := inputShim{Input: input}

		acc := agent.NewAccumulator(wrappedInput, s.metricCh)
		acc.SetPrecision(time.Nanosecond)

		if serviceInput, ok := input.(Dana.ServiceInput); ok {
			if err := serviceInput.Start(acc); err != nil {
				return fmt.Errorf("failed to start input: %w", err)
			}
		}
		gatherPromptCh := make(chan empty, 1)
		s.gatherPromptChans = append(s.gatherPromptChans, gatherPromptCh)
		wg.Add(1) // one per input
		go func(input Dana.Input) {
			s.startGathering(ctx, input, acc, gatherPromptCh, pollInterval)
			if serviceInput, ok := input.(Dana.ServiceInput); ok {
				serviceInput.Stop()
			}
			close(gatherPromptCh)
			wg.Done()
		}(input)
	}

	go s.stdinCollectMetricsPrompt(ctx, cancel, collectMetricsPrompt)
	go s.closeMetricChannelWhenInputsFinish(&wg)

loop:
	for {
		select {
		case <-quit: // user-triggered quit
			// cancel, but keep looping until the metric channel closes.
			cancel()
		case _, open := <-collectMetricsPrompt:
			if !open { // stdin-close-triggered quit
				cancel()
				continue
			}
			s.collectMetrics(ctx)
		case m, open := <-s.metricCh:
			if !open {
				break loop
			}
			b, err := serializer.Serialize(m)
			if err != nil {
				return fmt.Errorf("failed to serialize metric: %w", err)
			}
			// Write this to stdout
			if _, err := fmt.Fprint(s.stdout, string(b)); err != nil {
				return fmt.Errorf("failed to write %q to stdout: %w", string(b), err)
			}
		}
	}

	return nil
}

// LoadConfig loads and adds the inputs to the shim
func (s *Shim) LoadConfig(filePath *string) error {
	loadedInputs, err := LoadConfig(filePath)
	if err != nil {
		return err
	}
	return s.AddInputs(loadedInputs)
}

// DefaultImportedPlugins defaults to whatever plugins happen to be loaded and
// have registered themselves with the registry. This makes loading plugins
// without having to define a config dead easy.
func DefaultImportedPlugins() (i []Dana.Input, e error) {
	for _, inputCreatorFunc := range inputs.Inputs {
		i = append(i, inputCreatorFunc())
	}
	return i, nil
}

// LoadConfig loads the config and returns inputs that later need to be loaded.
func LoadConfig(filePath *string) ([]Dana.Input, error) {
	if filePath == nil || *filePath == "" {
		return DefaultImportedPlugins()
	}

	b, err := os.ReadFile(*filePath)
	if err != nil {
		return nil, err
	}

	s := expandEnvVars(b)

	conf := struct {
		Inputs map[string][]toml.Primitive
	}{}

	md, err := toml.Decode(s, &conf)
	if err != nil {
		return nil, err
	}

	return loadConfigIntoInputs(md, conf.Inputs)
}

func hasQuit(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (s *Shim) stdinCollectMetricsPrompt(ctx context.Context, cancel context.CancelFunc, collectMetricsPrompt chan<- os.Signal) {
	defer func() {
		cancel()
		close(collectMetricsPrompt)
	}()

	scanner := bufio.NewScanner(s.stdin)
	// for every line read from stdin, make sure we're not supposed to quit,
	// then push a message on to the collectMetricsPrompt
	for scanner.Scan() {
		// first check if we should quit
		if hasQuit(ctx) {
			return
		}

		// now push a non-blocking message to trigger metric collection.
		pushCollectMetricsRequest(collectMetricsPrompt)
	}
}

// pushCollectMetricsRequest pushes a non-blocking (nil) message to the
// collectMetricsPrompt channel to trigger metric collection.
// The channel is defined with a buffer of 1, so while it's full, subsequent
// requests are discarded.
func pushCollectMetricsRequest(collectMetricsPrompt chan<- os.Signal) {
	select {
	case collectMetricsPrompt <- nil:
	default:
	}
}

func (s *Shim) collectMetrics(ctx context.Context) {
	if hasQuit(ctx) {
		return
	}
	for i := 0; i < len(s.gatherPromptChans); i++ {
		// push a message out to each channel to collect metrics. don't block.
		select {
		case s.gatherPromptChans[i] <- empty{}:
		default:
		}
	}
}

func (s *Shim) startGathering(ctx context.Context, input Dana.Input, acc Dana.Accumulator, gatherPromptCh <-chan empty, pollInterval time.Duration) {
	if pollInterval == PollIntervalDisabled {
		return // don't poll
	}
	t := time.NewTicker(pollInterval)
	defer t.Stop()
	for {
		// give priority to stopping.
		if hasQuit(ctx) {
			return
		}
		// see what's up
		select {
		case <-ctx.Done():
			return
		case <-gatherPromptCh:
			if err := input.Gather(acc); err != nil {
				if _, perr := fmt.Fprintf(s.stderr, "failed to gather metrics: %s", err); perr != nil {
					acc.AddError(err)
					acc.AddError(perr)
				}
			}
		case <-t.C:
			if err := input.Gather(acc); err != nil {
				if _, perr := fmt.Fprintf(s.stderr, "failed to gather metrics: %s", err); perr != nil {
					acc.AddError(err)
					acc.AddError(perr)
				}
			}
		}
	}
}

func expandEnvVars(contents []byte) string {
	return os.Expand(string(contents), getEnv)
}

func getEnv(key string) string {
	v := os.Getenv(key)

	return envVarEscaper.Replace(v)
}

func loadConfigIntoInputs(md toml.MetaData, inputConfigs map[string][]toml.Primitive) ([]Dana.Input, error) {
	renderedInputs := make([]Dana.Input, 0, len(inputConfigs))
	for name, primitives := range inputConfigs {
		inputCreator, ok := inputs.Inputs[name]
		if !ok {
			return nil, errors.New("unknown input " + name)
		}

		for _, primitive := range primitives {
			inp := inputCreator()
			// Parse specific configuration
			if err := md.PrimitiveDecode(primitive, inp); err != nil {
				return nil, err
			}

			renderedInputs = append(renderedInputs, inp)
		}
	}
	return renderedInputs, nil
}

func (s *Shim) closeMetricChannelWhenInputsFinish(wg *sync.WaitGroup) {
	wg.Wait()
	close(s.metricCh)
}
