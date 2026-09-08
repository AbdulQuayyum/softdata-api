package validators

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var healthFacilitySlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var healthFacilityTypeValues = map[string]struct{}{
	"primary-health-centre": {}, "clinic": {}, "health-post": {}, "general-hospital": {},
	"other": {}, "teaching-hospital": {}, "specialist-hospital": {},
}

var healthFacilityLevelValues = map[string]struct{}{"primary": {}, "secondary": {}, "tertiary": {}}

var healthFacilityOwnershipValues = map[string]struct{}{
	"federal": {}, "state": {}, "local-government": {}, "private": {}, "military": {}, "other-public": {},
}

// ValidateHealthFacilityID validates a public health-facility slug.
func ValidateHealthFacilityID(field, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError(field, "Facility ID is required.")
	}
	if !healthFacilitySlugPattern.MatchString(value) {
		return "", invalidField(field, "Facility ID must be a valid lowercase public slug.")
	}
	return value, nil
}

// ValidateHealthFacilityListQuery validates healthcare pagination and filters.
func ValidateHealthFacilityListQuery(values url.Values) (interfaces.HealthFacilityQuery, error) {
	var errs ValidationErrors
	query := interfaces.HealthFacilityQuery{Page: 1, PageSize: 50}

	readSingle := func(field string) (string, bool) {
		items := values[field]
		if len(items) > 1 {
			errs.Add(field, codeMalformed, "Field may be provided at most once.")
			return "", false
		}
		if len(items) == 0 {
			return "", false
		}
		return items[0], true
	}
	parsePositive := func(field, raw string, defaultValue, maximum int) int {
		if raw == "" {
			return defaultValue
		}
		value, err := strconv.ParseUint(raw, 10, 0)
		if err != nil || value > uint64(^uint(0)>>1) || value < 1 || (maximum > 0 && int(value) > maximum) {
			errs.Add(field, codeInvalid, "Value must be a valid positive integer within the supported range.")
			return defaultValue
		}
		return int(value)
	}
	if raw, ok := readSingle("page"); ok {
		query.Page = parsePositive("page", raw, 1, 0)
	}
	if raw, ok := readSingle("page_size"); ok {
		query.PageSize = parsePositive("page_size", raw, 50, 100)
	}

	validateSlug := func(field string) string {
		raw, ok := readSingle(field)
		if !ok {
			return ""
		}
		value := strings.TrimSpace(raw)
		if value == "" {
			return ""
		}
		if !healthFacilitySlugPattern.MatchString(value) {
			errs.Add(field, codeInvalid, "Value must be a valid lowercase public slug.")
			return ""
		}
		return value
	}
	query.StateID = validateSlug("state_id")
	query.LGAID = validateSlug("lga_id")

	validateEnum := func(field string, values map[string]struct{}) string {
		raw, ok := readSingle(field)
		if !ok {
			return ""
		}
		value := strings.TrimSpace(raw)
		if value == "" {
			return ""
		}
		if _, valid := values[value]; !valid {
			errs.Add(field, codeInvalid, "Value is not supported.")
			return ""
		}
		return value
	}
	query.FacilityType = validateEnum("facility_type", healthFacilityTypeValues)
	query.FacilityLevel = validateEnum("facility_level", healthFacilityLevelValues)
	query.OwnershipType = validateEnum("ownership_type", healthFacilityOwnershipValues)
	if raw, ok := readSingle("search"); ok {
		query.Search = strings.TrimSpace(raw)
		if len([]rune(query.Search)) > 100 || strings.ContainsAny(query.Search, "\r\n\x00") {
			errs.Add("search", codeInvalid, "Search must be at most 100 characters.")
			query.Search = ""
		}
	}

	if len(errs.Fields) > 0 {
		return interfaces.HealthFacilityQuery{}, errs
	}
	return query, nil
}
