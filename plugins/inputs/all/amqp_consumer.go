//go:build !custom || inputs || inputs.amqp_consumer

package all

import _ "Dana/plugins/inputs/amqp_consumer" // register plugin
