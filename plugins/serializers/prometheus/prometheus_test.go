package prometheus

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"Dana"
	"Dana/plugins/serializers"
	"Dana/testutil"
)

func TestSerialize(t *testing.T) {
	tests := []struct {
		name     string
		config   FormatConfig
		metric   Dana.Metric
		expected []byte
	}{
		{
			name: "simple",
			metric: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
				},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "prometheus input untyped",
			metric: testutil.MustMetric(
				"prometheus",
				map[string]string{
					"code":   "400",
					"method": "post",
				},
				map[string]interface{}{
					"http_requests_total": 3.0,
				},
				time.Unix(0, 0),
				Dana.Untyped,
			),
			expected: []byte(`
# HELP http_requests_total Dana2 collected metric
# TYPE http_requests_total untyped
http_requests_total{code="400",method="post"} 3
`),
		},
		{
			name: "prometheus input counter",
			metric: testutil.MustMetric(
				"prometheus",
				map[string]string{
					"code":   "400",
					"method": "post",
				},
				map[string]interface{}{
					"http_requests_total": 3.0,
				},
				time.Unix(0, 0),
				Dana.Counter,
			),
			expected: []byte(`
# HELP http_requests_total Dana2 collected metric
# TYPE http_requests_total counter
http_requests_total{code="400",method="post"} 3
`),
		},
		{
			name: "prometheus input gauge",
			metric: testutil.MustMetric(
				"prometheus",
				map[string]string{
					"code":   "400",
					"method": "post",
				},
				map[string]interface{}{
					"http_requests_total": 3.0,
				},
				time.Unix(0, 0),
				Dana.Gauge,
			),
			expected: []byte(`
# HELP http_requests_total Dana2 collected metric
# TYPE http_requests_total gauge
http_requests_total{code="400",method="post"} 3
`),
		},
		{
			name: "prometheus input histogram no buckets",
			metric: testutil.MustMetric(
				"prometheus",
				map[string]string{},
				map[string]interface{}{
					"http_request_duration_seconds_sum":   53423,
					"http_request_duration_seconds_count": 144320,
				},
				time.Unix(0, 0),
				Dana.Histogram,
			),
			expected: []byte(`
# HELP http_request_duration_seconds Dana2 collected metric
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="+Inf"} 144320
http_request_duration_seconds_sum 53423
http_request_duration_seconds_count 144320
`),
		},
		{
			name: "prometheus input histogram only bucket",
			metric: testutil.MustMetric(
				"prometheus",
				map[string]string{
					"le": "0.5",
				},
				map[string]interface{}{
					"http_request_duration_seconds_bucket": 129389.0,
				},
				time.Unix(0, 0),
				Dana.Histogram,
			),
			expected: []byte(`
# HELP http_request_duration_seconds Dana2 collected metric
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.5"} 129389
http_request_duration_seconds_bucket{le="+Inf"} 0
http_request_duration_seconds_sum 0
http_request_duration_seconds_count 0
`),
		},
		{
			name: "simple with timestamp",
			config: FormatConfig{
				ExportTimestamp: true,
			},
			metric: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
				},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(1574279268, 0),
			),
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{host="example.org"} 42 1574279268000
`),
		},
		{
			name: "simple with CompactEncoding",
			config: FormatConfig{
				CompactEncoding: true,
			},
			metric: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
				},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(1574279268, 0),
			),
			expected: []byte(`
# TYPE cpu_time_idle untyped
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "untyped forced to counter",
			config: FormatConfig{
				TypeMappings: MetricTypes{Counter: []string{"cpu_time_idle"}},
			},
			metric: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
				},
				map[string]interface{}{
					"time_idle": 42,
				},
				time.Unix(0, 0),
			),
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle counter
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "untyped forced to gauge",
			config: FormatConfig{
				TypeMappings: MetricTypes{Gauge: []string{"cpu_time_idle"}},
			},
			metric: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
				},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle gauge
cpu_time_idle{host="example.org"} 42
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Serializer{
				FormatConfig{
					SortMetrics:     true,
					ExportTimestamp: tt.config.ExportTimestamp,
					StringAsLabel:   tt.config.StringAsLabel,
					CompactEncoding: tt.config.CompactEncoding,
					TypeMappings:    tt.config.TypeMappings,
				},
			}
			require.NoError(t, s.Init())

			actual, err := s.Serialize(tt.metric)
			require.NoError(t, err)

			require.Equal(t, strings.TrimSpace(string(tt.expected)),
				strings.TrimSpace(string(actual)))
		})
	}
}

