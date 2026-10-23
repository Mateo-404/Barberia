package auth

import "crypto/rand"

// cryptoRead existe para poder sustituirlo en tests; por defecto usa crypto/rand.
var cryptoRead = rand.Read
