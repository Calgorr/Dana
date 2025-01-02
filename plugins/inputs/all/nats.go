//go:build !custom || inputs || inputs.nats

package all

import _ "Dana/plugins/inputs/nats" // register plugin
