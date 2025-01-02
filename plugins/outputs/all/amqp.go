//go:build !custom || outputs || outputs.amqp

package all

import _ "Dana/plugins/outputs/amqp" // register plugin
