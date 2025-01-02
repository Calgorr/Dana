//go:build !custom || inputs || inputs.kinesis_consumer

package all

import _ "Dana/plugins/inputs/kinesis_consumer" // register plugin
