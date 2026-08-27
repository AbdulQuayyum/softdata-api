package validators

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	codeRequired   = "required"
	codeInvalid    = "invalid"
	codeMalformed  = "malformed"
	codeOutOfRange = "out_of_range"

	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
	maxUsageDays = 366
)

// FieldError describes a single field-level validation issue.
type FieldError struct {
	Field   string
	Code    string
	Message string
}

// ValidationErrors groups multiple field-level validation issues.
type ValidationErrors struct {
	Fields []FieldError
}

func (e *ValidationErrors) Add(field, code, message string) {
	if e == nil {
		return
	}
	e.Fields = append(e.Fields, FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}

func (e ValidationErrors) Error() string {
	if len(e.Fields) == 0 {
		return "validation failed"
	}

	var b strings.Builder
	b.WriteString("validation failed: ")
	for i, field := range e.Fields {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(field.Field)
		b.WriteString(": ")
		b.WriteString(field.Message)
	}
	return b.String()
}

// Pagination contains validated pagination values.
type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

// UsageQuery contains validated usage analytics query parameters.
type UsageQuery struct {
	Start    time.Time
	End      time.Time
	APIKeyID string
}

// ValidatePagination validates documented page and limit query parameters.
func ValidatePagination(pageValue, limitValue string) (Pagination, error) {
	var errs ValidationErrors

	page, err := parsePositiveInt(pageValue, defaultPage, "page", &errs)
	if err != nil {
		return Pagination{}, err
	}

	limit, err := parsePositiveInt(limitValue, defaultLimit, "limit", &errs)
	if err != nil {
		return Pagination{}, err
	}
	if limit > maxLimit {
		errs.Add("limit", codeOutOfRange, fmt.Sprintf("Limit must not exceed %d.", maxLimit))
	}

	if len(errs.Fields) > 0 {
		return Pagination{}, errs
	}

	offset, err := safeOffset(page, limit)
	if err != nil {
		return Pagination{}, err
	}

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// ValidateSearch trims surrounding whitespace from documented search inputs.
func ValidateSearch(value string) string {
	return strings.TrimSpace(value)
}

// ValidateUsageQuery validates documented usage analytics query parameters.
func ValidateUsageQuery(values url.Values, now time.Time) (UsageQuery, error) {
	var errs ValidationErrors

	startValues := values["start"]
	endValues := values["end"]
	apiKeyIDValues := values["api_key_id"]

	if len(startValues) > 1 {
		errs.Add("start", codeMalformed, "Start may be provided at most once.")
	}
	if len(endValues) > 1 {
		errs.Add("end", codeMalformed, "End may be provided at most once.")
	}
	if len(apiKeyIDValues) > 1 {
		errs.Add("api_key_id", codeMalformed, "API key ID may be provided at most once.")
	}

	startValue := firstValue(startValues)
	endValue := firstValue(endValues)
	apiKeyIDValue := firstValue(apiKeyIDValues)

	var query UsageQuery
	if apiKeyIDValue != "" {
		normalized, err := validateUsageAPIKeyID(apiKeyIDValue)
		if err != nil {
			if validationErr, ok := err.(ValidationErrors); ok {
				errs.Fields = append(errs.Fields, validationErr.Fields...)
			} else {
				return UsageQuery{}, err
			}
		} else {
			query.APIKeyID = normalized
		}
	}

	switch {
	case startValue == "" && endValue == "":
		start, end := currentUTCCalendarMonth(now)
		query.Start = start
		query.End = end
	case startValue == "" || endValue == "":
		if startValue == "" {
			errs.Add("start", codeRequired, "Start is required when end is provided.")
		}
		if endValue == "" {
			errs.Add("end", codeRequired, "End is required when start is provided.")
		}
	default:
		start, err := parseUsageDateQueryValue(startValue, "start")
		if err != nil {
			if validationErr, ok := err.(ValidationErrors); ok {
				errs.Fields = append(errs.Fields, validationErr.Fields...)
			} else {
				return UsageQuery{}, err
			}
		} else {
			query.Start = start
		}

		end, err := parseUsageDateQueryValue(endValue, "end")
		if err != nil {
			if validationErr, ok := err.(ValidationErrors); ok {
				errs.Fields = append(errs.Fields, validationErr.Fields...)
			} else {
				return UsageQuery{}, err
			}
		} else {
			query.End = end
		}

		if !query.Start.IsZero() && !query.End.IsZero() && !query.End.After(query.Start) {
			errs.Add("end", codeOutOfRange, "End must be after start.")
		}
		if !query.Start.IsZero() && !query.End.IsZero() && usageRangeExceedsMaximum(query.Start, query.End) {
			errs.Add("end", codeOutOfRange, fmt.Sprintf("Usage range must not exceed %d days.", maxUsageDays))
		}
	}

	if len(errs.Fields) > 0 {
		return UsageQuery{}, errs
	}

	return query, nil
}

func parsePositiveInt(value string, fallback int, field string, errs *ValidationErrors) (int, error) {
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		errs.Add(field, codeMalformed, fmt.Sprintf("%s must be a valid integer.", fieldLabel(field)))
		return 0, nil
	}
	if parsed < 1 {
		errs.Add(field, codeOutOfRange, fmt.Sprintf("%s must be at least 1.", fieldLabel(field)))
		return 0, nil
	}
	if parsed > int64(math.MaxInt) {
		errs.Add(field, codeOutOfRange, fmt.Sprintf("%s is too large.", fieldLabel(field)))
		return 0, nil
	}
	return int(parsed), nil
}

func safeOffset(page, limit int) (int, error) {
	if page < 1 || limit < 1 {
		return 0, ValidationErrors{Fields: []FieldError{{
			Field:   "page",
			Code:    codeOutOfRange,
			Message: "Pagination values are invalid.",
		}}}
	}
	offset := int64(page-1) * int64(limit)
	if offset < 0 || offset > int64(math.MaxInt) {
		return 0, ValidationErrors{Fields: []FieldError{{
			Field:   "page",
			Code:    codeOutOfRange,
			Message: "Pagination offset is too large.",
		}}}
	}
	return int(offset), nil
}

func fieldLabel(field string) string {
	if field == "" {
		return field
	}
	return strings.ToUpper(field[:1]) + field[1:]
}

func firstValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func parseUsageDateQueryValue(value, field string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, ValidationErrors{Fields: []FieldError{{
			Field:   field,
			Code:    codeRequired,
			Message: fieldLabel(field) + " is required.",
		}}}
	}

	parsed, err := time.ParseInLocation("2006-01-02", value, time.UTC)
	if err != nil {
		return time.Time{}, ValidationErrors{Fields: []FieldError{{
			Field:   field,
			Code:    codeMalformed,
			Message: fieldLabel(field) + " must use YYYY-MM-DD.",
		}}}
	}
	return parsed.UTC(), nil
}

func currentUTCCalendarMonth(now time.Time) (time.Time, time.Time) {
	utcNow := now.UTC()
	start := time.Date(utcNow.Year(), utcNow.Month(), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0)
}

func validateUsageAPIKeyID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ValidationErrors{Fields: []FieldError{{
			Field:   "api_key_id",
			Code:    codeRequired,
			Message: "API key ID is required.",
		}}}
	}
	return value, nil
}

func usageRangeExceedsMaximum(start, end time.Time) bool {
	if start.IsZero() || end.IsZero() {
		return false
	}
	return end.UTC().Sub(start.UTC()) > time.Duration(maxUsageDays)*24*time.Hour
}
