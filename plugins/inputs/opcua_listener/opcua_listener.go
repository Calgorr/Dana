//go:generate ../../../tools/readme_config_includer/generator
package opcua_listener

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"Dana"
	"Dana/config"
	"Dana/plugins/common/opcua"
	"Dana/plugins/common/opcua/input"
	"Dana/plugins/inputs"
)

type OpcUaListener struct {
	subscribeClientConfig
	client *subscribeClient
	Log    Dana.Logger `toml:"-"`
}

//go:embed sample.conf
var sampleConfig string

func (*OpcUaListener) SampleConfig() string {
	return sampleConfig
}

func (o *OpcUaListener) Init() (err error) {
	switch o.ConnectFailBehavior {
	case "":
		o.ConnectFailBehavior = "error"
	case "error", "ignore", "retry":
		// Do nothing as these are valid
	default:
		return fmt.Errorf("unknown setting %q for 'connect_fail_behavior'", o.ConnectFailBehavior)
	}
	o.client, err = o.subscribeClientConfig.createSubscribeClient(o.Log)
	return err
}

func (o *OpcUaListener) Start(acc Dana.Accumulator) error {
	return o.connect(acc)
}

func (o *OpcUaListener) Gather(acc Dana.Accumulator) error {
	if o.client.State() == opcua.Connected || o.subscribeClientConfig.ConnectFailBehavior == "ignore" {
		return nil
	}
	return o.connect(acc)
}

func (o *OpcUaListener) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	select {
	case <-o.client.stop(ctx):
		o.Log.Infof("Unsubscribed OPC UA successfully")
	case <-ctx.Done(): // Timeout context
		o.Log.Warn("Timeout while stopping OPC UA subscription")
	}
	cancel()
}

func (o *OpcUaListener) connect(acc Dana.Accumulator) error {
	ctx := context.Background()
	ch, err := o.client.startStreamValues(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			m, ok := <-ch
			if !ok {
				o.Log.Debug("Metric collection stopped due to closed channel")
				return
			}
			acc.AddMetric(m)
		}
	}()

	return nil
}

func init() {
	inputs.Add("opcua_listener", func() Dana.Input {
		return &OpcUaListener{
			subscribeClientConfig: subscribeClientConfig{
				InputClientConfig: input.InputClientConfig{
					OpcUAClientConfig: opcua.OpcUAClientConfig{
						Endpoint:       "opc.tcp://localhost:4840",
						SecurityPolicy: "auto",
						SecurityMode:   "auto",
						Certificate:    "/etc/Dana2/cert.pem",
						PrivateKey:     "/etc/Dana2/key.pem",
						AuthMethod:     "Anonymous",
						ConnectTimeout: config.Duration(5 * time.Second),
						RequestTimeout: config.Duration(10 * time.Second),
					},
					MetricName: "opcua",
					Timestamp:  input.TimestampSourceDana2,
				},
				SubscriptionInterval: config.Duration(100 * time.Millisecond),
			},
		}
	})
}
