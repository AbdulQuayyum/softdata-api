package validators

import (
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// ValidateAccountUpdate validates the documented account-update payload.
func ValidateAccountUpdate(input models.AccountUpdateInput) (models.AccountUpdateInput, error) {
	var errs ValidationErrors

	normalized := input

	if input.Username != nil {
		username := strings.ToLower(strings.TrimSpace(*input.Username))
		if username == "" {
			errs.Add("username", codeRequired, "Username cannot be empty.")
		} else {
			normalized.Username = &username
		}
	}

	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		if email == "" {
			empty := ""
			normalized.Email = &empty
		} else if err := validateEmail(email); err != nil {
			errs.Add("email", codeInvalid, "Email must be a valid email address.")
		} else {
			normalized.Email = &email
		}
	}

	if len(errs.Fields) > 0 {
		return models.AccountUpdateInput{}, errs
	}

	return normalized, nil
}

func validateEmail(value string) error {
	if value == "" {
		return nil
	}
	_, err := parseEmailAddress(value)
	return err
}
