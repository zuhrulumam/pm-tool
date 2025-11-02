package crypto

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
)

// GenerateRandomToken generates a random token of specified length
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateUUID generates a new UUID v4
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateAPIKey generates a random API key (32 bytes = 64 hex chars)
func GenerateAPIKey() (string, error) {
	return GenerateRandomToken(32)
}

// GenerateVerificationCode generates a numeric verification code
func GenerateVerificationCode(length int) (string, error) {
	const digits = "0123456789"
	bytes := make([]byte, length)
	
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = digits[b%byte(len(digits))]
	}

	return string(bytes), nil
}
