package shim

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"Dana"
	cfg "Dana/config"
	"Dana/plugins/inputs"
	"Dana/plugins/processors"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("SECRET_TOKEN", "xxxxxxxxxx")
	t.Setenv("SECRET_VALUE", `test"\test`)

	inputs.Add("test", func() Dana.Input {
		return &serviceInput{}
	})

	c := "./testdata/plugin.conf"
	conf, err := LoadConfig(&c)
	require.NoError(t, err)

	inp := conf.Input.(*serviceInput)

	require.Equal(t, "awesome name", inp.ServiceName)
	require.Equal(t, "xxxxxxxxxx", inp.SecretToken)
	require.Equal(t, `test"\test`, inp.SecretValue)
}

func TestLoadingSpecialTypes(t *testing.T) {
	inputs.Add("test", func() Dana.Input {
		return &testDurationInput{}
	})

	c := "./testdata/special.conf"
	conf, err := LoadConfig(&c)
	require.NoError(t, err)

	inp := conf.Input.(*testDurationInput)

	require.EqualValues(t, 3*time.Second, inp.Duration)
	require.EqualValues(t, 3*1000*1000, inp.Size)
	require.EqualValues(t, 52, inp.Hex)
}

func TestLoadingProcessorWithConfig(t *testing.T) {
	proc := &testConfigProcessor{}
	processors.Add("test_config_load", func() Dana.Processor {
		return proc
	})

	c := "./testdata/processor.conf"
	_, err := LoadConfig(&c)
	require.NoError(t, err)

	require.EqualValues(t, "yep", proc.Loaded)
}

type testDurationInput struct {
	Duration cfg.Duration `toml:"duration"`
	Size     cfg.Size     `toml:"size"`
	Hex      int64        `toml:"hex"`
}

func (i *testDurationInput) SampleConfig() string {
	return ""
}

func (i *testDurationInput) Description() string {
	return ""
}
func (i *testDurationInput) Gather(_ Dana.Accumulator) error {
	return nil
}

type testConfigProcessor struct {
	Loaded string `toml:"loaded"`
}

func (p *testConfigProcessor) SampleConfig() string {
	return ""
}

func (p *testConfigProcessor) Description() string {
	return ""
}
func (p *testConfigProcessor) Apply(metrics ...Dana.Metric) []Dana.Metric {
	return metrics
}
