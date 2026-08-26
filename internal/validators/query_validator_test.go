package validators

import (
	"errors"
	"testing"
)

func TestValidatePaginationUsesDefaults(t *testing.T) {
	result, err := ValidatePagination("", "")
	if err != nil {
		t.Fatalf("ValidatePagination() error = %v", err)
	}
	if result.Page != 1 || result.Limit != 20 || result.Offset != 0 {
		t.Fatalf("unexpected pagination result: %#v", result)
	}
}

func TestValidatePaginationParsesAndBounds(t *testing.T) {
	result, err := ValidatePagination("2", "25")
	if err != nil {
		t.Fatalf("ValidatePagination() error = %v", err)
	}
	if result.Page != 2 || result.Limit != 25 || result.Offset != 25 {
		t.Fatalf("unexpected pagination result: %#v", result)
	}
}

func TestValidatePaginationRejectsBadValues(t *testing.T) {
	t.Run("page", func(t *testing.T) {
		_, err := ValidatePagination("0", "20")
		var validationErr ValidationErrors
		if !errors.As(err, &validationErr) {
			t.Fatalf("ValidatePagination() error = %v, want ValidationErrors", err)
		}
		if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "page" {
			t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
		}
	})

	t.Run("limit", func(t *testing.T) {
		_, err := ValidatePagination("1", "101")
		var validationErr ValidationErrors
		if !errors.As(err, &validationErr) {
			t.Fatalf("ValidatePagination() error = %v, want ValidationErrors", err)
		}
		if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "limit" {
			t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
		}
	})
}

func TestValidateSearch(t *testing.T) {
	got := ValidateSearch("  kwara state  ")
	if got != "kwara state" {
		t.Fatalf("unexpected search value: %q", got)
	}
}
