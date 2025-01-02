//go:build !custom || outputs || outputs.nsq

package all

import _ "Dana/plugins/outputs/nsq" // register plugin
