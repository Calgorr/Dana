//go:build !custom || secretstores || secretstores.jose

package all

import _ "Dana/plugins/secretstores/jose" // register plugin
