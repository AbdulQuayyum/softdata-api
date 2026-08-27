package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
	"github.com/AbdulQuayyum/softdata-api/internal/validators"
)

type accountRegistrar interface {
	Register(context.Context, models.AccountCreateInput) (models.AccountResponse, error)
}

type authExecutor interface {
	Login(context.Context, string, string) (models.LoginResult, error)
	Refresh(context.Context, string) (models.TokenPair, error)
	Logout(context.Context, string) error
}

type AuthHandler struct {
	accounts accountRegistrar
	auth     authExecutor
}

func NewAuthHandler(accounts accountRegistrar, auth authExecutor) (*AuthHandler, error) {
	switch {
	case accounts == nil:
		return nil, fmt.Errorf("account registrar is required")
	case auth == nil:
		return nil, fmt.Errorf("auth service is required")
	}

	return &AuthHandler{
		accounts: accounts,
		auth:     auth,
	}, nil
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	if err := requireJSONContentType(r); err != nil {
		writeInvalidRequest(w, requestID)
		return
	}

	var input models.AccountCreateInput
	if err := decodeJSONBody(r, &input); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			_ = response.Error(w, err, requestID)
			return
		}
		writeInvalidRequest(w, requestID)
		return
	}

	normalized, err := validators.ValidateRegistration(input)
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	account, err := h.accounts.Register(r.Context(), normalized)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusCreated, account)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	if err := requireJSONContentType(r); err != nil {
		writeInvalidRequest(w, requestID)
		return
	}

	var input loginRequest
	if err := decodeJSONBody(r, &input); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			_ = response.Error(w, err, requestID)
			return
		}
		writeInvalidRequest(w, requestID)
		return
	}

	normalized, err := validators.ValidateLogin(validators.LoginInput{
		Username: input.Username,
		Password: input.Password,
	})
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	result, err := h.auth.Login(r.Context(), normalized.Username, normalized.Password)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, result)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	if err := requireJSONContentType(r); err != nil {
		writeInvalidRequest(w, requestID)
		return
	}

	var input refreshRequest
	if err := decodeJSONBody(r, &input); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			_ = response.Error(w, err, requestID)
			return
		}
		writeInvalidRequest(w, requestID)
		return
	}

	normalized, err := validators.ValidateRefresh(validators.RefreshInput{
		RefreshToken: input.RefreshToken,
	})
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	result, err := h.auth.Refresh(r.Context(), normalized.RefreshToken)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, result)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	if err := requireJSONContentType(r); err != nil {
		writeInvalidRequest(w, requestID)
		return
	}

	if _, ok := middlewares.AccountIdentityFromContext(r.Context()); !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	var input logoutRequest
	if err := decodeJSONBody(r, &input); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			_ = response.Error(w, err, requestID)
			return
		}
		writeInvalidRequest(w, requestID)
		return
	}

	normalized, err := validators.ValidateLogout(validators.LogoutInput{
		RefreshToken: input.RefreshToken,
	})
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	if err := h.auth.Logout(r.Context(), normalized.RefreshToken); err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	response.NoContent(w, http.StatusNoContent)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func allowMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	w.WriteHeader(http.StatusMethodNotAllowed)
	return false
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := middlewares.RequestIDFromContext(ctx)
	return requestID
}

func requireJSONContentType(r *http.Request) error {
	if r == nil {
		return fmt.Errorf("invalid request")
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	if mediaType != "application/json" {
		return fmt.Errorf("invalid request")
	}
	return nil
}

func decodeJSONBody[T any](r *http.Request, dst *T) error {
	if r == nil || r.Body == nil {
		return fmt.Errorf("invalid request")
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return err
		}
		return err
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

func validationErrorsFrom(err error) (validators.ValidationErrors, bool) {
	var validationErr validators.ValidationErrors
	if !errors.As(err, &validationErr) {
		return validators.ValidationErrors{}, false
	}
	return validationErr, true
}

func validationErrorsToResponse(err validators.ValidationErrors) []response.ValidationError {
	details := make([]response.ValidationError, 0, len(err.Fields))
	for _, field := range err.Fields {
		details = append(details, response.ValidationError{
			Field:   field.Field,
			Message: field.Message,
		})
	}
	return details
}

func writeInvalidRequest(w http.ResponseWriter, requestID string) {
	_ = response.Error(w, services.ErrInvalidPagination, requestID)
}
