//go:build !custom || parsers || parsers.prometheus

package all

import _ "Dana/plugins/parsers/prometheus" // register plugin
