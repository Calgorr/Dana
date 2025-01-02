//go:build !custom || secretstores || secretstores.os

package all

import _ "Dana/plugins/secretstores/os" // register plugin
