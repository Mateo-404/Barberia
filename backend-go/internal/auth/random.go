package auth

import "crypto/rand"

// cryptoRead can be replaced in tests; it defaults to crypto/rand.
var cryptoRead = rand.Read
