//go:build !custom || outputs || outputs.syslog

package all

import _ "Dana/plugins/outputs/syslog" // register plugin