func TestSerializeBatch(t *testing.T) {
	tests := []struct {
		name     string
		config   FormatConfig
		metrics  []Dana.Metric
		expected []byte
	}{
		{
			name: "simple",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "one.example.org",
					},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "two.example.org",
					},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{host="one.example.org"} 42
cpu_time_idle{host="two.example.org"} 42
`),
		},
		{
			name: "multiple metric families",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "one.example.org",
					},
					map[string]interface{}{
						"time_idle":  42.0,
						"time_guest": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_guest Dana2 collected metric
# TYPE cpu_time_guest untyped
cpu_time_guest{host="one.example.org"} 42
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{host="one.example.org"} 42
`),
		},
		{
			name: "histogram",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"prometheus",
					map[string]string{},
					map[string]interface{}{
						"http_request_duration_seconds_sum":   53423,
						"http_request_duration_seconds_count": 144320,
					},
					time.Unix(0, 0),
					Dana.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "0.05"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 24054.0,
					},
					time.Unix(0, 0),
					Dana.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "0.1"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 33444.0,
					},
					time.Unix(0, 0),
					Dana.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "0.2"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 100392.0,
					},
					time.Unix(0, 0),
					Dana.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "0.5"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 129389.0,
					},
					time.Unix(0, 0),
					Dana.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "1.0"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 133988.0,
					},
					time.Unix(0, 0),
					Dana.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "+Inf"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 144320.0,
					},
					time.Unix(0, 0),
					Dana.Histogram,
				),
			},
			expected: []byte(`
# HELP http_request_duration_seconds Dana2 collected metric
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.05"} 24054
http_request_duration_seconds_bucket{le="0.1"} 33444
http_request_duration_seconds_bucket{le="0.2"} 100392
http_request_duration_seconds_bucket{le="0.5"} 129389
http_request_duration_seconds_bucket{le="1"} 133988
http_request_duration_seconds_bucket{le="+Inf"} 144320
http_request_duration_seconds_sum 53423
http_request_duration_seconds_count 144320
`),
		},
		{
			name: "",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"prometheus",
					map[string]string{},
					map[string]interface{}{
						"rpc_duration_seconds_sum":   1.7560473e+07,
						"rpc_duration_seconds_count": 2693,
					},
					time.Unix(0, 0),
					Dana.Summary,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"quantile": "0.01"},
					map[string]interface{}{
						"rpc_duration_seconds": 3102.0,
					},
					time.Unix(0, 0),
					Dana.Summary,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"quantile": "0.05"},
					map[string]interface{}{
						"rpc_duration_seconds": 3272.0,
					},
					time.Unix(0, 0),
					Dana.Summary,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"quantile": "0.5"},
					map[string]interface{}{
						"rpc_duration_seconds": 4773.0,
					},
					time.Unix(0, 0),
					Dana.Summary,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"quantile": "0.9"},
					map[string]interface{}{
						"rpc_duration_seconds": 9001.0,
					},
					time.Unix(0, 0),
					Dana.Summary,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"quantile": "0.99"},
					map[string]interface{}{
						"rpc_duration_seconds": 76656.0,
					},
					time.Unix(0, 0),
					Dana.Summary,
				),
			},
			expected: []byte(`
# HELP rpc_duration_seconds Dana2 collected metric
# TYPE rpc_duration_seconds summary
rpc_duration_seconds{quantile="0.01"} 3102
rpc_duration_seconds{quantile="0.05"} 3272
rpc_duration_seconds{quantile="0.5"} 4773
rpc_duration_seconds{quantile="0.9"} 9001
rpc_duration_seconds{quantile="0.99"} 76656
rpc_duration_seconds_sum 1.7560473e+07
rpc_duration_seconds_count 2693
`),
		},
		{
			name: "newer sample",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 43.0,
					},
					time.Unix(1, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle 43
`),
		},
		{
			name: "colons are not replaced in metric name from measurement",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu::xyzzy",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu::xyzzy_time_idle Dana2 collected metric
# TYPE cpu::xyzzy_time_idle untyped
cpu::xyzzy_time_idle 42
`),
		},
		{
			name: "colons are not replaced in metric name from field",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time:idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time:idle Dana2 collected metric
# TYPE cpu_time:idle untyped
cpu_time:idle 42
`),
		},
		{
			name: "invalid label",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host-name": "example.org",
					},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{host_name="example.org"} 42
`),
		},
		{
			name: "colons are replaced in label name",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host:name": "example.org",
					},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{host_name="example.org"} 42
