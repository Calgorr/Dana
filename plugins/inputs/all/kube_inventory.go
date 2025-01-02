//go:build !custom || inputs || inputs.kube_inventory

package all

import _ "Dana/plugins/inputs/kube_inventory" // register plugin
