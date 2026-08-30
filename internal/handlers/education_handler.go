package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
	"github.com/AbdulQuayyum/softdata-api/internal/validators"
)

type educationService interface {
	ListUniversities(context.Context, services.UniversityListInput) ([]models.University, error)
	GetUniversity(context.Context, string) (models.University, error)
}

// EducationHandler serves public university endpoints.
type EducationHandler struct {
	service educationService
}

// NewEducationHandler constructs an education handler with its narrow service dependency.
func NewEducationHandler(service educationService) (*EducationHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("education service is required")
	}
	return &EducationHandler{service: service}, nil
}

// ListUniversities handles GET /v1/education/universities.
func (h *EducationHandler) ListUniversities(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateUniversityListQuery(r.URL.Query())
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	input := services.UniversityListInput{}
	if query.OwnershipType != nil {
		input.OwnershipType = *query.OwnershipType
	}
	if query.StateID != nil {
		input.StateID = *query.StateID
	}

	universities, err := h.service.ListUniversities(r.Context(), input)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, universities)
}

// GetUniversity handles GET /v1/education/universities/{university_id}.
func (h *EducationHandler) GetUniversity(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	universityID, err := validators.ValidateUniversityID(r.PathValue("university_id"))
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	university, err := h.service.GetUniversity(r.Context(), universityID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, university)
}
