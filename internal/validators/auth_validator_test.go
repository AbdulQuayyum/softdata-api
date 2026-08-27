package validators

import (
	"errors"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

func TestValidateRegistration(t *testing.T) {
	email := " Alice@Example.com "
	result, err := ValidateRegistration(models.AccountCreateInput{
		Username: " Alice ",
		Email:    &email,
		Password: "secret-password",
	})
	if err != nil {
		t.Fatalf("ValidateRegistration() error = %v", err)
	}
	if result.Username != "alice" {
		t.Fatalf("unexpected username: %q", result.Username)
	}
	if result.Email == nil || *result.Email != "alice@example.com" {
		t.Fatalf("unexpected email: %#v", result.Email)
	}
	if result.Password != "secret-password" {
		t.Fatalf("password was modified: %q", result.Password)
	}
}

func TestValidateRegistrationRejectsMissingFields(t *testing.T) {
	_, err := ValidateRegistration(models.AccountCreateInput{})
	var validationErr ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("ValidateRegistration() error = %v, want ValidationErrors", err)
	}
	if len(validationErr.Fields) != 2 {
		t.Fatalf("unexpected field error count: %#v", validationErr.Fields)
	}
	if validationErr.Fields[0].Field != "username" || validationErr.Fields[1].Field != "password" {
		t.Fatalf("unexpected field order: %#v", validationErr.Fields)
	}
}

func TestValidateLogin(t *testing.T) {
	result, err := ValidateLogin(LoginInput{
		Username: " Alice ",
		Password: "  secret  ",
	})
	if err != nil {
		t.Fatalf("ValidateLogin() error = %v", err)
	}
	if result.Username != "alice" {
		t.Fatalf("unexpected username: %q", result.Username)
	}
	if result.Password != "  secret  " {
		t.Fatalf("password was modified: %q", result.Password)
	}
}

func TestValidateRefreshAndLogout(t *testing.T) {
	if _, err := ValidateRefresh(RefreshInput{}); err == nil {
		t.Fatal("ValidateRefresh() error = nil, want error")
	}

	const token = "  opaque-refresh-token  "
	refresh, err := ValidateRefresh(RefreshInput{RefreshToken: token})
	if err != nil {
		t.Fatalf("ValidateRefresh() error = %v", err)
	}
	if refresh.RefreshToken != token {
		t.Fatalf("refresh token was modified: %q", refresh.RefreshToken)
	}

	logout, err := ValidateLogout(LogoutInput{RefreshToken: token})
	if err != nil {
		t.Fatalf("ValidateLogout() error = %v", err)
	}
	if logout.RefreshToken != token {
		t.Fatalf("logout token was modified: %q", logout.RefreshToken)
	}
}

func TestValidatePasswordChange(t *testing.T) {
	result, err := ValidatePasswordChange(PasswordChangeInput{
		CurrentPassword: "old-password",
		NewPassword:     "new-password",
	})
	if err != nil {
		t.Fatalf("ValidatePasswordChange() error = %v", err)
	}
	if result.CurrentPassword != "old-password" || result.NewPassword != "new-password" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
