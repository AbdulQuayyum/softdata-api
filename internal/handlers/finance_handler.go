package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/validators"
)

type financeService interface {
	ListPaymentServiceProviders(context.Context) ([]models.PaymentServiceProvider, error)
	ListPaymentServiceProvidersByType(context.Context, string) ([]models.PaymentServiceProvider, error)
	GetPaymentServiceProvider(context.Context, string) (models.PaymentServiceProvider, error)
}

// FinanceHandler serves public payment-service-provider endpoints.
type FinanceHandler struct {
	service financeService
}

// NewFinanceHandler constructs a finance handler with its narrow service dependency.
func NewFinanceHandler(service financeService) (*FinanceHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("finance service is required")
	}
	return &FinanceHandler{service: service}, nil
}

// ListPaymentServiceProviders handles GET /v1/finance/payment-service-providers.
func (h *FinanceHandler) ListPaymentServiceProviders(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidatePaymentServiceProviderListQuery(r.URL.Query())
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	var providers []models.PaymentServiceProvider
	if query.InstitutionType == nil {
		providers, err = h.service.ListPaymentServiceProviders(r.Context())
	} else {
		providers, err = h.service.ListPaymentServiceProvidersByType(r.Context(), *query.InstitutionType)
	}
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, providers)
}

// GetPaymentServiceProvider handles GET /v1/finance/payment-service-providers/{provider_id}.
func (h *FinanceHandler) GetPaymentServiceProvider(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	providerID, err := validators.ValidatePaymentServiceProviderID(r.PathValue("provider_id"))
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	provider, err := h.service.GetPaymentServiceProvider(r.Context(), providerID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, provider)
}
