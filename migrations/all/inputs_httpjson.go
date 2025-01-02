//go:build !custom || (migrations && (inputs || inputs.httpjson))

package all

import _ "Dana/migrations/inputs_httpjson" // register migration
