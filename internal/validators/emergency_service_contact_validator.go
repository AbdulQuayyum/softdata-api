package validators

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var (
	emergencyServiceContactSlugPattern      = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	emergencyServiceContactShortCodePattern = regexp.MustCompile(`^[0-9]{3}$`)
	emergencyServiceContactTelephonePattern = regexp.MustCompile(`^(?:0[0-9]{10}|0800[0-9]{8}|\+234[0-9]{10})$`)
)

// ValidateEmergencyServiceContactID validates a public emergency-contact slug.
func ValidateEmergencyServiceContactID(field, value string) (string, error) {
	if value == "" {
		return "", requiredError(field, "Contact ID is required.")
	}
	if strings.TrimSpace(value) != value || strings.Contains(value, "/") || strings.Contains(value, "\\") {
		return "", invalidField(field, "Contact ID must be a valid lowercase public slug.")
	}
	if len(value) > models.EmergencyServiceContactIDMaxLength || !emergencyServiceContactSlugPattern.MatchString(value) {
		return "", invalidField(field, "Contact ID must be a valid lowercase public slug.")
	}
	return value, nil
}

// ValidateEmergencyServiceContactListQuery validates emergency-contact pagination and filters.
func ValidateEmergencyServiceContactListQuery(values url.Values) (interfaces.EmergencyServiceContactQuery, error) {
	var errs ValidationErrors
	query := interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50}
	allowed := map[string]struct{}{
		"page": {}, "page_size": {}, "service_type": {}, "contact_type": {}, "coverage_type": {}, "contact_value": {}, "search": {},
	}
	for field := range values {
		if _, ok := allowed[field]; !ok {
			errs.Add(field, codeInvalid, "Query parameter is not supported.")
		}
	}
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
	if raw, ok := readSingle("service_type"); ok {
		query.ServiceType = strings.TrimSpace(raw)
		if query.ServiceType == "" {
			errs.Add("service_type", codeRequired, "Service type is required when provided.")
		} else if !validEmergencyServiceContactServiceType(query.ServiceType) {
			errs.Add("service_type", codeInvalid, "Value is not supported.")
		}
	}
	if raw, ok := readSingle("contact_type"); ok {
		query.ContactType = strings.TrimSpace(raw)
		if query.ContactType == "" {
			errs.Add("contact_type", codeRequired, "Contact type is required when provided.")
		} else if query.ContactType != "short_code" && query.ContactType != "telephone" {
			errs.Add("contact_type", codeInvalid, "Value is not supported.")
		}
	}
	if raw, ok := readSingle("coverage_type"); ok {
		query.CoverageType = strings.TrimSpace(raw)
		if query.CoverageType == "" {
			errs.Add("coverage_type", codeRequired, "Coverage type is required when provided.")
		} else if query.CoverageType != "national" && query.CoverageType != "state" {
			errs.Add("coverage_type", codeInvalid, "Value is not supported.")
		}
	}
	if raw, ok := readSingle("contact_value"); ok {
		query.ContactValue = strings.TrimSpace(raw)
		if query.ContactValue == "" {
			errs.Add("contact_value", codeRequired, "Contact value is required when provided.")
		} else if !(emergencyServiceContactShortCodePattern.MatchString(query.ContactValue) || emergencyServiceContactTelephonePattern.MatchString(query.ContactValue)) {
			errs.Add("contact_value", codeInvalid, "Contact value must be a supported short code or telephone string.")
		}
	}
	if raw, ok := readSingle("search"); ok {
		query.Search = strings.TrimSpace(raw)
		if query.Search == "" {
			query.Search = ""
		} else if len([]rune(query.Search)) > 100 || strings.ContainsAny(query.Search, "\r\n\x00") {
			errs.Add("search", codeInvalid, "Search must be at most 100 characters.")
			query.Search = ""
		}
	}
	if len(errs.Fields) > 0 {
		return interfaces.EmergencyServiceContactQuery{}, errs
	}
	return query, nil
}

func validEmergencyServiceContactServiceType(value string) bool {
	switch value {
	case "general_emergency", "disaster_management", "road_emergency", "police", "fire", "ambulance", "other":
		return true
	default:
		return false
	}
}
