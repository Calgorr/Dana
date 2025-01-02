//go:build !custom || inputs || inputs.clickhouse

package all

import _ "Dana/plugins/inputs/clickhouse" // register plugin
