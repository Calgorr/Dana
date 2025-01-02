//go:build !custom || serializers || serializers.json

package all

import (
	_ "Dana/plugins/serializers/json" // register plugin
)
