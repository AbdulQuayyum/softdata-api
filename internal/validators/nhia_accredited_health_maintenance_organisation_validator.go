package validators

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const nhiaAccreditedHMOMaxHMOIDLength = 64

var nhiaAccreditedHMOSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidateNHIAAccreditedHealthMaintenanceOrganisationID validates a public NHIA HMO slug.
func ValidateNHIAAccreditedHealthMaintenanceOrganisationID(field, value string) (string, error) {
	if value == "" {
		return "", requiredError(field, "Organisation ID is required.")
	}
	if strings.TrimSpace(value) != value || strings.Contains(value, "/") || strings.Contains(value, "\\") {
		return "", invalidField(field, "Organisation ID must be a valid lowercase public slug.")
	}
	if len(value) > models.NHIAAccreditedHealthMaintenanceOrganisationIDMaxLength || !nhiaAccreditedHMOSlugPattern.MatchString(value) {
		return "", invalidField(field, "Organisation ID must be a valid lowercase public slug.")
	}
	return value, nil
}

// ValidateNHIAAccreditedHealthMaintenanceOrganisationListQuery validates NHIA HMO pagination and filters.
func ValidateNHIAAccreditedHealthMaintenanceOrganisationListQuery(values url.Values) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery, error) {
	var errs ValidationErrors
	query := interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: 1, PageSize: 50}
	allowed := map[string]struct{}{
		"page": {}, "page_size": {}, "accreditation_status": {}, "hmo_id": {}, "search": {},
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
	if raw, ok := readSingle("accreditation_status"); ok {
		query.AccreditationStatus = strings.TrimSpace(raw)
		if query.AccreditationStatus == "" {
			errs.Add("accreditation_status", codeRequired, "Accreditation status is required when provided.")
		} else if query.AccreditationStatus != "accredited" {
			errs.Add("accreditation_status", codeInvalid, "Value is not supported.")
		}
	}
	if raw, ok := readSingle("hmo_id"); ok {
		query.HMOID = strings.TrimSpace(raw)
		if query.HMOID == "" {
			errs.Add("hmo_id", codeRequired, "HMO ID is required when provided.")
		} else if len([]rune(query.HMOID)) > nhiaAccreditedHMOMaxHMOIDLength || strings.ContainsAny(query.HMOID, "\r\n\x00") {
			errs.Add("hmo_id", codeInvalid, "HMO ID must be a bounded string.")
		}
	}
	if raw, ok := readSingle("search"); ok {
		query.Search = strings.TrimSpace(raw)
		if query.Search == "" {
			errs.Add("search", codeRequired, "Search is required when provided.")
		} else if len([]rune(query.Search)) > 100 || strings.ContainsAny(query.Search, "\r\n\x00") {
			errs.Add("search", codeInvalid, "Search must be at most 100 characters.")
			query.Search = ""
		}
	}

	if len(errs.Fields) > 0 {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}, errs
	}
	return query, nil
}
