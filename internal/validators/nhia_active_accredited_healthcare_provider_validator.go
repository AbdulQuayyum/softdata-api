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
	nhiaActiveAccreditedHealthcareProviderSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	nhiaActiveAccreditedHealthcareProviderCodePattern = regexp.MustCompile(`^(?:[A-Z]{2,3})/[0-9]{4}/P$`)
)

// ValidateNHIAActiveAccreditedHealthcareProviderID validates a public NHIA HCP slug.
func ValidateNHIAActiveAccreditedHealthcareProviderID(field, value string) (string, error) {
	if value == "" {
		return "", requiredError(field, "Provider ID is required.")
	}
	if strings.TrimSpace(value) != value || strings.Contains(value, "/") || strings.Contains(value, "\\") {
		return "", invalidField(field, "Provider ID must be a valid lowercase public slug.")
	}
	if len(value) > models.NHIAActiveAccreditedHealthcareProviderIDMaxLength || !nhiaActiveAccreditedHealthcareProviderSlugPattern.MatchString(value) {
		return "", invalidField(field, "Provider ID must be a valid lowercase public slug.")
	}
	return value, nil
}

// ValidateNHIAActiveAccreditedHealthcareProviderListQuery validates NHIA HCP pagination and filters.
func ValidateNHIAActiveAccreditedHealthcareProviderListQuery(values url.Values) (interfaces.NHIAActiveAccreditedHealthcareProviderQuery, error) {
	var errs ValidationErrors
	query := interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50}
	allowed := map[string]struct{}{
		"page": {}, "page_size": {}, "provider_code": {}, "facility_type": {}, "listing_status": {}, "search": {},
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
	if raw, ok := readSingle("provider_code"); ok {
		query.ProviderCode = strings.TrimSpace(raw)
		if query.ProviderCode == "" {
			errs.Add("provider_code", codeRequired, "Provider code is required when provided.")
		} else if !nhiaActiveAccreditedHealthcareProviderCodePattern.MatchString(query.ProviderCode) {
			errs.Add("provider_code", codeInvalid, "Provider code must use the supported NHIA code format.")
		}
	}
	if raw, ok := readSingle("facility_type"); ok {
		query.FacilityType = strings.TrimSpace(raw)
		if query.FacilityType == "" {
			errs.Add("facility_type", codeRequired, "Facility type is required when provided.")
		} else if query.FacilityType != "primary" && query.FacilityType != "primary_and_secondary" {
			errs.Add("facility_type", codeInvalid, "Value is not supported.")
		}
	}
	if raw, ok := readSingle("listing_status"); ok {
		query.ListingStatus = strings.TrimSpace(raw)
		if query.ListingStatus == "" {
			errs.Add("listing_status", codeRequired, "Listing status is required when provided.")
		} else if query.ListingStatus != "active_accredited" {
			errs.Add("listing_status", codeInvalid, "Value is not supported.")
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
		return interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}, errs
	}
	return query, nil
}
