package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/validators"
)

type geographyService interface {
	ListStates(context.Context) ([]models.State, error)
	GetState(context.Context, string) (models.State, error)
}

type GeographyHandler struct {
	service geographyService
}

func NewGeographyHandler(service geographyService) (*GeographyHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("geography service is required")
	}
	return &GeographyHandler{service: service}, nil
}

func (h *GeographyHandler) ListStates(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	states, err := h.service.ListStates(r.Context())
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, states)
}

func (h *GeographyHandler) GetState(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	stateID, err := validators.ValidateStateID(r.PathValue("state_id"))
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	state, err := h.service.GetState(r.Context(), stateID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, state)
}
