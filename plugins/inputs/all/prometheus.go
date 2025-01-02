//go:build !custom || inputs || inputs.prometheus

package all

import _ "Dana/plugins/inputs/prometheus" // register plugin
