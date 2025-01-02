//go:build !custom || migrations

package all

import _ "Dana/migrations/general_metricfilter" // register migration
