package printer

import (
	"sync"
	"testing"
	"time"

	"Dana"
	"Dana/metric"
	"Dana/testutil"
	"github.com/stretchr/testify/require"
)

func TestTracking(t *testing.T) {
	// Setup raw input and expected output
	inputRaw := []Dana.Metric{
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"value": uint64(3)},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"value": int64(4)},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"value": float64(5.5)},
			time.Unix(0, 0),
		),
	}

	expected := []Dana.Metric{
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"value": uint64(3)},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"value": int64(4)},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"value": float64(5.5)},
			time.Unix(0, 0),
		),
	}

	// Create fake notification for testing
	var mu sync.Mutex
	delivered := make([]Dana.DeliveryInfo, 0, len(inputRaw))
	notify := func(di Dana.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	// Convert raw input to tracking metric
	input := make([]Dana.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	// Prepare and start the plugin
	plugin := &Printer{}
	require.NoError(t, plugin.Init())

	// Process expected metrics and compare with resulting metrics
	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}
