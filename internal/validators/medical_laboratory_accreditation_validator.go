package validators

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var medicalLaboratoryAccreditationSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidateMedicalLaboratoryAccreditationID validates a public accreditation slug.
func ValidateMedicalLaboratoryAccreditationID(field, value string) (string, error) {
	if value == "" {
		return "", requiredError(field, "Accreditation ID is required.")
	}
	if len(value) > models.MedicalLaboratoryAccreditationIDMaxLength || !medicalLaboratoryAccreditationSlugPattern.MatchString(value) {
		return "", invalidField(field, "Accreditation ID must be a valid lowercase public slug.")
	}
	return value, nil
}

// ValidateMedicalLaboratoryAccreditationListQuery validates healthcare pagination and filters.
func ValidateMedicalLaboratoryAccreditationListQuery(values url.Values) (interfaces.MedicalLaboratoryAccreditationQuery, error) {
	var errs ValidationErrors
	query := interfaces.MedicalLaboratoryAccreditationQuery{Page: 1, PageSize: 50}

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

	validateSlug := func(field string) string {
		raw, ok := readSingle(field)
		if !ok {
			return ""
		}
		value := strings.TrimSpace(raw)
		if value == "" {
			return ""
		}
		if !medicalLaboratoryAccreditationSlugPattern.MatchString(value) {
			errs.Add(field, codeInvalid, "Value must be a valid lowercase public slug.")
			return ""
		}
		return value
	}
	query.StateID = validateSlug("state_id")

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
	query.AccreditationStatus = validateEnum("accreditation_status", map[string]struct{}{"accredited": {}, "expired": {}})
	if raw, ok := readSingle("search"); ok {
		query.Search = strings.TrimSpace(raw)
		if len([]rune(query.Search)) > 100 || strings.ContainsAny(query.Search, "\r\n\x00") {
			errs.Add("search", codeInvalid, "Search must be at most 100 characters.")
			query.Search = ""
		}
	}

	if len(errs.Fields) > 0 {
		return interfaces.MedicalLaboratoryAccreditationQuery{}, errs
	}
	return query, nil
}
