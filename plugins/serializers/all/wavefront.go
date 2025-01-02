//go:build !custom || serializers || serializers.wavefront

package all

import (
	_ "Dana/plugins/serializers/wavefront" // register plugin
)
