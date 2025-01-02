//go:build !custom || parsers || parsers.openmetrics

package all

import _ "Dana/plugins/parsers/openmetrics" // register plugin
