package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"
)

func DeriveAnonymousID(secret, normalizedIP, normalizedUserAgent string, at time.Time) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("anonymous id secret is required")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write(buildAnonymousInput(at, normalizedIP, normalizedUserAgent)); err != nil {
		return "", fmt.Errorf("derive anonymous id: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func buildAnonymousInput(at time.Time, normalizedIP, normalizedUserAgent string) []byte {
	var buf bytes.Buffer
	writeLengthPrefixedField(&buf, "softdata-anonymous-id")
	writeLengthPrefixedField(&buf, at.UTC().Format("2006-01-02"))
	writeLengthPrefixedField(&buf, normalizedIP)
	writeLengthPrefixedField(&buf, normalizedUserAgent)
	return buf.Bytes()
}

func writeLengthPrefixedField(buf *bytes.Buffer, value string) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	_, _ = buf.Write(length[:])
	_, _ = buf.WriteString(value)
}
