package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
	"github.com/AbdulQuayyum/softdata-api/internal/validators"
)

type apiKeyService interface {
	CreateKey(context.Context, string, models.APIKeyCreateInput) (models.APIKeyCreatedResponse, error)
	ListKeys(context.Context, string) ([]models.APIKeyMetadata, error)
	RevokeKey(context.Context, string, string) error
	RotateKey(context.Context, string, string) (models.APIKeyCreatedResponse, error)
}

type APIKeyHandler struct {
	apiKeys apiKeyService
}

func NewAPIKeyHandler(service apiKeyService) (*APIKeyHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("api key service is required")
	}

	return &APIKeyHandler{apiKeys: service}, nil
}

func (h *APIKeyHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	identity, ok := middlewares.AccountIdentityFromContext(r.Context())
	if !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	if err := requireJSONContentType(r); err != nil {
		writeInvalidRequest(w, requestID)
		return
	}

	var input models.APIKeyCreateInput
	if err := decodeJSONBody(r, &input); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			_ = response.Error(w, err, requestID)
			return
		}
		writeInvalidRequest(w, requestID)
		return
	}

	normalized, err := validators.ValidateAPIKeyCreate(input)
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	created, err := h.apiKeys.CreateKey(r.Context(), identity.AccountID, normalized)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusCreated, created)
}

func (h *APIKeyHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	identity, ok := middlewares.AccountIdentityFromContext(r.Context())
	if !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	keys, err := h.apiKeys.ListKeys(r.Context(), identity.AccountID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, keys)
}

func (h *APIKeyHandler) RevokeKey(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodDelete) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	identity, ok := middlewares.AccountIdentityFromContext(r.Context())
	if !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	keyID, err := apiKeyIDFromRequest(r)
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	if err := h.apiKeys.RevokeKey(r.Context(), identity.AccountID, keyID); err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	response.NoContent(w, http.StatusNoContent)
}

func (h *APIKeyHandler) RotateKey(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	identity, ok := middlewares.AccountIdentityFromContext(r.Context())
	if !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	keyID, err := apiKeyIDFromRequest(r)
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	rotated, err := h.apiKeys.RotateKey(r.Context(), identity.AccountID, keyID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, rotated)
}

func apiKeyIDFromRequest(r *http.Request) (string, error) {
	keyID, err := validators.ValidateAPIKeyID(r.PathValue("key_id"))
	if err != nil {
		return "", err
	}

	return keyID, nil
}
