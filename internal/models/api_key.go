package models

import "time"

type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
	APIKeyStatusExpired APIKeyStatus = "expired"
)

// APIKey is the internal persistence model for stored API keys.
type APIKey struct {
	ID         string `json:"-"`
	AccountID  string `json:"-"`
	Name       string
	KeyPrefix  string
	KeyHash    string `json:"-"`
	KeyLast4   string
	Status     APIKeyStatus
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// APIKeyCreateInput models the documented API-key creation payload.
type APIKeyCreateInput struct {
	Name string `json:"name"`
}

// APIKeyMetadata is the public API-key representation without the stored hash.
type APIKeyMetadata struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	KeyPrefix  string       `json:"key_prefix"`
	KeyLast4   string       `json:"key_last4"`
	Status     APIKeyStatus `json:"status"`
	LastUsedAt *time.Time   `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time   `json:"expires_at,omitempty"`
	RevokedAt  *time.Time   `json:"revoked_at,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// APIKeyCreatedResponse separates the one-time plaintext key from stored metadata.
type APIKeyCreatedResponse struct {
	Key    string         `json:"key"`
	APIKey APIKeyMetadata `json:"api_key"`
}
