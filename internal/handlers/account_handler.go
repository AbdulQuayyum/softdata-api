package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
	"github.com/AbdulQuayyum/softdata-api/internal/validators"
)

const accountRequestBodyLimit = 1 << 20

type accountService interface {
	GetCurrent(context.Context, string) (models.AccountResponse, error)
	UpdateCurrent(context.Context, string, models.AccountUpdateInput) (models.AccountResponse, error)
	DeactivateCurrent(context.Context, string) error
}

type AccountHandler struct {
	accounts accountService
}

func NewAccountHandler(accounts accountService) (*AccountHandler, error) {
	if accounts == nil {
		return nil, fmt.Errorf("account service is required")
	}

	return &AccountHandler{accounts: accounts}, nil
}

func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetCurrent(w, r)
	case http.MethodPatch:
		h.UpdateCurrent(w, r)
	case http.MethodDelete:
		h.DeleteCurrent(w, r)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPatch+", "+http.MethodDelete)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *AccountHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromContext(r.Context())
	identity, ok := middlewares.AccountIdentityFromContext(r.Context())
	if !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	account, err := h.accounts.GetCurrent(r.Context(), identity.AccountID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, account)
}

func (h *AccountHandler) UpdateCurrent(w http.ResponseWriter, r *http.Request) {
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

	var input models.AccountUpdateInput
	if err := decodeAccountJSONBody(r, &input); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			_ = response.Error(w, err, requestID)
			return
		}
		writeInvalidRequest(w, requestID)
		return
	}

	normalized, err := validators.ValidateAccountUpdate(input)
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	updated, err := h.accounts.UpdateCurrent(r.Context(), identity.AccountID, normalized)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, updated)
}

func (h *AccountHandler) DeleteCurrent(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromContext(r.Context())
	identity, ok := middlewares.AccountIdentityFromContext(r.Context())
	if !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	if err := h.accounts.DeactivateCurrent(r.Context(), identity.AccountID); err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	response.NoContent(w, http.StatusNoContent)
}

func decodeAccountJSONBody[T any](r *http.Request, dst *T) error {
	if r == nil || r.Body == nil {
		return fmt.Errorf("invalid request")
	}

	data, err := io.ReadAll(io.LimitReader(r.Body, accountRequestBodyLimit+1))
	if err != nil {
		return err
	}
	if len(data) == 0 || len(data) > accountRequestBodyLimit {
		return fmt.Errorf("invalid request")
	}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' || trimmed[len(trimmed)-1] != '}' {
		return fmt.Errorf("invalid request")
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("invalid request")
	}

	return nil
}
