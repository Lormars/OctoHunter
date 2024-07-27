package common

import (
	"fmt"
	"hash/fnv"
)

func Hash(key string) string {

	// Create a new FNV-1a hash
	hash := fnv.New64a()

	// Write the key to the hash
	hash.Write([]byte(key))

	// Get the hash as a byte slice
	hashBytes := hash.Sum(nil)

	// Convert the hash to a hexadecimal string
	hashString := fmt.Sprintf("%x", hashBytes)

	return hashString

}
