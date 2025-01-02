//go:build !custom || serializers || serializers.binary

package all

import (
	_ "Dana/plugins/serializers/binary" // register plugin
)
