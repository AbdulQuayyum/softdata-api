package services

import "errors"

var (
	ErrInvalidCredentials     = errors.New("services: invalid credentials")
	ErrInvalidRefreshToken    = errors.New("services: invalid refresh token")
	ErrDatasetNotFound        = errors.New("services: dataset not found")
	ErrInvalidDatasetKey      = errors.New("services: invalid dataset key")
	ErrInvalidDatasetGroup    = errors.New("services: invalid dataset group")
	ErrInvalidPagination      = errors.New("services: invalid pagination")
	ErrUsernameUnavailable    = errors.New("services: username unavailable")
	ErrEmailUnavailable       = errors.New("services: email unavailable")
	ErrAccountNotFound        = errors.New("services: account not found")
	ErrCurrentPasswordInvalid = errors.New("services: current password invalid")
	ErrAccountInactive        = errors.New("services: account inactive")
	ErrAPIKeyNotFound         = errors.New("services: api key not found")
	ErrAPIKeyNameRequired     = errors.New("services: api key name is required")
	ErrAPIKeyLimitReached     = errors.New("services: api key limit reached")
	ErrInvalidUsagePeriod     = errors.New("services: invalid usage period")
)
