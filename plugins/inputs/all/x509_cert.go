//go:build !custom || inputs || inputs.x509_cert

package all

import _ "Dana/plugins/inputs/x509_cert" // register plugin
