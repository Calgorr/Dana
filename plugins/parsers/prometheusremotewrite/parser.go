package prometheusremotewrite

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"

	"Dana"
	"Dana/metric"
	"Dana/plugins/parsers"
)

type Parser struct {
	DefaultTags map[string]string
}

func (p *Parser) Parse(buf []byte) ([]Dana.Metric, error) {
	var err error
	var metrics []Dana.Metric
	var req prompb.WriteRequest

	if err := req.Unmarshal(buf); err != nil {
		return nil, fmt.Errorf("unable to unmarshal request body: %w", err)
	}

	now := time.Now()

	for _, ts := range req.Timeseries {
		tags := make(map[string]string, len(p.DefaultTags)+len(ts.Labels))
		for key, value := range p.DefaultTags {
			tags[key] = value
		}

		for _, l := range ts.Labels {
			tags[l.Name] = l.Value
		}

		metricName := tags[model.MetricNameLabel]
		if metricName == "" {
			return nil, fmt.Errorf("metric name %q not found in tag-set or empty", model.MetricNameLabel)
		}
		delete(tags, model.MetricNameLabel)
		t := now
		for _, s := range ts.Samples {
			fields := make(map[string]interface{})
			if !math.IsNaN(s.Value) {
				fields[metricName] = s.Value
			}
			// converting to Dana2 metric
			if len(fields) > 0 {
				if s.Timestamp > 0 {
					t = time.Unix(0, s.Timestamp*1000000)
				}
				m := metric.New("prometheus_remote_write", tags, fields, t)
				metrics = append(metrics, m)
			}
		}

		for _, hp := range ts.Histograms {
			h := hp.ToFloatHistogram()

			if hp.Timestamp > 0 {
				t = time.Unix(0, hp.Timestamp*1000000)
			}

			fields := map[string]any{
				metricName + "_sum": h.Sum,
			}
			m := metric.New("prometheus_remote_write", tags, fields, t)
			metrics = append(metrics, m)

			fields = map[string]any{
				metricName + "_count": h.Count,
			}
			m = metric.New("prometheus_remote_write", tags, fields, t)
			metrics = append(metrics, m)

			count := 0.0
			iter := h.AllBucketIterator()
			for iter.Next() {
				bucket := iter.At()

				count = count + bucket.Count
				fields = map[string]any{
					metricName: count,
				}

				localTags := make(map[string]string, len(tags)+1)
				localTags[metricName+"_le"] = fmt.Sprintf("%g", bucket.Upper)
				for k, v := range tags {
					localTags[k] = v
				}

				m := metric.New("prometheus_remote_write", localTags, fields, t)
				metrics = append(metrics, m)
			}
		}
	}
	return metrics, err
}

func (p *Parser) ParseLine(line string) (Dana.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, errors.New("no metrics in line")
	}

	if len(metrics) > 1 {
		return nil, errors.New("more than one metric in line")
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func init() {
	parsers.Add("prometheusremotewrite",
		func(string) Dana.Parser {
			return &Parser{}
		})
}
