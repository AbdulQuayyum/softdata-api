package validators

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var publicEducationSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var allowedSchoolEducationLevels = map[string]struct{}{
	"pre-primary": {}, "primary": {}, "junior-secondary": {}, "senior-secondary": {},
}

var allowedSchoolOwnershipTypes = map[string]struct{}{"public": {}, "private": {}}

// ValidateEducationInstitutionID validates an institution or school public slug.
func ValidateEducationInstitutionID(field, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError(field, "Institution ID is required.")
	}
	if !publicEducationSlugPattern.MatchString(value) {
		return "", invalidField(field, "Institution ID must be a valid lowercase public slug.")
	}
	return value, nil
}

// ValidateSchoolListQuery validates the HTTP-facing school pagination and filters.
func ValidateSchoolListQuery(values url.Values) (interfaces.PrimaryAndSecondarySchoolQuery, error) {
	var errs ValidationErrors
	query := interfaces.PrimaryAndSecondarySchoolQuery{Page: 1, PageSize: 50}

	parseSingle := func(field string, target *[]string) {
		items := values[field]
		if len(items) > 1 {
			errs.Add(field, codeMalformed, "Field may be provided at most once.")
			return
		}
		if len(items) == 1 {
			*target = items
		}
	}

	var pageValues, pageSizeValues, stateValues, lgaValues, levelValues, ownershipValues, searchValues []string
	parseSingle("page", &pageValues)
	parseSingle("page_size", &pageSizeValues)
	parseSingle("state_id", &stateValues)
	parseSingle("lga_id", &lgaValues)
	parseSingle("education_level", &levelValues)
	parseSingle("ownership_type", &ownershipValues)
	parseSingle("search", &searchValues)
	if len(errs.Fields) > 0 {
		return interfaces.PrimaryAndSecondarySchoolQuery{}, errs
	}

	parsePositive := func(field string, raw []string, defaultValue int, minimum, maximum int) int {
		if len(raw) == 0 {
			return defaultValue
		}
		value, err := strconv.ParseUint(raw[0], 10, 0)
		if err != nil || value > uint64(^uint(0)>>1) || int(value) < minimum || (maximum > 0 && int(value) > maximum) {
			errs.Add(field, codeInvalid, "Value must be a valid integer within the supported range.")
			return defaultValue
		}
		return int(value)
	}
	query.Page = parsePositive("page", pageValues, 1, 1, 0)
	query.PageSize = parsePositive("page_size", pageSizeValues, 50, 1, 100)

	validateOptionalSlug := func(field string, raw []string) string {
		if len(raw) == 0 {
			return ""
		}
		value := strings.TrimSpace(raw[0])
		if value == "" {
			return ""
		}
		if !publicEducationSlugPattern.MatchString(value) {
			errs.Add(field, codeInvalid, "Value must be a valid lowercase public slug.")
			return ""
		}
		return value
	}
	query.StateID = validateOptionalSlug("state_id", stateValues)
	query.LGAID = validateOptionalSlug("lga_id", lgaValues)

	if len(levelValues) == 1 {
		query.EducationLevel = strings.TrimSpace(levelValues[0])
		if query.EducationLevel != "" {
			if _, ok := allowedSchoolEducationLevels[query.EducationLevel]; !ok {
				errs.Add("education_level", codeInvalid, "Education level is not supported.")
			}
		}
	}
	if len(ownershipValues) == 1 {
		query.OwnershipType = strings.TrimSpace(ownershipValues[0])
		if query.OwnershipType != "" {
			if _, ok := allowedSchoolOwnershipTypes[query.OwnershipType]; !ok {
				errs.Add("ownership_type", codeInvalid, "Ownership type is not supported.")
			}
		}
	}
	if len(searchValues) == 1 {
		query.Search = strings.TrimSpace(searchValues[0])
		if len(query.Search) > 100 {
			errs.Add("search", codeInvalid, "Search must be at most 100 characters.")
		}
	}

	if len(errs.Fields) > 0 {
		return interfaces.PrimaryAndSecondarySchoolQuery{}, errs
	}
	return query, nil
}
