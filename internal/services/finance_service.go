package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var financePaymentServiceProviderIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
var financeInternationalMoneyTransferOperatorIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var financeCurrencyIDPattern = regexp.MustCompile(`^[a-z]{3}$`)

var allowedFinanceInstitutionTypes = map[string]struct{}{
	"mobile_money_operator":               {},
	"switching_and_processing_company":    {},
	"payment_solution_service_provider":   {},
	"payment_terminal_service_provider":   {},
	"super_agent":                         {},
	"payment_service_holding_company":     {},
	"payment_terminal_service_aggregator": {},
}

// FinanceService provides payment-service-provider lookups over the finance repository.
type FinanceService struct {
	repository interfaces.FinanceRepository
}

func NewFinanceService(repository interfaces.FinanceRepository) (*FinanceService, error) {
	if repository == nil {
		return nil, fmt.Errorf("finance repository is required")
	}
	return &FinanceService{repository: repository}, nil
}

func (s *FinanceService) ListPaymentServiceProviders(ctx context.Context) ([]models.PaymentServiceProvider, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	providers, err := s.repository.ListPaymentServiceProviders(ctx)
	if err != nil {
		return nil, translateFinanceServiceError("list payment service providers", err)
	}
	return clonePaymentServiceProviderList(providers), nil
}

func (s *FinanceService) ListPaymentServiceProvidersByType(ctx context.Context, institutionType string) ([]models.PaymentServiceProvider, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	normalizedType, err := normalizeFinanceInstitutionType(institutionType)
	if err != nil {
		return nil, err
	}

	providers, err := s.repository.ListPaymentServiceProvidersByType(ctx, normalizedType)
	if err != nil {
		return nil, translateFinanceServiceError("list payment service providers by type", err)
	}
	return clonePaymentServiceProviderList(providers), nil
}

func (s *FinanceService) GetPaymentServiceProvider(ctx context.Context, providerID string) (models.PaymentServiceProvider, error) {
	if err := ctx.Err(); err != nil {
		return models.PaymentServiceProvider{}, err
	}

	normalizedID, err := normalizeFinancePaymentServiceProviderID(providerID)
	if err != nil {
		return models.PaymentServiceProvider{}, err
	}

	provider, err := s.repository.GetPaymentServiceProvider(ctx, normalizedID)
	if err != nil {
		return models.PaymentServiceProvider{}, translateFinanceServiceLookupError(err)
	}
	return provider, nil
}

func (s *FinanceService) ListInternationalMoneyTransferOperators(ctx context.Context) ([]models.InternationalMoneyTransferOperator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	operators, err := s.repository.ListInternationalMoneyTransferOperators(ctx)
	if err != nil {
		return nil, translateFinanceServiceError("list international money transfer operators", err)
	}
	return cloneInternationalMoneyTransferOperatorList(operators), nil
}

func (s *FinanceService) GetInternationalMoneyTransferOperator(ctx context.Context, operatorID string) (models.InternationalMoneyTransferOperator, error) {
	if err := ctx.Err(); err != nil {
		return models.InternationalMoneyTransferOperator{}, err
	}

	normalizedID, err := normalizeFinanceInternationalMoneyTransferOperatorID(operatorID)
	if err != nil {
		return models.InternationalMoneyTransferOperator{}, err
	}

	operator, err := s.repository.GetInternationalMoneyTransferOperator(ctx, normalizedID)
	if err != nil {
		return models.InternationalMoneyTransferOperator{}, translateFinanceIMTOLookupError(err)
	}
	return operator, nil
}

func (s *FinanceService) ListCurrencies(ctx context.Context) ([]models.Currency, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	currencies, err := s.repository.ListCurrencies(ctx)
	if err != nil {
		return nil, translateFinanceCurrencyServiceError("list currencies", err)
	}
	return cloneCurrencyList(currencies), nil
}

func (s *FinanceService) GetCurrency(ctx context.Context, currencyID string) (models.Currency, error) {
	if err := ctx.Err(); err != nil {
		return models.Currency{}, err
	}

	normalizedID, err := normalizeFinanceCurrencyID(currencyID)
	if err != nil {
		return models.Currency{}, err
	}

	currency, err := s.repository.GetCurrency(ctx, normalizedID)
	if err != nil {
		return models.Currency{}, translateFinanceCurrencyLookupError(err)
	}
	return currency, nil
}

func cloneInternationalMoneyTransferOperatorList(operators []models.InternationalMoneyTransferOperator) []models.InternationalMoneyTransferOperator {
	if len(operators) == 0 {
		return make([]models.InternationalMoneyTransferOperator, 0)
	}
	cloned := make([]models.InternationalMoneyTransferOperator, len(operators))
	copy(cloned, operators)
	return cloned
}

func clonePaymentServiceProviderList(providers []models.PaymentServiceProvider) []models.PaymentServiceProvider {
	if len(providers) == 0 {
		return make([]models.PaymentServiceProvider, 0)
	}
	cloned := make([]models.PaymentServiceProvider, len(providers))
	copy(cloned, providers)
	return cloned
}

func cloneCurrencyList(currencies []models.Currency) []models.Currency {
	if len(currencies) == 0 {
		return make([]models.Currency, 0)
	}
	cloned := make([]models.Currency, len(currencies))
	copy(cloned, currencies)
	return cloned
}

func normalizeFinancePaymentServiceProviderID(providerID string) (string, error) {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" || !financePaymentServiceProviderIDPattern.MatchString(providerID) {
		return "", ErrInvalidPaymentServiceProviderID
	}
	return providerID, nil
}

func normalizeFinanceInternationalMoneyTransferOperatorID(operatorID string) (string, error) {
	operatorID = strings.TrimSpace(operatorID)
	if operatorID == "" || !financeInternationalMoneyTransferOperatorIDPattern.MatchString(operatorID) {
		return "", ErrInvalidInternationalMoneyTransferOperatorID
	}
	return operatorID, nil
}

func normalizeFinanceCurrencyID(currencyID string) (string, error) {
	currencyID = strings.TrimSpace(currencyID)
	if currencyID == "" || !financeCurrencyIDPattern.MatchString(currencyID) {
		return "", ErrInvalidCurrencyID
	}
	return currencyID, nil
}

func normalizeFinanceInstitutionType(institutionType string) (string, error) {
	institutionType = strings.TrimSpace(institutionType)
	if institutionType == "" {
		return "", ErrInvalidPaymentServiceProviderType
	}
	if _, ok := allowedFinanceInstitutionTypes[institutionType]; !ok {
		return "", ErrInvalidPaymentServiceProviderType
	}
	return institutionType, nil
}

func translateFinanceServiceLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrPaymentServiceProviderNotFound):
		return ErrPaymentServiceProviderNotFound
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("get payment service provider: repository unavailable")
	default:
		return fmt.Errorf("get payment service provider: repository unavailable")
	}
}

func translateFinanceIMTOLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrInternationalMoneyTransferOperatorNotFound):
		return ErrInternationalMoneyTransferOperatorNotFound
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("get international money transfer operator: repository unavailable")
	default:
		return fmt.Errorf("get international money transfer operator: repository unavailable")
	}
}

func translateFinanceCurrencyLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrCurrencyNotFound):
		return ErrCurrencyNotFound
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("get currency: repository unavailable")
	default:
		return fmt.Errorf("get currency: repository unavailable")
	}
}

func translateFinanceCurrencyServiceError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("%s: repository unavailable", op)
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}

func translateFinanceServiceError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("%s: repository unavailable", op)
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}
