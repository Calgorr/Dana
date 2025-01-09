package pivot

import (
	"sync"
	"testing"
	"time"

	"Dana"
	"Dana/metric"
	"Dana/testutil"
	"github.com/stretchr/testify/require"
)

func TestPivot(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		pivot    *Pivot
		metrics  []Dana.Metric
		expected []Dana.Metric
	}{
		{
			name: "simple",
			pivot: &Pivot{
				TagKey:   "name",
				ValueKey: "value",
			},
			metrics: []Dana.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
			},
			expected: []Dana.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{
						"idle_time": int64(42),
					},
					now,
				),
			},
		},
		{
			name: "missing tag",
			pivot: &Pivot{
				TagKey:   "name",
				ValueKey: "value",
			},
			metrics: []Dana.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"foo": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
			},
			expected: []Dana.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"foo": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
			},
		},
		{
			name: "missing field",
			pivot: &Pivot{
				TagKey:   "name",
				ValueKey: "value",
			},
			metrics: []Dana.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"foo": int64(42),
					},
					now,
				),
			},
			expected: []Dana.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"foo": int64(42),
					},
					now,
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.pivot.Apply(tt.metrics...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func TestTracking(t *testing.T) {
	// Setup raw input and expected output
	inputRaw := []Dana.Metric{
		metric.New(
			"test",
			map[string]string{"name": "idle_time"},
			map[string]interface{}{"value": uint64(3)},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{"name": "system_time"},
			map[string]interface{}{"value": int64(4)},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{"name": "user_time"},
			map[string]interface{}{"value": float64(5.5)},
			time.Unix(0, 0),
		),
	}

	expected := []Dana.Metric{
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"idle_time": uint64(3)},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"system_time": int64(4)},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"user_time": float64(5.5)},
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
	plugin := &Pivot{
		TagKey:   "name",
		ValueKey: "value",
	}

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
