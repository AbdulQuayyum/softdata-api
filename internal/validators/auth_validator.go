package validators

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// LoginInput captures the documented login fields.
type LoginInput struct {
	Username string
	Password string
}

// RefreshInput captures the documented refresh-token field.
type RefreshInput struct {
	RefreshToken string
}

// LogoutInput captures the documented logout refresh-token field.
type LogoutInput struct {
	RefreshToken string
}

// PasswordChangeInput captures the documented password-change fields.
type PasswordChangeInput struct {
	CurrentPassword string
	NewPassword     string
}

// ValidateRegistration validates the documented registration payload.
func ValidateRegistration(input models.AccountCreateInput) (models.AccountCreateInput, error) {
	var errs ValidationErrors

	normalized := input
	normalized.Username = strings.ToLower(strings.TrimSpace(input.Username))
	if normalized.Username == "" {
		errs.Add("username", codeRequired, "Username is required.")
	}

	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		if email == "" {
			normalized.Email = nil
		} else if _, err := mail.ParseAddress(email); err != nil {
			errs.Add("email", codeInvalid, "Email must be a valid email address.")
		} else {
			normalized.Email = &email
		}
	}

	if input.Password == "" {
		errs.Add("password", codeRequired, "Password is required.")
	}

	if len(errs.Fields) > 0 {
		return models.AccountCreateInput{}, errs
	}

	return normalized, nil
}

// ValidateLogin validates the documented login payload.
func ValidateLogin(input LoginInput) (LoginInput, error) {
	var errs ValidationErrors

	normalized := input
	normalized.Username = strings.ToLower(strings.TrimSpace(input.Username))
	if normalized.Username == "" {
		errs.Add("username", codeRequired, "Username is required.")
	}
	if input.Password == "" {
		errs.Add("password", codeRequired, "Password is required.")
	}

	if len(errs.Fields) > 0 {
		return LoginInput{}, errs
	}

	return normalized, nil
}

// ValidateRefresh validates the documented refresh-token payload.
func ValidateRefresh(input RefreshInput) (RefreshInput, error) {
	if input.RefreshToken == "" {
		return RefreshInput{}, ValidationErrors{Fields: []FieldError{{
			Field:   "refresh_token",
			Code:    codeRequired,
			Message: "Refresh token is required.",
		}}}
	}
	return input, nil
}

// ValidateLogout validates the documented logout payload.
func ValidateLogout(input LogoutInput) (LogoutInput, error) {
	if input.RefreshToken == "" {
		return LogoutInput{}, ValidationErrors{Fields: []FieldError{{
			Field:   "refresh_token",
			Code:    codeRequired,
			Message: "Refresh token is required.",
		}}}
	}
	return input, nil
}

// ValidatePasswordChange validates the documented change-password payload.
func ValidatePasswordChange(input PasswordChangeInput) (PasswordChangeInput, error) {
	var errs ValidationErrors

	if input.CurrentPassword == "" {
		errs.Add("current_password", codeRequired, "Current password is required.")
	}
	if input.NewPassword == "" {
		errs.Add("new_password", codeRequired, "New password is required.")
	}

	if len(errs.Fields) > 0 {
		return PasswordChangeInput{}, errs
	}

	return input, nil
}

func requiredError(field, message string) error {
	return ValidationErrors{Fields: []FieldError{{
		Field:   field,
		Code:    codeRequired,
		Message: message,
	}}}
}

func validationError(field, code, message string) error {
	return ValidationErrors{Fields: []FieldError{{
		Field:   field,
		Code:    code,
		Message: message,
	}}}
}

func invalidField(field, message string) error {
	return validationError(field, codeInvalid, message)
}

func invalidRequired(field string) error {
	return requiredError(field, fmt.Sprintf("%s is required.", field))
}

func parseEmailAddress(value string) (string, error) {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return "", err
	}
	return addr.Address, nil
}
