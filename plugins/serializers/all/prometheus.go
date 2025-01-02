//go:build !custom || serializers || serializers.prometheus

package all

import (
	_ "Dana/plugins/serializers/prometheus" // register plugin
)
