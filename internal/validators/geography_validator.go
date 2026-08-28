package validators

import (
	"regexp"
	"strings"
)

var stateIDPattern = regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`)

// ValidateStateID validates the documented public state identifier.
func ValidateStateID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError("state_id", "State ID is required.")
	}
	if !stateIDPattern.MatchString(value) {
		return "", invalidField("state_id", "State ID must be a valid lowercase public slug.")
	}
	return value, nil
}
