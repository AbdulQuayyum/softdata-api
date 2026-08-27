package interfaces

import "errors"

var (
	// ErrNotFound reports a missing persistence record.
	ErrNotFound = errors.New("repository: not found")
	// ErrConflict reports a portable uniqueness conflict.
	ErrConflict = errors.New("repository: conflict")
	// ErrInvalidRateLimitInput reports an invalid rate-limit request.
	ErrInvalidRateLimitInput = errors.New("repository: invalid rate limit input")
	// ErrRateLimitUnavailable reports that Redis-backed rate limiting is unavailable.
	ErrRateLimitUnavailable = errors.New("repository: rate limit unavailable")
)
