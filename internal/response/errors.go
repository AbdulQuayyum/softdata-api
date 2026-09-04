package response

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

const (
	codeInvalidRequest     = "INVALID_REQUEST"
	codeValidationFailed   = "VALIDATION_FAILED"
	codeInvalidCredentials = "INVALID_CREDENTIALS"
	codeInvalidAPIKey      = "INVALID_API_KEY"
	codeResourceNotFound   = "RESOURCE_NOT_FOUND"
	codeResourceConflict   = "RESOURCE_CONFLICT"
	codeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	codeInternalError      = "INTERNAL_ERROR"
	codeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

const (
	messageInvalidRequest      = "The request was invalid."
	messageValidationFailed    = "One or more fields are invalid."
	messageInvalidCredentials  = "Invalid credentials."
	messageInvalidAPIKey       = "Invalid API key."
	messageResourceNotFound    = "The requested resource was not found."
	messageResourceConflict    = "The resource already exists."
	messageRateLimitExceeded   = "The request limit has been exceeded."
	messageInternalError       = "An unexpected server error occurred."
	messageServiceUnavailable  = "The service is temporarily unavailable."
	messageOperationNotAllowed = "The requested operation is not permitted."
)

// ValidationError describes a single field-level validation issue.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorBody matches the documented error payload body.
type ErrorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Details   []ValidationError `json:"details"`
	RequestID string            `json:"request_id"`
}

// ErrorResponse matches the documented error envelope.
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

type mappedError struct {
	status  int
	code    string
	message string
}

// Error writes a mapped public error response using the documented envelope.
func Error(w http.ResponseWriter, err error, requestID string) error {
	if err == nil {
		return nil
	}
	mapped := mapError(err)
	body := ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:      mapped.code,
			Message:   mapped.message,
			RequestID: requestID,
		},
	}
	return JSON(w, mapped.status, body)
}

// Validation writes a validation error response with stable field-level details.
func Validation(w http.ResponseWriter, requestID string, details []ValidationError) error {
	if details == nil {
		details = []ValidationError{}
	}
	body := ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:      codeValidationFailed,
			Message:   messageValidationFailed,
			Details:   details,
			RequestID: requestID,
		},
	}
	return JSON(w, http.StatusUnprocessableEntity, body)
}

func mapError(err error) mappedError {
	if err == nil {
		return mappedError{
			status:  http.StatusOK,
			code:    "",
			message: "",
		}
	}

	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return mappedError{status: http.StatusServiceUnavailable, code: codeServiceUnavailable, message: messageServiceUnavailable}
	case errors.Is(err, services.ErrInvalidCredentials), errors.Is(err, services.ErrInvalidRefreshToken), errors.Is(err, services.ErrCurrentPasswordInvalid):
		return mappedError{status: http.StatusUnauthorized, code: codeInvalidCredentials, message: messageInvalidCredentials}
	case errors.Is(err, services.ErrInvalidDatasetKey), errors.Is(err, services.ErrInvalidPagination), errors.Is(err, services.ErrAPIKeyNameRequired), errors.Is(err, services.ErrInvalidUsagePeriod):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrInvalidStateID):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrInvalidLocalGovernmentUnitID):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrInvalidGeopoliticalZoneID):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrInvalidLanguageID), errors.Is(err, services.ErrInvalidCountryLanguageCountryAreaID), errors.Is(err, services.ErrInvalidCountryLanguageLanguageID), errors.Is(err, services.ErrInvalidCountryLanguageStatus):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrDatasetNotFound), errors.Is(err, services.ErrAccountNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrAPIKeyNotFound):
		return mappedError{status: http.StatusUnauthorized, code: codeInvalidAPIKey, message: messageInvalidAPIKey}
	case errors.Is(err, services.ErrUsernameUnavailable), errors.Is(err, services.ErrEmailUnavailable), errors.Is(err, services.ErrAPIKeyLimitReached):
		return mappedError{status: http.StatusConflict, code: codeResourceConflict, message: messageResourceConflict}
	case errors.Is(err, services.ErrAccountInactive):
		return mappedError{status: http.StatusForbidden, code: codeInvalidRequest, message: messageOperationNotAllowed}
	case errors.Is(err, services.ErrStateNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrLanguageNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrUniversityNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrInvalidUniversityID), errors.Is(err, services.ErrInvalidUniversityOwnershipType), errors.Is(err, services.ErrInvalidUniversityStateID):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrCollegeOfEducationNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrInvalidCollegeOfEducationID), errors.Is(err, services.ErrInvalidCollegeOfEducationOwnershipType), errors.Is(err, services.ErrInvalidCollegeOfEducationStateID):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrPaymentServiceProviderNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrInvalidPaymentServiceProviderID), errors.Is(err, services.ErrInvalidPaymentServiceProviderType):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrInternationalMoneyTransferOperatorNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrInvalidInternationalMoneyTransferOperatorID):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrCurrencyNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrInvalidCurrencyID), errors.Is(err, services.ErrInvalidCurrencyCountryAreaID):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrCountryOrAreaNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrInvalidCountryOrAreaID), errors.Is(err, services.ErrInvalidCountryOrAreaRegionCode), errors.Is(err, services.ErrInvalidCountryOrAreaSubregionCode):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrLocalGovernmentUnitNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrTimeZoneNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, services.ErrInvalidTimeZoneID), errors.Is(err, services.ErrInvalidTimeZoneCountryAreaID):
		return mappedError{status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest}
	case errors.Is(err, services.ErrGeopoliticalZoneNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, interfaces.ErrNotFound):
		return mappedError{status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound}
	case errors.Is(err, interfaces.ErrConflict):
		return mappedError{status: http.StatusConflict, code: codeResourceConflict, message: messageResourceConflict}
	default:
		return mappedError{status: http.StatusInternalServerError, code: codeInternalError, message: messageInternalError}
	}
}

func validatePaginationMeta(meta PaginationMeta) error {
	if meta.Page < 1 || meta.Limit < 1 || meta.Total < 0 || meta.TotalPages < 0 {
		return fmt.Errorf("invalid pagination metadata")
	}
	if meta.Total == 0 && meta.TotalPages != 0 {
		return fmt.Errorf("invalid pagination metadata")
	}
	if meta.Total > 0 && meta.TotalPages == 0 {
		return fmt.Errorf("invalid pagination metadata")
	}
	return nil
}
