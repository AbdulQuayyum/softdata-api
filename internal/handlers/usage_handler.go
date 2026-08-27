package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
	"github.com/AbdulQuayyum/softdata-api/internal/validators"
)

type usageService interface {
	GetUsageSummary(context.Context, string, *string, time.Time, time.Time) (models.UsageSummaryReportResponse, error)
	GetUsageHistory(context.Context, string, time.Time, time.Time) ([]models.UsageDailyResponse, error)
	GetAPIKeyUsageHistory(context.Context, string, string, time.Time, time.Time) ([]models.UsageDailyResponse, error)
	ListEndpointUsage(context.Context, string, time.Time, time.Time) ([]models.EndpointUsageResponse, error)
	ListAPIKeyEndpointUsage(context.Context, string, string, time.Time, time.Time) ([]models.EndpointUsageResponse, error)
	GetDatasetGroupUsage(context.Context, string, time.Time, time.Time) ([]models.DatasetGroupUsageResponse, error)
	GetAPIKeyDatasetGroupUsage(context.Context, string, string, time.Time, time.Time) ([]models.DatasetGroupUsageResponse, error)
}

type UsageHandler struct {
	service usageService
	now     func() time.Time
}

func NewUsageHandler(service usageService) (*UsageHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("usage service is required")
	}

	return &UsageHandler{
		service: service,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}, nil
}

func (h *UsageHandler) UsageSummary(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	identity, ok := middlewares.AccountIdentityFromContext(r.Context())
	if !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	query, err := h.usageQuery(r)
	if err != nil {
		h.writeUsageQueryError(w, requestID, err)
		return
	}

	summary, err := h.service.GetUsageSummary(r.Context(), identity.AccountID, queryAPIKeyIDPtr(query.APIKeyID), query.Start, query.End)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, summary)
}

func (h *UsageHandler) UsageHistory(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	identity, ok := middlewares.AccountIdentityFromContext(r.Context())
	if !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	query, err := h.usageQuery(r)
	if err != nil {
		h.writeUsageQueryError(w, requestID, err)
		return
	}

	var history []models.UsageDailyResponse
	if query.APIKeyID == "" {
		history, err = h.service.GetUsageHistory(r.Context(), identity.AccountID, query.Start, query.End)
	} else {
		history, err = h.service.GetAPIKeyUsageHistory(r.Context(), identity.AccountID, query.APIKeyID, query.Start, query.End)
	}
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, history)
}

func (h *UsageHandler) EndpointUsage(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	identity, ok := middlewares.AccountIdentityFromContext(r.Context())
	if !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	query, err := h.usageQuery(r)
	if err != nil {
		h.writeUsageQueryError(w, requestID, err)
		return
	}

	var endpoints []models.EndpointUsageResponse
	if query.APIKeyID == "" {
		endpoints, err = h.service.ListEndpointUsage(r.Context(), identity.AccountID, query.Start, query.End)
	} else {
		endpoints, err = h.service.ListAPIKeyEndpointUsage(r.Context(), identity.AccountID, query.APIKeyID, query.Start, query.End)
	}
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, endpoints)
}

func (h *UsageHandler) DatasetGroupUsage(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	identity, ok := middlewares.AccountIdentityFromContext(r.Context())
	if !ok {
		_ = response.Error(w, services.ErrInvalidCredentials, requestID)
		return
	}

	query, err := h.usageQuery(r)
	if err != nil {
		h.writeUsageQueryError(w, requestID, err)
		return
	}

	var groups []models.DatasetGroupUsageResponse
	if query.APIKeyID == "" {
		groups, err = h.service.GetDatasetGroupUsage(r.Context(), identity.AccountID, query.Start, query.End)
	} else {
		groups, err = h.service.GetAPIKeyDatasetGroupUsage(r.Context(), identity.AccountID, query.APIKeyID, query.Start, query.End)
	}
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, groups)
}

func (h *UsageHandler) usageQuery(r *http.Request) (validators.UsageQuery, error) {
	now := time.Now().UTC()
	if h != nil && h.now != nil {
		now = h.now().UTC()
	}
	return validators.ValidateUsageQuery(r.URL.Query(), now)
}

func (h *UsageHandler) writeUsageQueryError(w http.ResponseWriter, requestID string, err error) {
	if validationErr, ok := validationErrorsFrom(err); ok {
		_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
		return
	}
	_ = response.Error(w, err, requestID)
}

func queryAPIKeyIDPtr(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}
