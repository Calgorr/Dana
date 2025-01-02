//go:build !custom || inputs || inputs.fluentd

package all

import _ "Dana/plugins/inputs/fluentd" // register plugin
