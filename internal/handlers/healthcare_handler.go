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

type healthFacilityService interface {
	ListHealthFacilities(context.Context, interfaces.HealthFacilityQuery) (interfaces.HealthFacilityListResult, error)
	GetHealthFacility(context.Context, string) (models.HealthFacility, error)
}

// HealthFacilityHandler serves the paginated public health-facility contract.
type HealthFacilityHandler struct {
	service healthFacilityService
}

// NewHealthFacilityHandler constructs a healthcare handler with a narrow service dependency.
func NewHealthFacilityHandler(service healthFacilityService) (*HealthFacilityHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("health facility service is required")
	}
	return &HealthFacilityHandler{service: service}, nil
}

// ListHealthFacilities handles GET /v1/healthcare/health-facilities.
func (h *HealthFacilityHandler) ListHealthFacilities(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateHealthFacilityListQuery(r.URL.Query())
	if err != nil {
		h.writeHealthcareValidation(w, requestID, err)
		return
	}
	result, err := h.service.ListHealthFacilities(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Paginated(w, http.StatusOK, result.Facilities, response.PaginationMeta{
		Page: result.Page, Limit: result.PageSize, Total: int64(result.Total), TotalPages: result.TotalPages,
	})
}

// GetHealthFacility handles GET /v1/healthcare/health-facilities/{facility_id}.
func (h *HealthFacilityHandler) GetHealthFacility(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	id, err := validators.ValidateHealthFacilityID("facility_id", r.PathValue("facility_id"))
	if err != nil {
		h.writeHealthcareValidation(w, requestID, err)
		return
	}
	facility, err := h.service.GetHealthFacility(r.Context(), id)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Success(w, http.StatusOK, facility)
}

func (h *HealthFacilityHandler) writeHealthcareValidation(w http.ResponseWriter, requestID string, err error) {
	validationErr, ok := validationErrorsFrom(err)
	if !ok {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.ValidationBadRequest(w, requestID, validationErrorsToResponse(validationErr))
}
