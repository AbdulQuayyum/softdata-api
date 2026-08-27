package validators

import (
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// ValidateAPIKeyCreate validates the documented API-key creation payload.
func ValidateAPIKeyCreate(input models.APIKeyCreateInput) (models.APIKeyCreateInput, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return models.APIKeyCreateInput{}, ValidationErrors{Fields: []FieldError{{
			Field:   "name",
			Code:    codeRequired,
			Message: "Name is required.",
		}}}
	}

	return models.APIKeyCreateInput{Name: name}, nil
}

// ValidateAPIKeyID validates documented API-key path parameters.
func ValidateAPIKeyID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ValidationErrors{Fields: []FieldError{{
			Field:   "key_id",
			Code:    codeRequired,
			Message: "Key ID is required.",
		}}}
	}
	return value, nil
}
