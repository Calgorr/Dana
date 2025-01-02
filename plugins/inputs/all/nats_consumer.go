//go:build !custom || inputs || inputs.nats_consumer

package all

import _ "Dana/plugins/inputs/nats_consumer" // register plugin
