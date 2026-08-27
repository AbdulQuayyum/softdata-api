package validators

import (
	"errors"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

func TestValidateAPIKeyCreate(t *testing.T) {
	result, err := ValidateAPIKeyCreate(models.APIKeyCreateInput{Name: " Portfolio Application "})
	if err != nil {
		t.Fatalf("ValidateAPIKeyCreate() error = %v", err)
	}
	if result.Name != "Portfolio Application" {
		t.Fatalf("unexpected name: %q", result.Name)
	}
}

func TestValidateAPIKeyCreateRejectsBlankName(t *testing.T) {
	_, err := ValidateAPIKeyCreate(models.APIKeyCreateInput{})
	var validationErr ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("ValidateAPIKeyCreate() error = %v, want ValidationErrors", err)
	}
	if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "name" {
		t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
	}
}

func TestValidateAPIKeyID(t *testing.T) {
	value, err := ValidateAPIKeyID(" key-123 ")
	if err != nil {
		t.Fatalf("ValidateAPIKeyID() error = %v", err)
	}
	if value != "key-123" {
		t.Fatalf("unexpected key id: %q", value)
	}
}
