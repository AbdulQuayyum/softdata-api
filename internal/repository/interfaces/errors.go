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
	// ErrInvalidDatasetPath reports an unsafe or unsupported dataset path.
	ErrInvalidDatasetPath = errors.New("repository: invalid dataset path")
	// ErrDatasetFileNotFound reports that the requested dataset file is missing.
	ErrDatasetFileNotFound = errors.New("repository: dataset file not found")
	// ErrDatasetFileTooLarge reports that the dataset file exceeds the configured limit.
	ErrDatasetFileTooLarge = errors.New("repository: dataset file too large")
	// ErrInvalidDatasetFile reports that the dataset file could not be decoded or validated.
	ErrInvalidDatasetFile = errors.New("repository: invalid dataset file")
	// ErrDatasetFileUnavailable reports that the dataset file cannot be accessed right now.
	ErrDatasetFileUnavailable = errors.New("repository: dataset file unavailable")
	// ErrStateNotFound reports that a requested state is not present in the dataset.
	ErrStateNotFound = errors.New("repository: state not found")
	// ErrGeopoliticalZoneNotFound reports that a requested geopolitical zone is not present in the dataset.
	ErrGeopoliticalZoneNotFound = errors.New("repository: geopolitical zone not found")
	// ErrLocalGovernmentUnitNotFound reports that a requested local-government unit is not present in the dataset.
	ErrLocalGovernmentUnitNotFound = errors.New("repository: local government unit not found")
)
