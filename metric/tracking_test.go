package metric

import (
	"sync"
	"testing"
	"time"

	"Dana"
	"github.com/stretchr/testify/require"
)

func mustMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	tm time.Time,
	tp ...Dana.ValueType,
) Dana.Metric {
	m := New(name, tags, fields, tm, tp...)
	return m
}

type deliveries struct {
	Info map[Dana.TrackingID]Dana.DeliveryInfo
}

func (d *deliveries) onDelivery(info Dana.DeliveryInfo) {
	d.Info[info.ID()] = info
}

func TestNewTrackingID(t *testing.T) {
	var wg sync.WaitGroup
	var a [100000]Dana.TrackingID
	var b [100000]Dana.TrackingID

	wg.Add(2)
	go func() {
		for i := 0; i < len(a); i++ {
			a[i] = newTrackingID()
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < len(b); i++ {
			b[i] = newTrackingID()
		}
		wg.Done()
	}()
	wg.Wait()

	// Find any duplicate TrackingIDs in arrays a and b. Arrays must be sorted in increasing order.
	for i, j := 0, 0; i < len(a) && j < len(b); {
		if a[i] == b[j] {
			t.Errorf("Duplicate TrackingID: a[%d]==%d and b[%d]==%d.", i, a[i], j, b[j])
			break
		}
		if a[i] > b[j] {
			j++
			continue
		}
		if a[i] < b[j] {
			i++
			continue
		}
	}
}

func TestTracking(t *testing.T) {
	tests := []struct {
		name      string
		metric    Dana.Metric
		actions   func(metric Dana.Metric)
		delivered bool
	}{
		{
			name: "accept",
			metric: mustMetric(
				"memory",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
				Dana.Gauge,
			),
			actions: func(m Dana.Metric) {
				m.Accept()
			},
			delivered: true,
		},
		{
			name: "reject",
			metric: mustMetric(
				"memory",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
				Dana.Gauge,
			),
			actions: func(m Dana.Metric) {
				m.Reject()
			},
			delivered: false,
		},
		{
			name: "accept copy",
			metric: mustMetric(
				"memory",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
				Dana.Gauge,
			),
			actions: func(m Dana.Metric) {
				m2 := m.Copy()
				m.Accept()
				m2.Accept()
			},
			delivered: true,
		},
		{
			name: "copy with accept and done",
			metric: mustMetric(
				"memory",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
				Dana.Gauge,
			),
			actions: func(m Dana.Metric) {
				m2 := m.Copy()
				m.Accept()
				m2.Drop()
			},
			delivered: true,
		},
		{
			name: "copy with mixed delivery",
			metric: mustMetric(
				"memory",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
				Dana.Gauge,
			),
			actions: func(m Dana.Metric) {
				m2 := m.Copy()
				m.Accept()
				m2.Reject()
			},
			delivered: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deliveries{
				Info: make(map[Dana.TrackingID]Dana.DeliveryInfo),
			}
			metric, id := WithTracking(tt.metric, d.onDelivery)
			tt.actions(metric)

			info := d.Info[id]
			require.Equal(t, tt.delivered, info.Delivered())
		})
	}
}

func TestGroupTracking(t *testing.T) {
	tests := []struct {
		name      string
		metrics   []Dana.Metric
		actions   func(metrics []Dana.Metric)
		delivered bool
	}{
		{
			name: "accept",
			metrics: []Dana.Metric{
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			actions: func(metrics []Dana.Metric) {
				metrics[0].Accept()
				metrics[1].Accept()
			},
			delivered: true,
		},
		{
			name: "reject",
			metrics: []Dana.Metric{
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			actions: func(metrics []Dana.Metric) {
				metrics[0].Reject()
				metrics[1].Reject()
			},
			delivered: false,
		},
		{
			name: "remove",
			metrics: []Dana.Metric{
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			actions: func(metrics []Dana.Metric) {
				metrics[0].Drop()
				metrics[1].Drop()
			},
			delivered: true,
		},
		{
			name: "mixed",
			metrics: []Dana.Metric{
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			actions: func(metrics []Dana.Metric) {
				metrics[0].Accept()
				metrics[1].Reject()
			},
			delivered: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deliveries{
				Info: make(map[Dana.TrackingID]Dana.DeliveryInfo),
			}
			metrics, id := WithGroupTracking(tt.metrics, d.onDelivery)
			tt.actions(metrics)

			info := d.Info[id]
			require.Equal(t, tt.delivered, info.Delivered())
		})
	}
}
