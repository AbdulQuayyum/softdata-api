package interfaces

import "errors"

var (
	// ErrNotFound reports a missing persistence record.
	ErrNotFound = errors.New("repository: not found")
	// ErrConflict reports a portable uniqueness conflict.
	ErrConflict = errors.New("repository: conflict")
)
