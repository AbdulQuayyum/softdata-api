package validators

import (
	"errors"
	"net/url"
	"testing"
	"time"
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

func TestValidateUsageQueryDefaultsToCurrentUTCMonth(t *testing.T) {
	query, err := ValidateUsageQuery(url.Values{}, time.Date(2026, 8, 27, 16, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ValidateUsageQuery() error = %v", err)
	}
	if !query.Start.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)) || !query.End.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected default usage range: %#v", query)
	}
}

func TestValidateUsageQueryParsesExplicitRange(t *testing.T) {
	values := url.Values{
		"start": []string{"2026-08-01"},
		"end":   []string{"2026-09-01"},
	}
	query, err := ValidateUsageQuery(values, time.Date(2026, 8, 27, 16, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ValidateUsageQuery() error = %v", err)
	}
	if !query.Start.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)) || !query.End.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected explicit usage range: %#v", query)
	}
}

func TestValidateUsageQueryRejectsPartialRangeAndDuplicates(t *testing.T) {
	t.Run("missing end", func(t *testing.T) {
		_, err := ValidateUsageQuery(url.Values{"start": []string{"2026-08-01"}}, time.Date(2026, 8, 27, 16, 30, 0, 0, time.UTC))
		var validationErr ValidationErrors
		if !errors.As(err, &validationErr) {
			t.Fatalf("ValidateUsageQuery() error = %v, want ValidationErrors", err)
		}
		if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "end" {
			t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
		}
	})

	t.Run("missing start", func(t *testing.T) {
		_, err := ValidateUsageQuery(url.Values{"end": []string{"2026-08-01"}}, time.Date(2026, 8, 27, 16, 30, 0, 0, time.UTC))
		var validationErr ValidationErrors
		if !errors.As(err, &validationErr) {
			t.Fatalf("ValidateUsageQuery() error = %v, want ValidationErrors", err)
		}
		if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "start" {
			t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
		}
	})

	t.Run("duplicate start", func(t *testing.T) {
		_, err := ValidateUsageQuery(url.Values{"start": []string{"2026-08-01", "2026-08-02"}, "end": []string{"2026-09-01"}}, time.Date(2026, 8, 27, 16, 30, 0, 0, time.UTC))
		var validationErr ValidationErrors
		if !errors.As(err, &validationErr) {
			t.Fatalf("ValidateUsageQuery() error = %v, want ValidationErrors", err)
		}
		if len(validationErr.Fields) == 0 || validationErr.Fields[0].Field != "start" {
			t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
		}
	})

	t.Run("same day", func(t *testing.T) {
		_, err := ValidateUsageQuery(url.Values{"start": []string{"2026-08-01"}, "end": []string{"2026-08-01"}}, time.Date(2026, 8, 27, 16, 30, 0, 0, time.UTC))
		var validationErr ValidationErrors
		if !errors.As(err, &validationErr) {
			t.Fatalf("ValidateUsageQuery() error = %v, want ValidationErrors", err)
		}
		if len(validationErr.Fields) == 0 || validationErr.Fields[0].Field != "end" {
			t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
		}
	})

	t.Run("malformed date", func(t *testing.T) {
		_, err := ValidateUsageQuery(url.Values{"start": []string{"2026-08-xx"}, "end": []string{"2026-09-01"}}, time.Date(2026, 8, 27, 16, 30, 0, 0, time.UTC))
		var validationErr ValidationErrors
		if !errors.As(err, &validationErr) {
			t.Fatalf("ValidateUsageQuery() error = %v, want ValidationErrors", err)
		}
		if len(validationErr.Fields) == 0 || validationErr.Fields[0].Field != "start" {
			t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
		}
	})
}

func TestValidateUsageQueryValidatesAPIKeyID(t *testing.T) {
	query, err := ValidateUsageQuery(url.Values{"api_key_id": []string{"  key-1  "}}, time.Date(2026, 8, 27, 16, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ValidateUsageQuery() error = %v", err)
	}
	if query.APIKeyID != "key-1" {
		t.Fatalf("unexpected api key id: %#v", query)
	}
}
