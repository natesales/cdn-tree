package authentication

import (
	"crypto/rand"
	"crypto/subtle"
	"golang.org/x/crypto/argon2"
)

// getIDKey computes an argon2 hash by given input and salt
func getIDKey(input []byte, salt []byte) []byte {
	return argon2.IDKey(input, salt, 1, 64*1024, 4, 32)
}

// GetPasswordHash hashes a plaintext string and returns an argon2 hash byte slice
func GetPasswordHash(plaintext string) ([]byte, error) {
	// Generate a random 16-byte salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return make([]byte, 0), err
	}

	hash := getIDKey([]byte(plaintext), salt)

	return append(salt, hash...), nil
}

// ValidHash validates a hash and provided plaintext password
func ValidHash(payload []byte, plaintext string) bool {
	salt := payload[:16]
	hash := payload[16:]

	providedHash := getIDKey([]byte(plaintext), salt)
	return subtle.ConstantTimeCompare(hash, providedHash) == 1
}
