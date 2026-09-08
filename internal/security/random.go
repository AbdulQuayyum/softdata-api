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
	buf, err := secureRandomBytes(16)
	if err != nil {
		return "", err
	}
	// sessions.access_token_jti is a PostgreSQL UUID. Generate a random
	// version-4 UUID so login and refresh can persist the signed token's ID.
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[:4], buf[4:6], buf[6:8], buf[8:10], buf[10:]), nil
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
