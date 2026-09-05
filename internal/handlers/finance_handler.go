package handlers

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

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
	ListCommercialBanks(context.Context) ([]models.CommercialBank, error)
	GetCommercialBank(context.Context, string) (models.CommercialBank, error)
	ListNonInterestFinancialInstitutions(context.Context) ([]models.NonInterestInstitution, error)
	GetNonInterestFinancialInstitution(context.Context, string) (models.NonInterestInstitution, error)
	ListMerchantBanks(context.Context) ([]models.MerchantBank, error)
	GetMerchantBank(context.Context, string) (models.MerchantBank, error)
	ListPaymentServiceBanks(context.Context) ([]models.PaymentServiceBank, error)
	GetPaymentServiceBank(context.Context, string) (models.PaymentServiceBank, error)
	ListFinancialHoldingCompanies(context.Context) ([]models.FinancialHoldingCompany, error)
	GetFinancialHoldingCompany(context.Context, string) (models.FinancialHoldingCompany, error)
	ListDevelopmentFinanceInstitutions(context.Context) ([]models.DevelopmentFinanceInstitution, error)
	GetDevelopmentFinanceInstitution(context.Context, string) (models.DevelopmentFinanceInstitution, error)
	ListPrimaryMortgageInstitutions(context.Context) ([]models.PrimaryMortgageInstitution, error)
	GetPrimaryMortgageInstitution(context.Context, string) (models.PrimaryMortgageInstitution, error)
}

func writeFinanceList[T any](h *FinanceHandler, w http.ResponseWriter, r *http.Request, list func(context.Context) ([]T, error)) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	rows, err := list(r.Context())
	if err != nil {
		_ = response.Error(w, err, requestIDFromContext(r.Context()))
		return
	}
	_ = response.List(w, http.StatusOK, qualifyLogoURLs(h, rows))
}

func writeFinanceDetail[T any](h *FinanceHandler, w http.ResponseWriter, r *http.Request, field string, validate func(string) error, get func(context.Context, string) (T, error)) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	id := r.PathValue(field)
	if err := validate(id); err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}
	row, err := get(r.Context(), id)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Success(w, http.StatusOK, qualifyLogoURL(h, row))
}

// ListNonInterestFinancialInstitutions handles GET /v1/finance/non-interest-financial-institutions.
func (h *FinanceHandler) ListNonInterestFinancialInstitutions(w http.ResponseWriter, r *http.Request) {
	writeFinanceList(h, w, r, h.service.ListNonInterestFinancialInstitutions)
}

// GetNonInterestFinancialInstitution handles GET /v1/finance/non-interest-financial-institutions/{institution_id}.
func (h *FinanceHandler) GetNonInterestFinancialInstitution(w http.ResponseWriter, r *http.Request) {
	writeFinanceDetail(h, w, r, "institution_id", validators.ValidateNonInterestFinancialInstitutionID, h.service.GetNonInterestFinancialInstitution)
}

// ListMerchantBanks handles GET /v1/finance/merchant-banks.
func (h *FinanceHandler) ListMerchantBanks(w http.ResponseWriter, r *http.Request) {
	writeFinanceList(h, w, r, h.service.ListMerchantBanks)
}

// GetMerchantBank handles GET /v1/finance/merchant-banks/{bank_id}.
func (h *FinanceHandler) GetMerchantBank(w http.ResponseWriter, r *http.Request) {
	writeFinanceDetail(h, w, r, "bank_id", validators.ValidateMerchantBankID, h.service.GetMerchantBank)
}

// ListPaymentServiceBanks handles GET /v1/finance/payment-service-banks.
func (h *FinanceHandler) ListPaymentServiceBanks(w http.ResponseWriter, r *http.Request) {
	writeFinanceList(h, w, r, h.service.ListPaymentServiceBanks)
}

// GetPaymentServiceBank handles GET /v1/finance/payment-service-banks/{bank_id}.
func (h *FinanceHandler) GetPaymentServiceBank(w http.ResponseWriter, r *http.Request) {
	writeFinanceDetail(h, w, r, "bank_id", validators.ValidatePaymentServiceBankID, h.service.GetPaymentServiceBank)
}

