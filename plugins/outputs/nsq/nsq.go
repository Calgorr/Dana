//go:generate ../../../tools/readme_config_includer/generator
package nsq

import (
	_ "embed"
	"fmt"

	"github.com/nsqio/go-nsq"

	"Dana"
	"Dana/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type NSQ struct {
	Server string
	Topic  string
	Log    Dana.Logger `toml:"-"`

	producer   *nsq.Producer
	serializer Dana.Serializer
}

func (*NSQ) SampleConfig() string {
	return sampleConfig
}

func (n *NSQ) SetSerializer(serializer Dana.Serializer) {
	n.serializer = serializer
}

func (n *NSQ) Connect() error {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(n.Server, config)

	if err != nil {
		return err
	}

	n.producer = producer
	return nil
}

func (n *NSQ) Close() error {
	n.producer.Stop()
	return nil
}

func (n *NSQ) Write(metrics []Dana.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		buf, err := n.serializer.Serialize(metric)
		if err != nil {
			n.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		err = n.producer.Publish(n.Topic, buf)
		if err != nil {
			return fmt.Errorf("failed to send NSQD message: %w", err)
		}
	}
	return nil
}

func init() {
	outputs.Add("nsq", func() Dana.Output {
		return &NSQ{}
	})
}
