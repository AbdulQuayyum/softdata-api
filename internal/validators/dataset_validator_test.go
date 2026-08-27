package validators

import (
	"errors"
	"testing"
)

func TestValidateDatasetKey(t *testing.T) {
	value, err := ValidateDatasetKey(" ng-states ")
	if err != nil {
		t.Fatalf("ValidateDatasetKey() error = %v", err)
	}
	if value != "ng-states" {
		t.Fatalf("unexpected dataset key: %q", value)
	}
}

func TestValidateDatasetKeyRejectsBlank(t *testing.T) {
	_, err := ValidateDatasetKey("   ")
	var validationErr ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("ValidateDatasetKey() error = %v, want ValidationErrors", err)
	}
	if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "dataset_id" {
		t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
	}
}

func TestValidateDatasetListQuery(t *testing.T) {
	result, err := ValidateDatasetListQuery("  states  ", "3", "10")
	if err != nil {
		t.Fatalf("ValidateDatasetListQuery() error = %v", err)
	}
	if result.Search != "states" {
		t.Fatalf("unexpected search value: %q", result.Search)
	}
	if result.Page != 3 || result.Limit != 10 || result.Offset != 20 {
		t.Fatalf("unexpected pagination: %#v", result.Pagination)
	}
}
