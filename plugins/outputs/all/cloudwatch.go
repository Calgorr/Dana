//go:build !custom || outputs || outputs.cloudwatch

package all

import _ "Dana/plugins/outputs/cloudwatch" // register plugin
