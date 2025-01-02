//go:build !custom || inputs || inputs.kafka_consumer

package all

import _ "Dana/plugins/inputs/kafka_consumer" // register plugin
