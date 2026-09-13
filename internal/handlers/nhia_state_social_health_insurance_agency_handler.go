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

type nhiaStateSocialHealthInsuranceAgencyService interface {
	ListNHIAStateSocialHealthInsuranceAgencies(context.Context, interfaces.NHIAStateSocialHealthInsuranceAgencyQuery) (interfaces.NHIAStateSocialHealthInsuranceAgencyListResult, error)
	GetNHIAStateSocialHealthInsuranceAgency(context.Context, string) (models.NHIAStateSocialHealthInsuranceAgency, error)
}

type NHIAStateSocialHealthInsuranceAgencyHandler struct {
	service nhiaStateSocialHealthInsuranceAgencyService
}

func NewNHIAStateSocialHealthInsuranceAgencyHandler(service nhiaStateSocialHealthInsuranceAgencyService) (*NHIAStateSocialHealthInsuranceAgencyHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("nhia state social health insurance agency service is required")
	}
	return &NHIAStateSocialHealthInsuranceAgencyHandler{service: service}, nil
}

func (h *NHIAStateSocialHealthInsuranceAgencyHandler) ListNHIAStateSocialHealthInsuranceAgencies(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateNHIAStateSocialHealthInsuranceAgencyListQuery(r.URL.Query())
	if err != nil {
		h.writeValidation(w, requestID, err)
		return
	}
	result, err := h.service.ListNHIAStateSocialHealthInsuranceAgencies(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.PaginatedPageSize(w, http.StatusOK, result.Records, response.PageSizePaginationMeta{
		Page: result.Page, PageSize: result.PageSize, Total: int64(result.Total), TotalPages: result.TotalPages,
	})
}

func (h *NHIAStateSocialHealthInsuranceAgencyHandler) GetNHIAStateSocialHealthInsuranceAgency(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	id, err := validators.ValidateNHIAStateSocialHealthInsuranceAgencyID("agency_id", r.PathValue("agency_id"))
	if err != nil {
		h.writeValidation(w, requestID, err)
		return
	}
	record, err := h.service.GetNHIAStateSocialHealthInsuranceAgency(r.Context(), id)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Success(w, http.StatusOK, record)
}

func (h *NHIAStateSocialHealthInsuranceAgencyHandler) writeValidation(w http.ResponseWriter, requestID string, err error) {
	validationErr, ok := validationErrorsFrom(err)
	if !ok {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.ValidationBadRequest(w, requestID, validationErrorsToResponse(validationErr))
}
