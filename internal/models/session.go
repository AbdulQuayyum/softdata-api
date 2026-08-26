package models

import "time"

// Session is the internal persistence model for a refresh-token session.
type Session struct {
	ID               string  `json:"-"`
	AccountID        string  `json:"-"`
	RefreshTokenHash string  `json:"-"`
	AccessTokenJTI   *string `json:"-"`
	UserAgent        *string
	IPAddress        *string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	LastUsedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
