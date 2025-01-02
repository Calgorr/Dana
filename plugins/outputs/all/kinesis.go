//go:build !custom || outputs || outputs.kinesis

package all

import _ "Dana/plugins/outputs/kinesis" // register plugin
