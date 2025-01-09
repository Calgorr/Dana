package prometheus

import (
	"Dana"
	dto "github.com/prometheus/client_model/go"
)

func mapValueType(mt dto.MetricType) Dana.ValueType {
	switch mt {
	case dto.MetricType_COUNTER:
		return Dana.Counter
	case dto.MetricType_GAUGE:
		return Dana.Gauge
	case dto.MetricType_SUMMARY:
		return Dana.Summary
	case dto.MetricType_HISTOGRAM:
		return Dana.Histogram
	default:
		return Dana.Untyped
	}
}

func getTagsFromLabels(m *dto.Metric, defaultTags map[string]string) map[string]string {
	result := make(map[string]string, len(defaultTags)+len(m.Label))
	for key, value := range defaultTags {
		result[key] = value
	}

	for _, label := range m.Label {
		if v := label.GetValue(); v != "" {
			result[label.GetName()] = v
		}
	}

	return result
}
