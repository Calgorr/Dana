//go:build !custom || inputs || inputs.cloudwatch_metric_streams

package all

import _ "Dana/plugins/inputs/cloudwatch_metric_streams" // register plugin
