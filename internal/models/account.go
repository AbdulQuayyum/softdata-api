package models

import "time"

type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusDeleted   AccountStatus = "deleted"
)

// Account is the internal persistence model for a developer account.
type Account struct {
	ID              string `json:"-"`
	Username        string
	Email           *string
	PasswordHash    string `json:"-"`
	Status          AccountStatus
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// AccountCreateInput models the registration payload supported by the docs.
type AccountCreateInput struct {
	Username string  `json:"username"`
	Email    *string `json:"email,omitempty"`
	Password string  `json:"password"`
}

// AccountUpdateInput models the current account update payload.
type AccountUpdateInput struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
}

// AccountResponse is the safe public account representation.
type AccountResponse struct {
	ID              string        `json:"id"`
	Username        string        `json:"username"`
	Email           *string       `json:"email,omitempty"`
	Status          AccountStatus `json:"status"`
	EmailVerifiedAt *time.Time    `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time    `json:"last_login_at,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}
