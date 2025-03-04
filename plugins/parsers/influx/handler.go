package influx

import (
	"bytes"
	"errors"
	"strconv"
	"time"

	"Dana"
	"Dana/metric"
)

// MetricHandler implements the Handler interface and produces Dana.Metric.
type MetricHandler struct {
	timePrecision time.Duration
	timeFunc      TimeFunc
	metric        Dana.Metric
}

func NewMetricHandler() *MetricHandler {
	return &MetricHandler{
		timePrecision: time.Nanosecond,
		timeFunc:      time.Now,
	}
}

func (h *MetricHandler) SetTimePrecision(p time.Duration) {
	h.timePrecision = p
	// When the timestamp is omitted from the metric, the timestamp
	// comes from the server clock, truncated to the nearest unit of
	// measurement provided in precision.
	//
	// When a timestamp is provided in the metric, precision is
	// overloaded to hold the unit of measurement of the timestamp.
}

func (h *MetricHandler) SetTimeFunc(f TimeFunc) {
	h.timeFunc = f
}

func (h *MetricHandler) Metric() Dana.Metric {
	if h.metric.Time().IsZero() {
		h.metric.SetTime(h.timeFunc().Truncate(h.timePrecision))
	}
	return h.metric
}

func (h *MetricHandler) SetMeasurement(name []byte) error {
	h.metric = metric.New(nameUnescape(name),
		nil, nil, time.Time{})
	return nil
}

func (h *MetricHandler) AddTag(key, value []byte) error {
	tk := unescape(key)
	tv := unescape(value)
	h.metric.AddTag(tk, tv)
	return nil
}

func (h *MetricHandler) AddInt(key, value []byte) error {
	fk := unescape(key)
	fv, err := parseIntBytes(bytes.TrimSuffix(value, []byte("i")), 10, 64)
	if err != nil {
		var numErr *strconv.NumError
		if errors.As(err, &numErr) {
			return numErr.Err
		}
		return err
	}
	h.metric.AddField(fk, fv)
	return nil
}

func (h *MetricHandler) AddUint(key, value []byte) error {
	fk := unescape(key)
	fv, err := parseUintBytes(bytes.TrimSuffix(value, []byte("u")), 10, 64)
	if err != nil {
		var numErr *strconv.NumError
		if errors.As(err, &numErr) {
			return numErr.Err
		}
		return err
	}
	h.metric.AddField(fk, fv)
	return nil
}

func (h *MetricHandler) AddFloat(key, value []byte) error {
	fk := unescape(key)
	fv, err := parseFloatBytes(value, 64)
	if err != nil {
		var numErr *strconv.NumError
		if errors.As(err, &numErr) {
			return numErr.Err
		}
		return err
	}
	h.metric.AddField(fk, fv)
	return nil
}

func (h *MetricHandler) AddString(key, value []byte) error {
	fk := unescape(key)
	fv := stringFieldUnescape(value)
	h.metric.AddField(fk, fv)
	return nil
}

func (h *MetricHandler) AddBool(key, value []byte) error {
	fk := unescape(key)
	fv, err := parseBoolBytes(value)
	if err != nil {
		return errors.New("unparsable bool")
	}
	h.metric.AddField(fk, fv)
	return nil
}

func (h *MetricHandler) SetTimestamp(tm []byte) error {
	v, err := parseIntBytes(tm, 10, 64)
	if err != nil {
		var numErr *strconv.NumError
		if errors.As(err, &numErr) {
			return numErr.Err
		}
		return err
	}

	// time precision is overloaded to mean time unit here
	ns := v * int64(h.timePrecision)
	h.metric.SetTime(time.Unix(0, ns))
	return nil
}
