package services

import "errors"

var (
	ErrInvalidCredentials     = errors.New("services: invalid credentials")
	ErrInvalidRefreshToken    = errors.New("services: invalid refresh token")
	ErrUsernameUnavailable    = errors.New("services: username unavailable")
	ErrEmailUnavailable       = errors.New("services: email unavailable")
	ErrAccountNotFound        = errors.New("services: account not found")
	ErrCurrentPasswordInvalid = errors.New("services: current password invalid")
	ErrAccountInactive        = errors.New("services: account inactive")
)
