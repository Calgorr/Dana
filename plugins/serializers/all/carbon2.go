//go:build !custom || serializers || serializers.carbon2

package all

import (
	_ "Dana/plugins/serializers/carbon2" // register plugin
)
