package validators

import (
	"net/url"
	"regexp"
	"strings"
)

var universityIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)

var allowedUniversityOwnershipTypes = map[string]struct{}{
	"federal": {},
	"state":   {},
	"private": {},
}

// UniversityListQuery contains validated filters for university listings.
type UniversityListQuery struct {
	OwnershipType *string
	StateID       *string
}

// ValidateUniversityID validates the documented public university identifier.
func ValidateUniversityID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError("university_id", "University ID is required.")
	}
	if !universityIDPattern.MatchString(value) || uuidLikePattern.MatchString(value) {
		return "", invalidField("university_id", "University ID must be a valid lowercase public slug.")
	}
	return value, nil
}

// ValidateUniversityOwnershipType validates the documented public university ownership type.
func ValidateUniversityOwnershipType(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError("ownership_type", "Ownership type is required.")
	}
	if _, ok := allowedUniversityOwnershipTypes[value]; !ok {
		return "", invalidField("ownership_type", "Ownership type must be one of the supported university ownership categories.")
	}
	return value, nil
}

// ValidateUniversityStateID validates the documented public university state identifier.
func ValidateUniversityStateID(value string) (string, error) {
	return ValidateStateID(value)
}

// ValidateUniversityListQuery validates the documented university list query inputs.
func ValidateUniversityListQuery(values url.Values) (UniversityListQuery, error) {
	var errs ValidationErrors

	ownershipTypeValues := values["ownership_type"]
	stateIDValues := values["state_id"]

	if len(ownershipTypeValues) > 1 {
		errs.Add("ownership_type", codeMalformed, "Ownership type may be provided at most once.")
	}
	if len(stateIDValues) > 1 {
		errs.Add("state_id", codeMalformed, "State ID may be provided at most once.")
	}
	if len(errs.Fields) > 0 {
		return UniversityListQuery{}, errs
	}

	var query UniversityListQuery

	if len(ownershipTypeValues) == 1 {
		normalized, err := ValidateUniversityOwnershipType(ownershipTypeValues[0])
		if err != nil {
			if validationErr, ok := err.(ValidationErrors); ok {
				errs.Fields = append(errs.Fields, validationErr.Fields...)
			} else {
				return UniversityListQuery{}, err
			}
		} else {
			query.OwnershipType = &normalized
		}
	}

	if len(stateIDValues) == 1 {
		normalized, err := ValidateUniversityStateID(stateIDValues[0])
		if err != nil {
			if validationErr, ok := err.(ValidationErrors); ok {
				errs.Fields = append(errs.Fields, validationErr.Fields...)
			} else {
				return UniversityListQuery{}, err
			}
		} else {
			query.StateID = &normalized
		}
	}

	if len(errs.Fields) > 0 {
		return UniversityListQuery{}, errs
	}

	return query, nil
}
