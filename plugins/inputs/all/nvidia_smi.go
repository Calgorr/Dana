//go:build !custom || inputs || inputs.nvidia_smi

package all

import _ "Dana/plugins/inputs/nvidia_smi" // register plugin
