//go:build !custom || serializers || serializers.graphite

package all

import (
	_ "Dana/plugins/serializers/graphite" // register plugin
)
