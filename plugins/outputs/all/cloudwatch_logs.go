//go:build !custom || outputs || outputs.cloudwatch_logs

package all

import _ "Dana/plugins/outputs/cloudwatch_logs" // register plugin
