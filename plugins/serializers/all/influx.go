//go:build !custom || serializers || serializers.influx

package all

import (
	_ "Dana/plugins/serializers/influx" // register plugin
)
