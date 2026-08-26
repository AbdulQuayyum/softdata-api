package security

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	apiKeyPrefix        = "sd_live_"
	apiKeySuffixByteLen = 32
)

func GenerateAPIKey() (string, string, string, string, error) {
	suffix, err := secureRandomString(apiKeySuffixByteLen)
	if err != nil {
		return "", "", "", "", err
	}

	plaintext := apiKeyPrefix + suffix
	hash := HashAPIKey(plaintext)
	keyPrefix, keyLast4 := APIKeyDisplayParts(plaintext)
	return plaintext, hash, keyPrefix, keyLast4, nil
}

func HashAPIKey(apiKey string) string {
	sum := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(sum[:])
}

func APIKeyDisplayParts(apiKey string) (string, string) {
	if !strings.HasPrefix(apiKey, apiKeyPrefix) {
		return "", ""
	}

	last4 := ""
	if len(apiKey) >= 4 {
		last4 = apiKey[len(apiKey)-4:]
	}

	return apiKeyPrefix, last4
}

func ValidateAPIKeyFormat(apiKey string) error {
	if strings.TrimSpace(apiKey) == "" {
		return errors.New("api key is required")
	}
	if !strings.HasPrefix(apiKey, apiKeyPrefix) {
		return fmt.Errorf("api key must use %s prefix", apiKeyPrefix)
	}
	return nil
}

func DecodeAPIKeySuffix(apiKey string) (string, error) {
	if err := ValidateAPIKeyFormat(apiKey); err != nil {
		return "", err
	}

	suffix := strings.TrimPrefix(apiKey, apiKeyPrefix)
	if _, err := base64.RawURLEncoding.DecodeString(suffix); err != nil {
		return "", errors.New("api key suffix is invalid")
	}
	return suffix, nil
}
