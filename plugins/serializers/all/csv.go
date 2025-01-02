//go:build !custom || serializers || serializers.csv

package all

import (
	_ "Dana/plugins/serializers/csv" // register plugin
)
