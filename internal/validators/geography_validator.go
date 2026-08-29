package validators

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var stateIDPattern = regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`)
var geopoliticalZoneIDPattern = regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`)
var localGovernmentUnitIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
var localGovernmentUnitSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var uuidLikePattern = regexp.MustCompile(`^[a-f0-9]{8}(?:-[a-f0-9]{4}){3}-[a-f0-9]{12}$`)

var approvedStateIDs = []string{
	"abia",
	"adamawa",
	"akwa-ibom",
	"anambra",
	"bauchi",
	"bayelsa",
	"benue",
	"borno",
	"cross-river",
	"delta",
	"ebonyi",
	"edo",
	"ekiti",
	"enugu",
	"fct",
	"gombe",
	"imo",
	"jigawa",
	"kaduna",
	"kano",
	"katsina",
	"kebbi",
	"kogi",
	"kwara",
	"lagos",
	"nasarawa",
	"niger",
	"ogun",
	"ondo",
	"osun",
	"oyo",
	"plateau",
	"rivers",
	"sokoto",
	"taraba",
	"yobe",
	"zamfara",
}

var validStateIDSet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(approvedStateIDs))
	for _, id := range approvedStateIDs {
		set[id] = struct{}{}
	}
	return set
}()

var validStateIDsByPrefix = func() []string {
	ids := append([]string(nil), approvedStateIDs...)
	sort.Slice(ids, func(i, j int) bool {
		if len(ids[i]) != len(ids[j]) {
			return len(ids[i]) > len(ids[j])
		}
		return ids[i] < ids[j]
	})
	return ids
}()
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
	if _, ok := validStateIDSet[value]; !ok {
		return "", invalidField("state_id", "State ID must reference a supported public state.")
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

// LocalGovernmentUnitListQuery contains validated filters for local-government-unit listings.
type LocalGovernmentUnitListQuery struct {
	StateID *string
}

// ValidateLocalGovernmentUnitListQuery validates the documented local-government-unit list query.
func ValidateLocalGovernmentUnitListQuery(values url.Values) (LocalGovernmentUnitListQuery, error) {
	var errs ValidationErrors

	stateValues := values["state_id"]
	if len(stateValues) > 1 {
		errs.Add("state_id", codeMalformed, "State ID may be provided at most once.")
		return LocalGovernmentUnitListQuery{}, errs
	}
	if len(stateValues) == 0 {
		return LocalGovernmentUnitListQuery{}, nil
	}

	normalized, err := ValidateStateID(stateValues[0])
	if err != nil {
		if validationErr, ok := err.(ValidationErrors); ok {
			errs.Fields = append(errs.Fields, validationErr.Fields...)
		} else {
			return LocalGovernmentUnitListQuery{}, err
		}
	}
	if len(errs.Fields) > 0 {
		return LocalGovernmentUnitListQuery{}, errs
	}

	query := LocalGovernmentUnitListQuery{StateID: &normalized}
	return query, nil
}

// ValidateLocalGovernmentUnitID validates the documented public local-government-unit identifier.
func ValidateLocalGovernmentUnitID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError("lga_id", "LGA ID is required.")
	}
	if !localGovernmentUnitIDPattern.MatchString(value) || uuidLikePattern.MatchString(value) {
		return "", invalidField("lga_id", "LGA ID must be a valid state-prefixed lowercase public slug.")
	}

	stateID, unitSlug, ok := splitLocalGovernmentUnitID(value)
	if !ok || unitSlug == "" {
		return "", invalidField("lga_id", "LGA ID must begin with a supported state ID followed by a unit slug.")
	}
	if _, ok := validStateIDSet[stateID]; !ok {
		return "", invalidField("lga_id", "LGA ID must begin with a supported state ID followed by a unit slug.")
	}
	if !localGovernmentUnitSlugPattern.MatchString(unitSlug) || uuidLikePattern.MatchString(unitSlug) {
		return "", invalidField("lga_id", "LGA ID must be a valid state-prefixed lowercase public slug.")
	}
	return value, nil
}

func splitLocalGovernmentUnitID(value string) (string, string, bool) {
	for _, stateID := range validStateIDsByPrefix {
		prefix := stateID + "-"
		if strings.HasPrefix(value, prefix) {
			return stateID, strings.TrimPrefix(value, prefix), true
		}
	}
	return "", "", false
}
