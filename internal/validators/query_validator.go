package validators

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	codeRequired   = "required"
	codeInvalid    = "invalid"
	codeMalformed  = "malformed"
	codeOutOfRange = "out_of_range"

	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
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
