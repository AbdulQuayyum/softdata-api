package validators

import (
	"errors"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

func TestValidateAccountUpdate(t *testing.T) {
	email := " Alice@Example.com "
	result, err := ValidateAccountUpdate(models.AccountUpdateInput{
		Username: ptrString(" Alice "),
		Email:    &email,
	})
	if err != nil {
		t.Fatalf("ValidateAccountUpdate() error = %v", err)
	}
	if result.Username == nil || *result.Username != "alice" {
		t.Fatalf("unexpected username: %#v", result.Username)
	}
	if result.Email == nil || *result.Email != "alice@example.com" {
		t.Fatalf("unexpected email: %#v", result.Email)
	}
}

func TestValidateAccountUpdatePreservesExplicitEmptyEmail(t *testing.T) {
	empty := "   "
	result, err := ValidateAccountUpdate(models.AccountUpdateInput{
		Email: &empty,
	})
	if err != nil {
		t.Fatalf("ValidateAccountUpdate() error = %v", err)
	}
	if result.Email == nil {
		t.Fatal("explicit empty email was treated as omitted")
	}
	if *result.Email != "" {
		t.Fatalf("unexpected explicit empty email: %q", *result.Email)
	}
}

func TestValidateAccountUpdateRejectsEmptyUsername(t *testing.T) {
	_, err := ValidateAccountUpdate(models.AccountUpdateInput{
		Username: ptrString(" "),
	})
	if err == nil {
		t.Fatal("ValidateAccountUpdate() error = nil, want error")
	}
	var validationErr ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("ValidateAccountUpdate() error = %v, want ValidationErrors", err)
	}
	if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "username" {
		t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
	}
}

func ptrString(v string) *string {
	return &v
}
