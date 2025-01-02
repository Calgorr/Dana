//go:build !custom || serializers || serializers.nowmetric

package all

import (
	_ "Dana/plugins/serializers/nowmetric" // register plugin
)
