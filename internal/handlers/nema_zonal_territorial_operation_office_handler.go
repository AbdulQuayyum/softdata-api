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

type nemaZonalTerritorialOperationOfficeService interface {
	ListNEMAZonalTerritorialOperationOffices(context.Context, interfaces.NEMAZonalTerritorialOperationOfficeQuery) (interfaces.NEMAZonalTerritorialOperationOfficeListResult, error)
	GetNEMAZonalTerritorialOperationOffice(context.Context, string) (models.NEMAZonalTerritorialOperationOffice, error)
}

type NEMAZonalTerritorialOperationOfficeHandler struct {
	service nemaZonalTerritorialOperationOfficeService
}

func NewNEMAZonalTerritorialOperationOfficeHandler(service nemaZonalTerritorialOperationOfficeService) (*NEMAZonalTerritorialOperationOfficeHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("nema zonal territorial operation office service is required")
	}
	return &NEMAZonalTerritorialOperationOfficeHandler{service: service}, nil
}

func (h *NEMAZonalTerritorialOperationOfficeHandler) ListNEMAZonalTerritorialOperationOffices(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateNEMAZonalTerritorialOperationOfficeListQuery(r.URL.Query())
	if err != nil {
		h.writeValidation(w, requestID, err)
		return
	}
	result, err := h.service.ListNEMAZonalTerritorialOperationOffices(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.PaginatedPageSize(w, http.StatusOK, result.Records, response.PageSizePaginationMeta{
		Page: result.Page, PageSize: result.PageSize, Total: int64(result.Total), TotalPages: result.TotalPages,
	})
}

func (h *NEMAZonalTerritorialOperationOfficeHandler) GetNEMAZonalTerritorialOperationOffice(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	id, err := validators.ValidateNEMAZonalTerritorialOperationOfficeID("office_id", r.PathValue("office_id"))
	if err != nil {
		h.writeValidation(w, requestID, err)
		return
	}
	record, err := h.service.GetNEMAZonalTerritorialOperationOffice(r.Context(), id)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Success(w, http.StatusOK, record)
}

func (h *NEMAZonalTerritorialOperationOfficeHandler) writeValidation(w http.ResponseWriter, requestID string, err error) {
	validationErr, ok := validationErrorsFrom(err)
	if !ok {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.ValidationBadRequest(w, requestID, validationErrorsToResponse(validationErr))
}
