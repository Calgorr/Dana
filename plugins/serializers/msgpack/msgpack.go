package msgpack

import (
	"Dana"
	"Dana/plugins/serializers"
)

// Serializer encodes metrics in MessagePack format
type Serializer struct{}

func marshalMetric(buf []byte, metric Dana.Metric) ([]byte, error) {
	return (&Metric{
		Name:   metric.Name(),
		Time:   MessagePackTime{time: metric.Time()},
		Tags:   metric.Tags(),
		Fields: metric.Fields(),
	}).MarshalMsg(buf)
}

// Serialize implements serializers.Serializer.Serialize
// github.com/influxdata/telegraf/plugins/serializers/Serializer
func (s *Serializer) Serialize(metric Dana.Metric) ([]byte, error) {
	return marshalMetric(nil, metric)
}

// SerializeBatch implements serializers.Serializer.SerializeBatch
// github.com/influxdata/telegraf/plugins/serializers/Serializer
func (s *Serializer) SerializeBatch(metrics []Dana.Metric) ([]byte, error) {
	buf := make([]byte, 0)
	for _, m := range metrics {
		var err error
		buf, err = marshalMetric(buf, m)

		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func init() {
	serializers.Add("msgpack",
		func() Dana.Serializer {
			return &Serializer{}
		},
	)
}