// ListFinancialHoldingCompanies handles GET /v1/finance/financial-holding-companies.
func (h *FinanceHandler) ListFinancialHoldingCompanies(w http.ResponseWriter, r *http.Request) {
	writeFinanceList(h, w, r, h.service.ListFinancialHoldingCompanies)
}

// GetFinancialHoldingCompany handles GET /v1/finance/financial-holding-companies/{company_id}.
func (h *FinanceHandler) GetFinancialHoldingCompany(w http.ResponseWriter, r *http.Request) {
	writeFinanceDetail(h, w, r, "company_id", validators.ValidateFinancialHoldingCompanyID, h.service.GetFinancialHoldingCompany)
}

// ListDevelopmentFinanceInstitutions handles GET /v1/finance/development-finance-institutions.
func (h *FinanceHandler) ListDevelopmentFinanceInstitutions(w http.ResponseWriter, r *http.Request) {
	writeFinanceList(h, w, r, h.service.ListDevelopmentFinanceInstitutions)
}

// GetDevelopmentFinanceInstitution handles GET /v1/finance/development-finance-institutions/{institution_id}.
func (h *FinanceHandler) GetDevelopmentFinanceInstitution(w http.ResponseWriter, r *http.Request) {
	writeFinanceDetail(h, w, r, "institution_id", validators.ValidateDevelopmentFinanceInstitutionID, h.service.GetDevelopmentFinanceInstitution)
}

// ListPrimaryMortgageInstitutions handles GET /v1/finance/primary-mortgage-institutions.
func (h *FinanceHandler) ListPrimaryMortgageInstitutions(w http.ResponseWriter, r *http.Request) {
	writeFinanceList(h, w, r, h.service.ListPrimaryMortgageInstitutions)
}

// GetPrimaryMortgageInstitution handles GET /v1/finance/primary-mortgage-institutions/{institution_id}.
func (h *FinanceHandler) GetPrimaryMortgageInstitution(w http.ResponseWriter, r *http.Request) {
	writeFinanceDetail(h, w, r, "institution_id", validators.ValidatePrimaryMortgageInstitutionID, h.service.GetPrimaryMortgageInstitution)
}

// ListCommercialBanks handles GET /v1/finance/commercial-banks.
func (h *FinanceHandler) ListCommercialBanks(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	banks, err := h.service.ListCommercialBanks(r.Context())
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.List(w, http.StatusOK, qualifyLogoURLs(h, banks))
}

// GetCommercialBank handles GET /v1/finance/commercial-banks/{bank_id}.
func (h *FinanceHandler) GetCommercialBank(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	requestID := requestIDFromContext(r.Context())
	bankID := r.PathValue("bank_id")
	if err := validators.ValidateCommercialBankID(bankID); err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}
	bank, err := h.service.GetCommercialBank(r.Context(), bankID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}
	_ = response.Success(w, http.StatusOK, qualifyLogoURL(h, bank))
}

// FinanceHandler serves public payment-service-provider endpoints.
type FinanceHandler struct {
	service      financeService
	publicAPIURL string
}

// NewFinanceHandler constructs a finance handler with its narrow service dependency.
func NewFinanceHandler(service financeService) (*FinanceHandler, error) {
	return NewFinanceHandlerWithPublicAPIURL(service, "")
}

// NewFinanceHandlerWithPublicAPIURL configures absolute asset URLs for public responses.
func NewFinanceHandlerWithPublicAPIURL(service financeService, publicAPIURL string) (*FinanceHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("finance service is required")
	}
	return &FinanceHandler{service: service, publicAPIURL: strings.TrimRight(strings.TrimSpace(publicAPIURL), "/")}, nil
}

func qualifyLogoURLs[T any](h *FinanceHandler, rows []T) []T {
	for i := range rows {
		rows[i] = qualifyLogoURL(h, rows[i])
	}
	return rows
}

func qualifyLogoURL[T any](h *FinanceHandler, row T) T {
	if h == nil || h.publicAPIURL == "" {
		return row
	}
	value := reflect.ValueOf(&row).Elem()
	if value.Kind() != reflect.Struct {
		return row
	}
	logo := value.FieldByName("LogoURL")
	if !logo.IsValid() || !logo.CanSet() || logo.Kind() != reflect.String {
		return row
	}
	path := logo.String()
	if strings.HasPrefix(path, "/") {
		logo.SetString(h.publicAPIURL + path)
	}
	return row
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
