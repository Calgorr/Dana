//go:build !custom || inputs || inputs.kubernetes

package all

import _ "Dana/plugins/inputs/kubernetes" // register plugin
