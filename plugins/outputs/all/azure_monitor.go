//go:build !custom || outputs || outputs.azure_monitor

package all

import _ "Dana/plugins/outputs/azure_monitor" // register plugin
