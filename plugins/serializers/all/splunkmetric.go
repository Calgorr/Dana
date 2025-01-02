//go:build !custom || serializers || serializers.splunkmetric

package all

import (
	_ "Dana/plugins/serializers/splunkmetric" // register plugin
)
