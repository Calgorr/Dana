package zabbix

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"Dana"
	"Dana/config"
	"Dana/testutil"
)

type (
	// Operations is an interface to simulate aggregator operations
	Operations interface{}
	// OperationAdd is an array of metrics to add to the aggregator
	OperationAdd []Dana.Metric
	// OperationPush simulate a push call to the aggregator
	OperationPush struct{}
	// OperationCheck is an array of metrics to check if they are generated by the aggregator
	OperationCheck []Dana.Metric
	// OperationCrossClearIntervalTime is used to simulate a time cross the clear interval
	OperationCrossClearIntervalTime struct{}
)

func TestAddAndPush(t *testing.T) {
	tests := map[string][]Operations{
		"metric without extra tags does not generate LLD metric": {
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA"},
					map[string]interface{}{"value": 1},
					time.Now()),
			},
			OperationPush{},
		},
		"simple Add, Push and check generated LLD metric": {
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "bar"},
					map[string]interface{}{"value": 1},
					time.Now()),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
		},
		"same metric with different tag values": {
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "bar1"},
					map[string]interface{}{"value": 1},
					time.Now()),
			},
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "bar2"},
					map[string]interface{}{"value": 1},
					time.Now()),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar1"},{"{#FOO}":"bar2"}]}`},
					time.Now(),
				),
			},
		},
		"add two metrics, Push and check generated LLD metric": {
			OperationAdd{
				testutil.MustMetric(
					"nameA",
					map[string]string{"host": "hostA", "foo": "bar"},
					map[string]interface{}{"value": 1},
					time.Now()),
				testutil.MustMetric(
					"nameB",
					map[string]string{"host": "hostA", "foo": "bar"},
					map[string]interface{}{"value": 1},
					time.Now()),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"nameA.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"nameB.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
		},
		"add two similar metrics, one with one more extra tag": {
			OperationAdd{
				testutil.MustMetric(
					"nameA",
					map[string]string{"host": "hostA", "foo1": "bar"},
					map[string]interface{}{"value": 1},
					time.Now()),
				testutil.MustMetric(
					"nameA",
					map[string]string{"host": "hostA", "foo1": "bar", "foo2": "baz"},
					map[string]interface{}{"value": 1},
					time.Now()),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"nameA.foo1": `{"data":[{"{#FOO1}":"bar"}]}`},
					time.Now(),
				),
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"nameA.foo1.foo2": `{"data":[{"{#FOO1}":"bar","{#FOO2}":"baz"}]}`},
					time.Now(),
				),
			},
		},
		"same metric several times generate only one LLD": {
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "bar"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "bar"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "bar"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
		},
		"same metric several times, with different tag ordering, generate only one LLD": {
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "bar", "baz": "qux"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "baz": "qux", "foo": "bar"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
				testutil.MustMetric(
					"name",
					map[string]string{"baz": "qux", "foo": "bar", "host": "hostA"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.baz.foo": `{"data":[{"{#BAZ}":"qux","{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
		},
		"after sending correctly an LLD, same tag values does not generate the same LLD": {
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
		},
		"after lld_clear_interval, already seen LLDs could be resend": {
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			OperationCrossClearIntervalTime{}, // The clear of the previous LLD seen is done in the next push
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
		},
		"clear interval does not interfere with the send of empty LLDs": {
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			// In this interval between push, the metric is not received
			OperationCrossClearIntervalTime{}, // The clear of the previous LLD seen is done in the next push
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[]}`},
					time.Now(),
				),
			},
		},
		"one metric changes the value of the tag, it should send the new value and not send and empty lld": {
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar1"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar1"}]}`},
					time.Now(),
				),
			},
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar2"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar2"}]}`},
					time.Now(),
				),
			},
		},
		"if one input stop sending metrics, an empty LLD is sent": {
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
			OperationPush{},
			OperationCheck{testutil.MustMetric(
				lldName,
				map[string]string{"host": "hostA"},
				map[string]interface{}{"name.foo": `{"data":[]}`},
				time.Now(),
			)},
		},
		"from two inputs, one stop sending metrics, an empty LLD is sent just for that stopped input": {
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "bar"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostB", "foo": "bar"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostB"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostB", "foo": "bar"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[]}`},
					time.Now(),
				),
			},
		},
		"different hosts sending the same metric should generate different LLDs": {
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostB", "foo": "bar"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostB"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
		},
		"same measurement with different tags should generate different LLDs": {
			OperationAdd{testutil.MustMetric(
				"name",
				map[string]string{"host": "hostA", "foo": "a"},
				map[string]interface{}{"value": 1},
				time.Now(),
			)},
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "a", "bar": "b"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"a"}]}`},
					time.Now(),
				),
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.bar.foo": `{"data":[{"{#BAR}":"b","{#FOO}":"a"}]}`},
					time.Now(),
				),
			},
		},
		"a set with a new combination of tag values already seen should generate a new lld": {
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "a", "bar": "b"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "x", "bar": "y"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "a", "bar": "y"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{
						"name.bar.foo": `{"data":[{"{#BAR}":"b","{#FOO}":"a"},{"{#BAR}":"y","{#FOO}":"a"},{"{#BAR}":"y","{#FOO}":"x"}]}`,
					},
					time.Now(),
				),
			},
		},
		"same host and metric with and without extra tag": {
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "a"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{
						"name.foo": `{"data":[{"{#FOO}":"a"}]}`,
					},
					time.Now(),
				),
			},
			OperationCrossClearIntervalTime{},
			OperationPush{},
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA", "foo": "a"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{
						"name.foo": `{"data":[{"{#FOO}":"a"}]}`,
					},
					time.Now(),
				),
			},
			OperationAdd{
				testutil.MustMetric(
					"name",
					map[string]string{"host": "hostA"},
					map[string]interface{}{"value": 1},
					time.Now(),
				),
			},
			OperationPush{},
			// Clean name.foo because it has not been since the last push
			OperationCheck{
				testutil.MustMetric(
					lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{
						"name.foo": `{"data":[]}`,
					},
					time.Now(),
				),
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			zl := zabbixLLD{
				log:           testutil.Logger{},
				clearInterval: config.Duration(time.Hour),
				lastClear:     time.Now(),
				hostTag:       "host",
				current:       make(map[uint64]lldInfo),
			}

			var metrics []Dana.Metric
			for _, op := range test {
				switch o := (op).(type) {
				case OperationAdd:
					for _, m := range o {
						require.NoError(t, zl.Add(m))
					}
				case OperationPush:
					metrics = zl.Push()
				case OperationCheck:
					metrics = sortMetricJSONData(metrics)
					testutil.RequireMetricsEqual(t, o, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
				case OperationCrossClearIntervalTime:
					// Simulate the time passing by and crossing the clear interval time.
					// Add an extra millisecond to be sure to cross the interval in the next operation.
					zl.lastClear = time.Now().Add(-time.Duration(zl.clearInterval)).Add(-time.Millisecond)
				}
			}
		})
	}
}

func TestPush(t *testing.T) {
	tests := map[string]struct {
		ReceivedData         map[uint64]lldInfo
		PreviousReceivedData map[uint64]lldInfo
		Metrics              []Dana.Metric
	}{
		"an empty ReceivedData does not generate any metric": {
			ReceivedData:         map[uint64]lldInfo{},
			PreviousReceivedData: map[uint64]lldInfo{},
		},
		"simple one host with one lld with one set of values": {
			ReceivedData: map[uint64]lldInfo{
				0: {
					Hostname: "hostA",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						1: {
							"{#FOO}": "bar",
						},
					},
				},
			},
			PreviousReceivedData: map[uint64]lldInfo{},
			Metrics: []Dana.Metric{
				testutil.MustMetric(lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
		},
		"one host with one lld with two set of values": {
			ReceivedData: map[uint64]lldInfo{
				0: {
					Hostname: "hostA",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						1: {
							"{#FOO}": "bar1",
						},
						2: {
							"{#FOO}": "bar2",
						},
					},
				},
			},
			PreviousReceivedData: map[uint64]lldInfo{},
			Metrics: []Dana.Metric{
				testutil.MustMetric(lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.foo": `{"data":[{"{#FOO}":"bar1"},{"{#FOO}":"bar2"}]}`},
					time.Now(),
				),
			},
		},
		"one host with one lld with one multiset of values": {
			ReceivedData: map[uint64]lldInfo{
				0: {
					Hostname: "hostA",
					Key:      "disk.fooA.fooB.fooC",
					Data: map[uint64]map[string]string{
						1: {
							"{#FOOA}": "bar1",
							"{#FOOB}": "bar2",
							"{#FOOC}": "bar3",
						},
					},
				},
			},
			PreviousReceivedData: map[uint64]lldInfo{},
			Metrics: []Dana.Metric{
				testutil.MustMetric(lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.fooA.fooB.fooC": `{"data":[{"{#FOOA}":"bar1","{#FOOB}":"bar2","{#FOOC}":"bar3"}]}`},
					time.Now(),
				),
			},
		},
		"one host with three lld with one set of values, not sorted": {
			ReceivedData: map[uint64]lldInfo{
				0: {
					Hostname: "hostA",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						1: {
							"{#FOO}": "bar",
						},
					},
				},
				1: {
					Hostname: "hostA",
					Key:      "net.iface",
					Data: map[uint64]map[string]string{
						1: {
							"{#IFACE}": "eth0",
						},
					},
				},
				2: {
					Hostname: "hostA",
					Key:      "proc.pid",
					Data: map[uint64]map[string]string{
						1: {
							"{#PID}": "1234",
						},
					},
				},
			},
			PreviousReceivedData: map[uint64]lldInfo{},
			Metrics: []Dana.Metric{
				testutil.MustMetric(lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"proc.pid": `{"data":[{"{#PID}":"1234"}]}`},
					time.Now(),
				),
				testutil.MustMetric(lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
				testutil.MustMetric(lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"net.iface": `{"data":[{"{#IFACE}":"eth0"}]}`},
					time.Now(),
				),
			},
		},
		"two host with the same lld with one set of values": {
			ReceivedData: map[uint64]lldInfo{
				0: {
					Hostname: "hostA",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						1: {
							"{#FOO}": "bar",
						},
					},
				},
				1: {
					Hostname: "hostB",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						1: {
							"{#FOO}": "bar",
						},
					},
				},
			},
			PreviousReceivedData: map[uint64]lldInfo{},
			Metrics: []Dana.Metric{
				testutil.MustMetric(lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
				testutil.MustMetric(lldName,
					map[string]string{"host": "hostB"},
					map[string]interface{}{"disk.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
		},
		"ignore generating a new lld if it was sent the last time": {
			ReceivedData: map[uint64]lldInfo{
				2658406801034663970: {
					Hostname: "hostA",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						1: {
							"{#FOO}": "bar",
						},
					},
				},
			},
			PreviousReceivedData: map[uint64]lldInfo{
				2658406801034663970: {
					Hostname: "hostA",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						1: {
							"{#FOO}": "bar",
						},
					},
				},
			},
		},
		"send an empty LLD if one metric has stopped being sent": {
			ReceivedData: map[uint64]lldInfo{},
			PreviousReceivedData: map[uint64]lldInfo{
				0: {
					Hostname: "hostA",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						1: {
							"{#FOO}": "bar",
						},
					},
				},
			},
			Metrics: []Dana.Metric{
				testutil.MustMetric(lldName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.foo": `{"data":[]}`},
					time.Now(),
				),
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			zl := zabbixLLD{
				clearInterval: 4,
				log:           testutil.Logger{},
				current:       test.ReceivedData,
				previous:      test.PreviousReceivedData,
				hostTag:       "host",
			}

			// Hash the previous data
			for series, info := range zl.previous {
				info.DataHash = info.hash()
				zl.previous[series] = info
			}

			metrics := zl.Push()
			// Sort the "data" dict in the metrics values to get always the same order.
			metrics = sortMetricJSONData(metrics)

			testutil.RequireMetricsEqual(t, test.Metrics, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}

type MetricValue struct {
	Data []map[string]string `json:"data"`
}

// sortMetricJSONData given a list of metrics, if the name is equal to lldName, the JSON data dictionaries are sorted.
// This is needed because the order of the JSON data dictionaries is not guaranteed but we need them sorted to compare
// them against the expected tests values. This sorting should be done using the keys of the dictionaries.
// Example:
//
//	Original metrics: lld,host=foo disk.foo={"data":[{"{#FOO2}":"bar2"},{"{#FOO1}":"bar1"}]}
//	Sorted metrics:   lld,host=foo disk.foo={"data":[{"{#FOO1}":"bar1"},{"{#FOO2}":"bar2"}]}
func sortMetricJSONData(metrics []Dana.Metric) []Dana.Metric {
	for _, m := range metrics {
		if m.Name() == lldName {
			for _, f := range m.FieldList() {
				// f is a string with format: '{"data":[{"{#PID}":"1234"}]}'
				var data MetricValue
				err := json.Unmarshal([]byte(f.Value.(string)), &data)
				if err != nil {
					panic(err)
				}

				// Sort data comparing the content as a string
				sort.Slice(data.Data, func(i, j int) bool {
					return fmt.Sprintf("%v", data.Data[i]) < fmt.Sprintf("%v", data.Data[j])
				})

				dataJSON, err := json.Marshal(data)
				if err != nil {
					panic(err)
				}

				f.Value = string(dataJSON)
			}
		}
	}

	return metrics
}

func TestAdd(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)

	tests := map[string]struct {
		Metrics []Dana.Metric
		Current map[uint64]lldInfo
	}{
		"metric without tags is ignored": {
			Metrics: []Dana.Metric{
				testutil.MustMetric("disk", map[string]string{}, map[string]interface{}{"a": 0}, time.Now()),
			},
			Current: map[uint64]lldInfo{},
		},
		"metric with only the host tag is not used for LLD": {
			Metrics: []Dana.Metric{
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "bar"},
					map[string]interface{}{"a": 0},
					time.Now(),
				),
			},
			Current: map[uint64]lldInfo{},
		},
		"add one metric with one tag and not host tag, use the system hostname": {
			Metrics: []Dana.Metric{
				testutil.MustMetric(
					"disk",
					map[string]string{"foo": "bar"},
					map[string]interface{}{"a": 0},
					time.Now(),
				),
			},
			Current: map[uint64]lldInfo{
				1: {
					Hostname: hostname,
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						2011740591878200733: {
							"{#FOO}": "bar",
						},
					},
				},
			},
		},
		"add one metric with one extra tag": {
			Metrics: []Dana.Metric{
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "bar", "foo": "bar"},
					map[string]interface{}{"a": 0},
					time.Now(),
				),
			},
			Current: map[uint64]lldInfo{
				1: {
					Hostname: "bar",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						13756738031738276742: {
							"{#FOO}": "bar",
						},
					},
				},
			},
		},
		"same metric with different field values is only stored once": {
			Metrics: []Dana.Metric{
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "bar", "foo": "bar"},
					map[string]interface{}{"a": 0},
					time.Now(),
				),
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "bar", "foo": "bar"},
					map[string]interface{}{"a": 999},
					time.Now(),
				),
			},
			Current: map[uint64]lldInfo{
				1: {
					Hostname: "bar",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						13756738031738276742: {
							"{#FOO}": "bar",
						},
					},
				},
			},
		},
		"for the same measurement and tags, the different combinations of tag values are stored under the same key": {
			Metrics: []Dana.Metric{
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "bar", "foo": "bar1"},
					map[string]interface{}{"a": 0},
					time.Now(),
				),
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "bar", "foo": "bar2"},
					map[string]interface{}{"a": 0},
					time.Now(),
				),
			},
			Current: map[uint64]lldInfo{
				1: {
					Hostname: "bar",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						9541037171803204811: {
							"{#FOO}": "bar1",
						},
						10966311568236988310: {
							"{#FOO}": "bar2",
						},
					},
				},
			},
		},
		"same measurement and tags for different hosts are stored in different keys": {
			Metrics: []Dana.Metric{
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "barA", "foo": "bar"},
					map[string]interface{}{"a": 0},
					time.Now(),
				),
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "barB", "foo": "bar"},
					map[string]interface{}{"a": 0},
					time.Now(),
				),
			},
			Current: map[uint64]lldInfo{
				1: {
					Hostname: "barA",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						4916699111010086803: {
							"{#FOO}": "bar",
						},
					},
				},
				2: {
					Hostname: "barB",
					Key:      "disk.foo",
					Data: map[uint64]map[string]string{
						4917655686126441148: {
							"{#FOO}": "bar",
						},
					},
				},
			},
		},
		"different number of tags for the same measurement are stored in different keys": {
			Metrics: []Dana.Metric{
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "bar", "foo1": "bar", "foo2": "bar"},
					map[string]interface{}{"a": 0},
					time.Now(),
				),
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "bar", "foo1": "bar"},
					map[string]interface{}{"a": 0},
					time.Now(),
				),
			},
			Current: map[uint64]lldInfo{
				1: {
					Hostname: "bar",
					Key:      "disk.foo1.foo2",
					Data: map[uint64]map[string]string{
						12473238139685120014: {
							"{#FOO1}": "bar",
							"{#FOO2}": "bar",
						},
					},
				},
				2: {
					Hostname: "bar",
					Key:      "disk.foo1",
					Data: map[uint64]map[string]string{
						4193955122073793785: {
							"{#FOO1}": "bar",
						},
					},
				},
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			zl := zabbixLLD{
				log:           testutil.Logger{},
				clearInterval: config.Duration(time.Hour),
				lastClear:     time.Now(),
				hostTag:       "host",
				current:       make(map[uint64]lldInfo),
			}

			for _, m := range test.Metrics {
				require.NoError(t, zl.Add(m))
			}

			// Calculate series ID for the test data.
			// Metric hashes could not be calculated because we don't have enough information.
			for id, info := range test.Current {
				calculatedID := lldSeriesID(info.Hostname, info.Key)
				if id == calculatedID {
					continue
				}

				test.Current[calculatedID] = info

				// Drop old ID
				delete(test.Current, id)
			}

			require.Equal(t, test.Current, zl.current)
		})
	}
}
