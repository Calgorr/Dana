//go:build !custom || serializers || serializers.prometheusremotewrite

package all

import (
	_ "Dana/plugins/serializers/prometheusremotewrite" // register plugin
)
