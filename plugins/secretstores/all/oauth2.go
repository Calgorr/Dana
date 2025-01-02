//go:build !custom || secretstores || secretstores.oauth2

package all

import _ "Dana/plugins/secretstores/oauth2" // register plugin
