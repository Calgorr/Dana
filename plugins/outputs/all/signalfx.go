//go:build !custom || outputs || outputs.signalfx

package all

import _ "Dana/plugins/outputs/signalfx" // register plugin
