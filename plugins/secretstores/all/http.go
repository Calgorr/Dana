//go:build !custom || secretstores || secretstores.http

package all

import _ "Dana/plugins/secretstores/http" // register plugin
