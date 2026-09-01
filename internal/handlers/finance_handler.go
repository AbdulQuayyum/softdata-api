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

type financeService interface {
	ListPaymentServiceProviders(context.Context) ([]models.PaymentServiceProvider, error)
	ListPaymentServiceProvidersByType(context.Context, string) ([]models.PaymentServiceProvider, error)
	GetPaymentServiceProvider(context.Context, string) (models.PaymentServiceProvider, error)
	ListInternationalMoneyTransferOperators(context.Context) ([]models.InternationalMoneyTransferOperator, error)
	GetInternationalMoneyTransferOperator(context.Context, string) (models.InternationalMoneyTransferOperator, error)
	ListCurrencies(context.Context, services.CurrencyListInput) ([]models.Currency, error)
	GetCurrency(context.Context, string) (models.Currency, error)
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

// ListInternationalMoneyTransferOperators handles GET /v1/finance/international-money-transfer-operators.
func (h *FinanceHandler) ListInternationalMoneyTransferOperators(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	operators, err := h.service.ListInternationalMoneyTransferOperators(r.Context())
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, operators)
}

// GetInternationalMoneyTransferOperator handles GET /v1/finance/international-money-transfer-operators/{operator_id}.
func (h *FinanceHandler) GetInternationalMoneyTransferOperator(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	operatorID, err := validators.ValidateInternationalMoneyTransferOperatorID(r.PathValue("operator_id"))
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	operator, err := h.service.GetInternationalMoneyTransferOperator(r.Context(), operatorID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, operator)
}

// ListCurrencies handles GET /v1/finance/currencies.
func (h *FinanceHandler) ListCurrencies(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateCurrencyListQuery(r.URL.Query())
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	currencies, err := h.service.ListCurrencies(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, currencies)
}

// GetCurrency handles GET /v1/finance/currencies/{currency_id}.
func (h *FinanceHandler) GetCurrency(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	currencyID, err := validators.ValidateCurrencyID(r.PathValue("currency_id"))
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	currency, err := h.service.GetCurrency(r.Context(), currencyID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, currency)
}
