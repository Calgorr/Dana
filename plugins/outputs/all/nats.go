//go:build !custom || outputs || outputs.nats

package all

import _ "Dana/plugins/outputs/nats" // register plugin
