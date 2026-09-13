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

type nhiaAccreditedHealthMaintenanceOrganisationService interface {
	ListNHIAAccreditedHealthMaintenanceOrganisations(context.Context, interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult, error)
	GetNHIAAccreditedHealthMaintenanceOrganisation(context.Context, string) (models.NHIAAccreditedHealthMaintenanceOrganisation, error)
}

// NHIAAccreditedHealthMaintenanceOrganisationHandler serves the public NHIA accredited HMO contract.
type NHIAAccreditedHealthMaintenanceOrganisationHandler struct {
	service nhiaAccreditedHealthMaintenanceOrganisationService
}

// NewNHIAAccreditedHealthMaintenanceOrganisationHandler constructs a handler with a narrow service dependency.
func NewNHIAAccreditedHealthMaintenanceOrganisationHandler(service nhiaAccreditedHealthMaintenanceOrganisationService) (*NHIAAccreditedHealthMaintenanceOrganisationHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("nhia accredited health maintenance organisation service is required")
	}
	return &NHIAAccreditedHealthMaintenanceOrganisationHandler{service: service}, nil
}

// ListNHIAAccreditedHealthMaintenanceOrganisations handles GET /v1/healthcare/nhia-accredited-health-maintenance-organisations.
func (h *NHIAAccreditedHealthMaintenanceOrganisationHandler) ListNHIAAccreditedHealthMaintenanceOrganisations(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateNHIAAccreditedHealthMaintenanceOrganisationListQuery(r.URL.Query())
	if err != nil {
		h.writeValidation(w, requestID, err)
		return
	}
	result, err := h.service.ListNHIAAccreditedHealthMaintenanceOrganisations(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.PaginatedPageSize(w, http.StatusOK, result.Records, response.PageSizePaginationMeta{
		Page: result.Page, PageSize: result.PageSize, Total: int64(result.Total), TotalPages: result.TotalPages,
	})
}

// GetNHIAAccreditedHealthMaintenanceOrganisation handles GET /v1/healthcare/nhia-accredited-health-maintenance-organisations/{organisation_id}.
func (h *NHIAAccreditedHealthMaintenanceOrganisationHandler) GetNHIAAccreditedHealthMaintenanceOrganisation(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	id, err := validators.ValidateNHIAAccreditedHealthMaintenanceOrganisationID("organisation_id", r.PathValue("organisation_id"))
	if err != nil {
		h.writeValidation(w, requestID, err)
		return
	}
	record, err := h.service.GetNHIAAccreditedHealthMaintenanceOrganisation(r.Context(), id)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Success(w, http.StatusOK, record)
}

func (h *NHIAAccreditedHealthMaintenanceOrganisationHandler) writeValidation(w http.ResponseWriter, requestID string, err error) {
	validationErr, ok := validationErrorsFrom(err)
	if !ok {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.ValidationBadRequest(w, requestID, validationErrorsToResponse(validationErr))
}
