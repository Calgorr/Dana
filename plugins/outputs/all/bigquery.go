//go:build !custom || outputs || outputs.bigquery

package all

import _ "Dana/plugins/outputs/bigquery" // register plugin
