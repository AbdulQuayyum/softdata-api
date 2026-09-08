package handlers

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/validators"
)

type extendedEducationService interface {
	ListPolytechnics(context.Context) ([]models.Polytechnic, error)
	GetPolytechnic(context.Context, string) (models.Polytechnic, error)
	ListMonotechnics(context.Context) ([]models.Monotechnic, error)
	GetMonotechnic(context.Context, string) (models.Monotechnic, error)
	ListCollegesOfAgriculture(context.Context) ([]models.CollegeOfAgriculture, error)
	GetCollegeOfAgriculture(context.Context, string) (models.CollegeOfAgriculture, error)
	ListCollegesOfHealthSciencesAndTechnology(context.Context) ([]models.CollegeOfHealthSciencesAndTechnology, error)
	GetCollegeOfHealthSciencesAndTechnology(context.Context, string) (models.CollegeOfHealthSciencesAndTechnology, error)
	ListCollegesOfNursingAndMidwifery(context.Context) ([]models.CollegeOfNursingAndMidwifery, error)
	GetCollegeOfNursingAndMidwifery(context.Context, string) (models.CollegeOfNursingAndMidwifery, error)
	ListVocationalEnterpriseInstitutions(context.Context) ([]models.VocationalEnterpriseInstitution, error)
	GetVocationalEnterpriseInstitution(context.Context, string) (models.VocationalEnterpriseInstitution, error)
	ListTechnicalColleges(context.Context) ([]models.TechnicalCollege, error)
	GetTechnicalCollege(context.Context, string) (models.TechnicalCollege, error)
	ListPrimaryAndSecondarySchools(context.Context, interfaces.PrimaryAndSecondarySchoolQuery) (interfaces.PrimaryAndSecondarySchoolListResult, error)
	GetPrimaryAndSecondarySchool(context.Context, string) (models.PrimaryAndSecondarySchool, error)
}

func (h *EducationHandler) extendedService() (extendedEducationService, error) {
	service, ok := h.service.(extendedEducationService)
	if !ok {
		return nil, fmt.Errorf("education service unavailable")
	}
	return service, nil
}

func (h *EducationHandler) listEducationInstitutions(w http.ResponseWriter, r *http.Request, list func(context.Context) (any, error)) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	rows, err := list(r.Context())
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	value := reflect.ValueOf(rows)
	if value.IsValid() && value.Kind() == reflect.Slice && value.IsNil() {
		rows = reflect.MakeSlice(value.Type(), 0, 0).Interface()
	}
	_ = response.JSON(w, http.StatusOK, struct {
		Success bool `json:"success"`
		Data    any  `json:"data"`
	}{true, rows})
}

func (h *EducationHandler) getEducationInstitution(w http.ResponseWriter, r *http.Request, field, value string, get func(context.Context, string) (any, error)) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	id, err := validators.ValidateEducationInstitutionID(field, value)
	if err != nil {
		h.writeValidationOrError(w, requestID, err)
		return
	}
	row, err := get(r.Context(), id)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Success(w, http.StatusOK, row)
}

func (h *EducationHandler) writeValidationOrError(w http.ResponseWriter, requestID string, err error) {
	if validationErr, ok := validationErrorsFrom(err); ok {
		_ = response.ValidationBadRequest(w, requestID, validationErrorsToResponse(validationErr))
		return
	}
	_ = response.Error(w, err, requestID)
}

func (h *EducationHandler) ListPolytechnics(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.listEducationInstitutions(w, r, func(ctx context.Context) (any, error) { return s.ListPolytechnics(ctx) })
}
func (h *EducationHandler) GetPolytechnic(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.getEducationInstitution(w, r, "institution_id", r.PathValue("institution_id"), func(ctx context.Context, id string) (any, error) { return s.GetPolytechnic(ctx, id) })
}
func (h *EducationHandler) ListMonotechnics(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.listEducationInstitutions(w, r, func(ctx context.Context) (any, error) { return s.ListMonotechnics(ctx) })
}
func (h *EducationHandler) GetMonotechnic(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.getEducationInstitution(w, r, "institution_id", r.PathValue("institution_id"), func(ctx context.Context, id string) (any, error) { return s.GetMonotechnic(ctx, id) })
}
func (h *EducationHandler) ListCollegesOfAgriculture(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.listEducationInstitutions(w, r, func(ctx context.Context) (any, error) { return s.ListCollegesOfAgriculture(ctx) })
}
func (h *EducationHandler) GetCollegeOfAgriculture(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.getEducationInstitution(w, r, "institution_id", r.PathValue("institution_id"), func(ctx context.Context, id string) (any, error) { return s.GetCollegeOfAgriculture(ctx, id) })
}
func (h *EducationHandler) ListCollegesOfHealthSciencesAndTechnology(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.listEducationInstitutions(w, r, func(ctx context.Context) (any, error) { return s.ListCollegesOfHealthSciencesAndTechnology(ctx) })
}
func (h *EducationHandler) GetCollegeOfHealthSciencesAndTechnology(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.getEducationInstitution(w, r, "institution_id", r.PathValue("institution_id"), func(ctx context.Context, id string) (any, error) {
		return s.GetCollegeOfHealthSciencesAndTechnology(ctx, id)
	})
}
func (h *EducationHandler) ListCollegesOfNursingAndMidwifery(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.listEducationInstitutions(w, r, func(ctx context.Context) (any, error) { return s.ListCollegesOfNursingAndMidwifery(ctx) })
}
func (h *EducationHandler) GetCollegeOfNursingAndMidwifery(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.getEducationInstitution(w, r, "institution_id", r.PathValue("institution_id"), func(ctx context.Context, id string) (any, error) { return s.GetCollegeOfNursingAndMidwifery(ctx, id) })
}
func (h *EducationHandler) ListVocationalEnterpriseInstitutions(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.listEducationInstitutions(w, r, func(ctx context.Context) (any, error) { return s.ListVocationalEnterpriseInstitutions(ctx) })
}
func (h *EducationHandler) GetVocationalEnterpriseInstitution(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.getEducationInstitution(w, r, "institution_id", r.PathValue("institution_id"), func(ctx context.Context, id string) (any, error) {
		return s.GetVocationalEnterpriseInstitution(ctx, id)
	})
}
func (h *EducationHandler) ListTechnicalColleges(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.listEducationInstitutions(w, r, func(ctx context.Context) (any, error) { return s.ListTechnicalColleges(ctx) })
}
func (h *EducationHandler) GetTechnicalCollege(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.getEducationInstitution(w, r, "institution_id", r.PathValue("institution_id"), func(ctx context.Context, id string) (any, error) { return s.GetTechnicalCollege(ctx, id) })
}

func (h *EducationHandler) ListPrimaryAndSecondarySchools(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateSchoolListQuery(r.URL.Query())
	if err != nil {
		h.writeValidationOrError(w, requestID, err)
		return
	}
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	result, err := s.ListPrimaryAndSecondarySchools(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Paginated(w, http.StatusOK, result.Schools, response.PaginationMeta{Page: result.Page, Limit: result.PageSize, Total: int64(result.Total), TotalPages: result.TotalPages})
}

func (h *EducationHandler) GetPrimaryAndSecondarySchool(w http.ResponseWriter, r *http.Request) {
	s, err := h.extendedService()
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	h.getEducationInstitution(w, r, "school_id", r.PathValue("school_id"), func(ctx context.Context, id string) (any, error) { return s.GetPrimaryAndSecondarySchool(ctx, id) })
}
