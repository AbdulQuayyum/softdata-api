package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func secureRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("random length must be positive")
	}

	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("secure random generation failed: %w", err)
	}

	return buf, nil
}

func secureRandomString(length int) (string, error) {
	buf, err := secureRandomBytes(length)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func newTokenID() (string, error) {
	return secureRandomString(16)
}

func GenerateRefreshToken() (string, string, error) {
	plaintext, err := secureRandomString(32)
	if err != nil {
		return "", "", err
	}

	return plaintext, HashRefreshToken(plaintext), nil
}

func HashRefreshToken(refreshToken string) string {
	sum := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(sum[:])
}
