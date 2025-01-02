//go:build !custom || inputs || inputs.nomad

package all

import _ "Dana/plugins/inputs/nomad" // register plugin
