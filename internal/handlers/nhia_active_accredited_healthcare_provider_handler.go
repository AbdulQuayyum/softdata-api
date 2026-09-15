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

type nhiaActiveAccreditedHealthcareProviderService interface {
	ListNHIAActiveAccreditedHealthcareProviders(context.Context, interfaces.NHIAActiveAccreditedHealthcareProviderQuery) (interfaces.NHIAActiveAccreditedHealthcareProviderListResult, error)
	GetNHIAActiveAccreditedHealthcareProvider(context.Context, string) (models.NHIAActiveAccreditedHealthcareProvider, error)
}

type NHIAActiveAccreditedHealthcareProviderHandler struct {
	service nhiaActiveAccreditedHealthcareProviderService
}

func NewNHIAActiveAccreditedHealthcareProviderHandler(service nhiaActiveAccreditedHealthcareProviderService) (*NHIAActiveAccreditedHealthcareProviderHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("nhia active accredited healthcare provider service is required")
	}
	return &NHIAActiveAccreditedHealthcareProviderHandler{service: service}, nil
}

func (h *NHIAActiveAccreditedHealthcareProviderHandler) ListNHIAActiveAccreditedHealthcareProviders(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateNHIAActiveAccreditedHealthcareProviderListQuery(r.URL.Query())
	if err != nil {
		h.writeValidation(w, requestID, err)
		return
	}
	result, err := h.service.ListNHIAActiveAccreditedHealthcareProviders(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.PaginatedPageSize(w, http.StatusOK, result.Records, response.PageSizePaginationMeta{
		Page: result.Page, PageSize: result.PageSize, Total: int64(result.Total), TotalPages: result.TotalPages,
	})
}

func (h *NHIAActiveAccreditedHealthcareProviderHandler) GetNHIAActiveAccreditedHealthcareProvider(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	id, err := validators.ValidateNHIAActiveAccreditedHealthcareProviderID("provider_id", r.PathValue("provider_id"))
	if err != nil {
		h.writeValidation(w, requestID, err)
		return
	}
	record, err := h.service.GetNHIAActiveAccreditedHealthcareProvider(r.Context(), id)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Success(w, http.StatusOK, record)
}

func (h *NHIAActiveAccreditedHealthcareProviderHandler) writeValidation(w http.ResponseWriter, requestID string, err error) {
	validationErr, ok := validationErrorsFrom(err)
	if !ok {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.ValidationBadRequest(w, requestID, validationErrorsToResponse(validationErr))
}
