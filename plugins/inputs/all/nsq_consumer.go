//go:build !custom || inputs || inputs.nsq_consumer

package all

import _ "Dana/plugins/inputs/nsq_consumer" // register plugin
