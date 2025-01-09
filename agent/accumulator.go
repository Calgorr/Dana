package agent

import (
	"time"

	"Dana"
	"Dana/metric"
)

type MetricMaker interface {
	LogName() string
	MakeMetric(m Dana.Metric) Dana.Metric
	Log() Dana.Logger
}

type accumulator struct {
	maker     MetricMaker
	metrics   chan<- Dana.Metric
	precision time.Duration
}

func NewAccumulator(
	maker MetricMaker,
	metrics chan<- Dana.Metric,
) Dana.Accumulator {
	acc := accumulator{
		maker:     maker,
		metrics:   metrics,
		precision: time.Nanosecond,
	}
	return &acc
}

func (ac *accumulator) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addMeasurement(measurement, tags, fields, Dana.Untyped, t...)
}

func (ac *accumulator) AddGauge(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addMeasurement(measurement, tags, fields, Dana.Gauge, t...)
}

func (ac *accumulator) AddCounter(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addMeasurement(measurement, tags, fields, Dana.Counter, t...)
}

func (ac *accumulator) AddSummary(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addMeasurement(measurement, tags, fields, Dana.Summary, t...)
}

func (ac *accumulator) AddHistogram(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addMeasurement(measurement, tags, fields, Dana.Histogram, t...)
}

func (ac *accumulator) AddMetric(m Dana.Metric) {
	m.SetTime(m.Time().Round(ac.precision))
	if m := ac.maker.MakeMetric(m); m != nil {
		ac.metrics <- m
	}
}

func (ac *accumulator) addMeasurement(
	measurement string,
	tags map[string]string,
	fields map[string]interface{},
	tp Dana.ValueType,
	t ...time.Time,
) {
	m := metric.New(measurement, tags, fields, ac.getTime(t), tp)
	if m := ac.maker.MakeMetric(m); m != nil {
		ac.metrics <- m
	}
}

// AddError passes a runtime error to the accumulator.
// The error will be tagged with the plugin name and written to the log.
func (ac *accumulator) AddError(err error) {
	if err == nil {
		return
	}
	ac.maker.Log().Errorf("Error in plugin: %v", err)
}

func (ac *accumulator) SetPrecision(precision time.Duration) {
	ac.precision = precision
}

func (ac *accumulator) getTime(t []time.Time) time.Time {
	var timestamp time.Time
	if len(t) > 0 {
		timestamp = t[0]
	} else {
		timestamp = time.Now()
	}
	return timestamp.Round(ac.precision)
}

func (ac *accumulator) WithTracking(maxTracked int) Dana.TrackingAccumulator {
	return &trackingAccumulator{
		Accumulator: ac,
		delivered:   make(chan Dana.DeliveryInfo, maxTracked),
	}
}

type trackingAccumulator struct {
	Dana.Accumulator
	delivered chan Dana.DeliveryInfo
}

func (a *trackingAccumulator) AddTrackingMetric(m Dana.Metric) Dana.TrackingID {
	dm, id := metric.WithTracking(m, a.onDelivery)
	a.AddMetric(dm)
	return id
}

func (a *trackingAccumulator) AddTrackingMetricGroup(group []Dana.Metric) Dana.TrackingID {
	db, id := metric.WithGroupTracking(group, a.onDelivery)
	for _, m := range db {
		a.AddMetric(m)
	}
	return id
}

func (a *trackingAccumulator) Delivered() <-chan Dana.DeliveryInfo {
	return a.delivered
}

func (a *trackingAccumulator) onDelivery(info Dana.DeliveryInfo) {
	select {
	case a.delivered <- info:
	default:
		// This is a programming error in the input.  More items were sent for
		// tracking than space requested.
		panic("channel is full")
	}
}
