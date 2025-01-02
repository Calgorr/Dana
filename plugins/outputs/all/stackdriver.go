//go:build !custom || outputs || outputs.stackdriver

package all

import _ "Dana/plugins/outputs/stackdriver" // register plugin
