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

type datasetService interface {
	ListDatasets(context.Context, string, int, int) (services.DatasetListResult, error)
	GetDataset(context.Context, string) (models.DatasetResponse, error)
	ListDatasetSources(context.Context, string) ([]models.DatasetSourceResponse, error)
	ListDatasetVersions(context.Context, string) ([]models.DatasetVersionResponse, error)
}

type DatasetHandler struct {
	service datasetService
}

func NewDatasetHandler(service datasetService) (*DatasetHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("dataset service is required")
	}

	return &DatasetHandler{service: service}, nil
}

func (h *DatasetHandler) ListDatasets(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	query, err := validateDatasetListQuery(r)
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	result, err := h.service.ListDatasets(r.Context(), query.Search, query.Page, query.Limit)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Paginated(w, http.StatusOK, result.Datasets, response.PaginationMeta{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *DatasetHandler) GetDataset(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	datasetID, err := validateDatasetPathValue(r)
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	dataset, err := h.service.GetDataset(r.Context(), datasetID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, dataset)
}

func (h *DatasetHandler) ListDatasetSources(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	datasetID, err := validateDatasetPathValue(r)
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	sources, err := h.service.ListDatasetSources(r.Context(), datasetID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, sources)
}

func (h *DatasetHandler) ListDatasetVersions(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	datasetID, err := validateDatasetPathValue(r)
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	versions, err := h.service.ListDatasetVersions(r.Context(), datasetID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, versions)
}

func validateDatasetPathValue(r *http.Request) (string, error) {
	datasetID, err := validators.ValidateDatasetKey(r.PathValue("dataset_id"))
	if err != nil {
		return "", err
	}
	return datasetID, nil
}

func validateDatasetListQuery(r *http.Request) (validators.DatasetListQuery, error) {
	values := r.URL.Query()
	var errs validators.ValidationErrors

	if len(values["search"]) > 1 {
		errs.Add("search", "malformed", "Search may be provided at most once.")
	}
	if len(values["page"]) > 1 {
		errs.Add("page", "malformed", "Page may be provided at most once.")
	}
	if len(values["limit"]) > 1 {
		errs.Add("limit", "malformed", "Limit may be provided at most once.")
	}
	if len(errs.Fields) > 0 {
		return validators.DatasetListQuery{}, errs
	}

	return validators.ValidateDatasetListQuery(values.Get("search"), values.Get("page"), values.Get("limit"))
}
