package secretstores

import (
	"Dana"
)

// Creator is the function to create a new parser
type Creator func(id string) Dana.SecretStore

// SecretStores contains the registry of all known secret-stores
var SecretStores = make(map[string]Creator)

// Add adds a secret-store to the registry. Usually this function is called in the plugin's init function
func Add(name string, creator Creator) {
	SecretStores[name] = creator
}
