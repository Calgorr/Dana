//go:build !custom || outputs || outputs.datadog

package all

import _ "Dana/plugins/outputs/datadog" // register plugin
