package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/validators"
)

type emergencyServiceContactService interface {
	ListEmergencyServiceContacts(context.Context, interfaces.EmergencyServiceContactQuery) (interfaces.EmergencyServiceContactListResult, error)
	GetEmergencyServiceContact(context.Context, string) (models.EmergencyServiceContact, error)
}

type EmergencyServiceContactHandler struct {
	service emergencyServiceContactService
}

func NewEmergencyServiceContactHandler(service emergencyServiceContactService) (*EmergencyServiceContactHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("emergency service contact service is required")
	}
	return &EmergencyServiceContactHandler{service: service}, nil
}

func (h *EmergencyServiceContactHandler) ListEmergencyServiceContacts(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateEmergencyServiceContactListQuery(r.URL.Query())
	if err != nil {
		h.writeValidation(w, requestID, err)
		return
	}
	result, err := h.service.ListEmergencyServiceContacts(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.PaginatedPageSize(w, http.StatusOK, result.Records, response.PageSizePaginationMeta{
		Page: result.Page, PageSize: result.PageSize, Total: int64(result.Total), TotalPages: result.TotalPages,
	})
}

func (h *EmergencyServiceContactHandler) GetEmergencyServiceContact(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	id, err := validators.ValidateEmergencyServiceContactID("contact_id", r.PathValue("contact_id"))
	if err != nil {
		h.writeValidation(w, requestID, err)
		return
	}
	record, err := h.service.GetEmergencyServiceContact(r.Context(), id)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Success(w, http.StatusOK, record)
}

func (h *EmergencyServiceContactHandler) writeValidation(w http.ResponseWriter, requestID string, err error) {
	validationErr, ok := validationErrorsFrom(err)
	if !ok {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.ValidationBadRequest(w, requestID, validationErrorsToResponse(validationErr))
}
