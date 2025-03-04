package splunkmetric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"Dana"
	"Dana/metric"
	"Dana/plugins/serializers"
)

func TestSerializeMetricFloat(t *testing.T) {
	// Test sub-second time
	now := time.Unix(1529875740, 819000000)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m := metric.New("cpu", tags, fields, now)

	s := &Serializer{}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)
	expS := `{"_value":91.5,"cpu":"cpu0","metric_name":"cpu.usage_idle","time":1529875740.819}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricFloatHec(t *testing.T) {
	// Test sub-second time
	now := time.Unix(1529875740, 819000000)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m := metric.New("cpu", tags, fields, now)

	s := &Serializer{HecRouting: true}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)
	expS := `{"time":1529875740.819,"event":"metric","fields":{"_value":91.5,"cpu":"cpu0","metric_name":"cpu.usage_idle"}}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricInt(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m := metric.New("cpu", tags, fields, now)

	s := &Serializer{}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := `{"_value":90,"cpu":"cpu0","metric_name":"cpu.usage_idle","time":0}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricIntHec(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m := metric.New("cpu", tags, fields, now)

	s := &Serializer{HecRouting: true}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := `{"time":0,"event":"metric","fields":{"_value":90,"cpu":"cpu0","metric_name":"cpu.usage_idle"}}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricBool(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"container-name": "Dana2-test",
	}
	fields := map[string]interface{}{
		"oomkiller": true,
	}
	m := metric.New("docker", tags, fields, now)

	s := &Serializer{}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := `{"_value":1,"container-name":"Dana2-test","metric_name":"docker.oomkiller","time":0}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricBoolHec(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"container-name": "Dana2-test",
	}
	fields := map[string]interface{}{
		"oomkiller": false,
	}
	m := metric.New("docker", tags, fields, now)

	s := &Serializer{HecRouting: true}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := `{"time":0,"event":"metric","fields":{"_value":0,"container-name":"Dana2-test","metric_name":"docker.oomkiller"}}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMetricString(t *testing.T) {
	now := time.Unix(0, 0)
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"processorType": "ARMv7 Processor rev 4 (v7l)",
		"usage_idle":    int64(5),
	}
	m := metric.New("cpu", tags, fields, now)

	s := &Serializer{}
	var buf []byte
	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expS := `{"_value":5,"cpu":"cpu0","metric_name":"cpu.usage_idle","time":0}`
	require.Equal(t, expS, string(buf))
	require.NoError(t, err)
}

func TestSerializeBatch(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)

	n := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 92.0,
		},
		time.Unix(0, 0),
	)

	metrics := []Dana.Metric{m, n}
	s := &Serializer{}
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	expS := `{"_value":42,"metric_name":"cpu.value","time":0}{"_value":92,"metric_name":"cpu.value","time":0}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMulti(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"user":   42.0,
			"system": 8.0,
		},
		time.Unix(0, 0),
	)

	metrics := []Dana.Metric{m}
	s := &Serializer{MultiMetric: true}
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	expS := `{"metric_name:cpu.system":8,"metric_name:cpu.user":42,"time":0}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeBatchHec(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	n := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 92.0,
		},
		time.Unix(0, 0),
	)
	metrics := []Dana.Metric{m, n}
	s := &Serializer{HecRouting: true}
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	expS := `{"time":0,"event":"metric","fields":{"_value":42,"metric_name":"cpu.value"}}` +
		`{"time":0,"event":"metric","fields":{"_value":92,"metric_name":"cpu.value"}}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeMultiHec(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"usage":  42.0,
			"system": 8.0,
		},
		time.Unix(0, 0),
	)

	metrics := []Dana.Metric{m}
	s := &Serializer{
		HecRouting:  true,
		MultiMetric: true,
	}
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	expS := `{"time":0,"event":"metric","fields":{"metric_name:cpu.system":8,"metric_name:cpu.usage":42}}`
	require.Equal(t, expS, string(buf))
}

func TestSerializeOmitEvent(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"usage":  42.0,
			"system": 8.0,
		},
		time.Unix(0, 0),
	)

	metrics := []Dana.Metric{m}
	s := &Serializer{
		HecRouting:   true,
		MultiMetric:  true,
		OmitEventTag: true,
	}
	buf, err := s.SerializeBatch(metrics)
	require.NoError(t, err)

	expS := `{"time":0,"fields":{"metric_name:cpu.system":8,"metric_name:cpu.usage":42}}`
	require.Equal(t, expS, string(buf))
}

func BenchmarkSerialize(b *testing.B) {
	s := &Serializer{}
	metrics := serializers.BenchmarkMetrics(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.Serialize(metrics[i%len(metrics)])
		require.NoError(b, err)
	}
}

func BenchmarkSerializeBatch(b *testing.B) {
	s := &Serializer{}
	m := serializers.BenchmarkMetrics(b)
	metrics := m[:]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.SerializeBatch(metrics)
		require.NoError(b, err)
	}
}
