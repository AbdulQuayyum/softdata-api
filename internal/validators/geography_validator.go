package validators

import (
	"regexp"
	"strings"
)

var stateIDPattern = regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`)
var geopoliticalZoneIDPattern = regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`)
var validGeopoliticalZoneIDs = map[string]struct{}{
	"north-central": {},
	"north-east":    {},
	"north-west":    {},
	"south-east":    {},
	"south-south":   {},
	"south-west":    {},
}

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

// ValidateGeopoliticalZoneID validates the documented public geopolitical-zone identifier.
func ValidateGeopoliticalZoneID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError("zone_id", "Zone ID is required.")
	}
	if !geopoliticalZoneIDPattern.MatchString(value) {
		return "", invalidField("zone_id", "Zone ID must be a valid lowercase public slug.")
	}
	if _, ok := validGeopoliticalZoneIDs[value]; !ok {
		return "", invalidField("zone_id", "Zone ID must be one of the supported geopolitical zones.")
	}
	return value, nil
}
