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
	// ErrLanguageNotFound reports that a requested language is not present in the dataset.
	ErrLanguageNotFound = errors.New("repository: language not found")
	// ErrInvalidCountryLanguageCountryAreaID reports an invalid country/area filter.
	ErrInvalidCountryLanguageCountryAreaID = errors.New("repository: invalid country language country area id")
	// ErrInvalidCountryLanguageLanguageID reports an invalid language filter.
	ErrInvalidCountryLanguageLanguageID = errors.New("repository: invalid country language language id")
	// ErrInvalidCountryLanguageStatus reports an invalid relationship status filter.
	ErrInvalidCountryLanguageStatus = errors.New("repository: invalid country language status")
	// ErrTimeZoneNotFound reports that a requested time zone is not present in the dataset.
	ErrTimeZoneNotFound = errors.New("repository: time zone not found")
	// ErrCountryOrAreaNotFound reports that a requested country or area is not present in the dataset.
	ErrCountryOrAreaNotFound = errors.New("repository: country or area not found")
	// ErrUniversityNotFound reports that a requested university is not present in the dataset.
	ErrUniversityNotFound = errors.New("repository: university not found")
	// ErrCollegeOfEducationNotFound reports that a requested college of education is not present in the dataset.
	ErrCollegeOfEducationNotFound = errors.New("repository: college of education not found")
	// ErrPaymentServiceProviderNotFound reports that a requested payment service provider is not present in the dataset.
	ErrPaymentServiceProviderNotFound = errors.New("repository: payment service provider not found")
	// ErrInternationalMoneyTransferOperatorNotFound reports that a requested IMTO is not present in the dataset.
	ErrInternationalMoneyTransferOperatorNotFound = errors.New("repository: international money transfer operator not found")
	// ErrCurrencyNotFound reports that a requested currency is not present in the dataset.
	ErrCurrencyNotFound = errors.New("repository: currency not found")
	// ErrInvalidCurrencyCountryAreaID reports that a currency country/area filter is invalid or unknown.
	ErrInvalidCurrencyCountryAreaID = errors.New("repository: invalid currency country area id")
)
