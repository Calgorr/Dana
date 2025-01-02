//go:build !custom || outputs || outputs.loki

package all

import _ "Dana/plugins/outputs/loki" // register plugin
