package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateRandomString generates a slow but true-random string from a given characterset of a given length.
// example charset="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
func GenerateRandomString(length int, charset string) string {
	// Create a byte slice of the required length
	randomBytes := make([]byte, length)

	// Generate random numbers within the range of the charset length
	charsetLength := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		randomIndex, _ := rand.Int(rand.Reader, charsetLength)
		randomBytes[i] = charset[randomIndex.Int64()]
	}

	// Convert byte slice to string
	return string(randomBytes)
}
