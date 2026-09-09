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

type medicalLaboratoryAccreditationService interface {
	ListMedicalLaboratoryAccreditations(context.Context, interfaces.MedicalLaboratoryAccreditationQuery) (interfaces.MedicalLaboratoryAccreditationListResult, error)
	GetMedicalLaboratoryAccreditation(context.Context, string) (models.MedicalLaboratoryAccreditation, error)
}

// MedicalLaboratoryAccreditationHandler serves the paginated public medical laboratory accreditation contract.
type MedicalLaboratoryAccreditationHandler struct {
	service medicalLaboratoryAccreditationService
}

// NewMedicalLaboratoryAccreditationHandler constructs a healthcare handler with a narrow service dependency.
func NewMedicalLaboratoryAccreditationHandler(service medicalLaboratoryAccreditationService) (*MedicalLaboratoryAccreditationHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("medical laboratory accreditation service is required")
	}
	return &MedicalLaboratoryAccreditationHandler{service: service}, nil
}

// ListMedicalLaboratoryAccreditations handles GET /v1/healthcare/medical-laboratory-accreditations.
func (h *MedicalLaboratoryAccreditationHandler) ListMedicalLaboratoryAccreditations(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateMedicalLaboratoryAccreditationListQuery(r.URL.Query())
	if err != nil {
		h.writeHealthcareValidation(w, requestID, err)
		return
	}
	result, err := h.service.ListMedicalLaboratoryAccreditations(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.PaginatedPageSize(w, http.StatusOK, result.Records, response.PageSizePaginationMeta{
		Page: result.Page, PageSize: result.PageSize, Total: int64(result.Total), TotalPages: result.TotalPages,
	})
}

// GetMedicalLaboratoryAccreditation handles GET /v1/healthcare/medical-laboratory-accreditations/{accreditation_id}.
func (h *MedicalLaboratoryAccreditationHandler) GetMedicalLaboratoryAccreditation(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	id, err := validators.ValidateMedicalLaboratoryAccreditationID("accreditation_id", r.PathValue("accreditation_id"))
	if err != nil {
		h.writeHealthcareValidation(w, requestID, err)
		return
	}
	facility, err := h.service.GetMedicalLaboratoryAccreditation(r.Context(), id)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Success(w, http.StatusOK, facility)
}

func (h *MedicalLaboratoryAccreditationHandler) writeHealthcareValidation(w http.ResponseWriter, requestID string, err error) {
	validationErr, ok := validationErrorsFrom(err)
	if !ok {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.ValidationBadRequest(w, requestID, validationErrorsToResponse(validationErr))
}
