//go:build !custom || secretstores || secretstores.docker

package all

import _ "Dana/plugins/secretstores/docker" // register plugin