`),
		},
		{
			name: "discard strings",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
						"cpu":       "cpu0",
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle 42
`),
		},
		{
			name: "string as label",
			config: FormatConfig{
				StringAsLabel: true,
			},
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
						"cpu":       "cpu0",
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{cpu="cpu0"} 42
`),
		},
		{
			name: "string as label duplicate tag",
			config: FormatConfig{
				StringAsLabel: true,
			},
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 42.0,
						"cpu":       "cpu1",
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{cpu="cpu0"} 42
`),
		},
		{
			name: "replace characters when using string as label",
			config: FormatConfig{
				StringAsLabel: true,
			},
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"host:name": "example.org",
						"time_idle": 42.0,
					},
					time.Unix(1574279268, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Dana2 collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{host_name="example.org"} 42
`),
		},
		{
			name: "multiple fields grouping",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_guest":  8106.04,
						"time_system": 26271.4,
						"time_user":   92904.33,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu1",
					},
					map[string]interface{}{
						"time_guest":  8181.63,
						"time_system": 25351.49,
						"time_user":   96912.57,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu2",
					},
					map[string]interface{}{
						"time_guest":  7470.04,
						"time_system": 24998.43,
						"time_user":   96034.08,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu3",
					},
					map[string]interface{}{
						"time_guest":  7517.95,
						"time_system": 24970.82,
						"time_user":   94148,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_guest Dana2 collected metric
# TYPE cpu_time_guest untyped
cpu_time_guest{cpu="cpu0"} 8106.04
cpu_time_guest{cpu="cpu1"} 8181.63
cpu_time_guest{cpu="cpu2"} 7470.04
cpu_time_guest{cpu="cpu3"} 7517.95
# HELP cpu_time_system Dana2 collected metric
# TYPE cpu_time_system untyped
cpu_time_system{cpu="cpu0"} 26271.4
cpu_time_system{cpu="cpu1"} 25351.49
cpu_time_system{cpu="cpu2"} 24998.43
cpu_time_system{cpu="cpu3"} 24970.82
# HELP cpu_time_user Dana2 collected metric
# TYPE cpu_time_user untyped
cpu_time_user{cpu="cpu0"} 92904.33
cpu_time_user{cpu="cpu1"} 96912.57
cpu_time_user{cpu="cpu2"} 96034.08
cpu_time_user{cpu="cpu3"} 94148
`),
		},
		{
			name: "summary with no quantile",
			metrics: []Dana.Metric{
				testutil.MustMetric(
					"prometheus",
					map[string]string{},
					map[string]interface{}{
						"rpc_duration_seconds_sum":   1.7560473e+07,
						"rpc_duration_seconds_count": 2693,
					},
					time.Unix(0, 0),
					Dana.Summary,
				),
			},
			expected: []byte(`
# HELP rpc_duration_seconds Dana2 collected metric
# TYPE rpc_duration_seconds summary
rpc_duration_seconds_sum 1.7560473e+07
rpc_duration_seconds_count 2693
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Serializer{
				FormatConfig{
					SortMetrics:     true,
					ExportTimestamp: tt.config.ExportTimestamp,
					StringAsLabel:   tt.config.StringAsLabel,
				},
			}
			actual, err := s.SerializeBatch(tt.metrics)
			require.NoError(t, err)

			require.Equal(t,
				strings.TrimSpace(string(tt.expected)),
				strings.TrimSpace(string(actual)))
		})
	}
}

func BenchmarkSerialize(b *testing.B) {
	s := &Serializer{}
	require.NoError(b, s.Init())
	metrics := serializers.BenchmarkMetrics(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.Serialize(metrics[i%len(metrics)])
		require.NoError(b, err)
	}
}

func BenchmarkSerializeBatch(b *testing.B) {
	s := &Serializer{}
	require.NoError(b, s.Init())
	m := serializers.BenchmarkMetrics(b)
	metrics := m[:]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.SerializeBatch(metrics)
		require.NoError(b, err)
	}
}
