package services

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type financeRepositoryStub struct {
	listAllResult      []models.PaymentServiceProvider
	listByType         map[string][]models.PaymentServiceProvider
	getResultByID      map[string]models.PaymentServiceProvider
	listIMTOResult     []models.InternationalMoneyTransferOperator
	getIMTOResult      map[string]models.InternationalMoneyTransferOperator
	listCurrencyResult []models.Currency
	getCurrencyResult  map[string]models.Currency
	listAllErr         error
	listByTypeErr      error
	getErr             error
	listIMTOErr        error
	getIMTOErr         error
	listCurrencyErr    error
	getCurrencyErr     error
	listAllCalls       int
	listByTypeCall     int
	getCalls           int
	listIMTOCalls      int
	getIMTOCalls       int
	listCurrencyCalls  int
	getCurrencyCalls   int
	lastType           string
	lastID             string
	lastIMTOID         string
	lastCurrencyID     string
}

func (s *financeRepositoryStub) ListPaymentServiceProviders(context.Context) ([]models.PaymentServiceProvider, error) {
	s.listAllCalls++
	if s.listAllErr != nil {
		return nil, s.listAllErr
	}
	return clonePaymentServiceProviderList(s.listAllResult), nil
}

func (s *financeRepositoryStub) ListPaymentServiceProvidersByType(_ context.Context, institutionType string) ([]models.PaymentServiceProvider, error) {
	s.listByTypeCall++
	s.lastType = institutionType
	if s.listByTypeErr != nil {
		return nil, s.listByTypeErr
	}
	if s.listByType != nil {
		return clonePaymentServiceProviderList(s.listByType[institutionType]), nil
	}
	return nil, nil
}

func (s *financeRepositoryStub) GetPaymentServiceProvider(_ context.Context, providerID string) (models.PaymentServiceProvider, error) {
	s.getCalls++
	s.lastID = providerID
	if s.getErr != nil {
		return models.PaymentServiceProvider{}, s.getErr
	}
	if s.getResultByID != nil {
		if provider, ok := s.getResultByID[providerID]; ok {
			return provider, nil
		}
	}
	return models.PaymentServiceProvider{}, interfaces.ErrPaymentServiceProviderNotFound
}

func (s *financeRepositoryStub) ListInternationalMoneyTransferOperators(context.Context) ([]models.InternationalMoneyTransferOperator, error) {
	s.listIMTOCalls++
	if s.listIMTOErr != nil {
		return nil, s.listIMTOErr
	}
	return cloneInternationalMoneyTransferOperatorList(s.listIMTOResult), nil
}

func (s *financeRepositoryStub) GetInternationalMoneyTransferOperator(_ context.Context, operatorID string) (models.InternationalMoneyTransferOperator, error) {
	s.getIMTOCalls++
	s.lastIMTOID = operatorID
	if s.getIMTOErr != nil {
		return models.InternationalMoneyTransferOperator{}, s.getIMTOErr
	}
	if s.getIMTOResult != nil {
		if operator, ok := s.getIMTOResult[operatorID]; ok {
			return operator, nil
		}
	}
	return models.InternationalMoneyTransferOperator{}, interfaces.ErrInternationalMoneyTransferOperatorNotFound
}

func (s *financeRepositoryStub) ListCurrencies(context.Context) ([]models.Currency, error) {
	s.listCurrencyCalls++
	if s.listCurrencyErr != nil {
		return nil, s.listCurrencyErr
	}
	return cloneCurrencyList(s.listCurrencyResult), nil
}

func (s *financeRepositoryStub) GetCurrency(_ context.Context, currencyID string) (models.Currency, error) {
	s.getCurrencyCalls++
	s.lastCurrencyID = currencyID
	if s.getCurrencyErr != nil {
		return models.Currency{}, s.getCurrencyErr
	}
	if s.getCurrencyResult != nil {
		if currency, ok := s.getCurrencyResult[currencyID]; ok {
			return currency, nil
		}
	}
	return models.Currency{}, interfaces.ErrCurrencyNotFound
}

func TestFinanceServiceListAllAndTypeAndLookup(t *testing.T) {
	t.Parallel()

	all := []models.PaymentServiceProvider{
		{ID: "mobile-money-operator-abeg-technologies-limited", Name: "Abeg Technologies Limited", InstitutionType: "mobile_money_operator", CountryCode: "NG"},
		{ID: "mobile-money-operator-chams-mobile", Name: "Chams Mobile", InstitutionType: "mobile_money_operator", CountryCode: "NG"},
	}
	stub := &financeRepositoryStub{
		listAllResult: all,
		listByType: map[string][]models.PaymentServiceProvider{
			"mobile_money_operator": all,
			"switching_and_processing_company": {
				{ID: "switching-and-processing-company-unified-payment-services-limited", Name: "Unified Payment Services Limited", InstitutionType: "switching_and_processing_company", CountryCode: "NG"},
			},
		},
		getResultByID: map[string]models.PaymentServiceProvider{
			"switching-and-processing-company-unified-payment-services-limited": {
				ID: "switching-and-processing-company-unified-payment-services-limited", Name: "Unified Payment Services Limited", InstitutionType: "switching_and_processing_company", CountryCode: "NG",
			},
		},
	}
	svc, err := NewFinanceService(stub)
	if err != nil {
		t.Fatalf("NewFinanceService() error = %v", err)
	}

	listed, err := svc.ListPaymentServiceProviders(context.Background())
	if err != nil {
		t.Fatalf("ListPaymentServiceProviders() error = %v", err)
	}
	if listed == nil || len(listed) != len(all) {
		t.Fatalf("unexpected list-all result: %#v", listed)
	}
	listed[0].Name = "Changed"
	again, err := svc.ListPaymentServiceProviders(context.Background())
	if err != nil {
		t.Fatalf("ListPaymentServiceProviders() second call error = %v", err)
	}
	if again[0].Name != all[0].Name {
		t.Fatal("ListPaymentServiceProviders() shared mutable slice state")
	}

	for _, institutionType := range []string{
		"mobile_money_operator",
		"switching_and_processing_company",
		"payment_solution_service_provider",
		"payment_terminal_service_provider",
		"super_agent",
		"payment_service_holding_company",
		"payment_terminal_service_aggregator",
	} {
		stub.listByTypeCall = 0
		filtered, err := svc.ListPaymentServiceProvidersByType(context.Background(), institutionType)
		if err != nil {
			t.Fatalf("ListPaymentServiceProvidersByType(%s) error = %v", institutionType, err)
		}
		if stub.lastType != institutionType {
			t.Fatalf("repository received %q, want %q", stub.lastType, institutionType)
		}
		if filtered == nil {
			t.Fatalf("ListPaymentServiceProvidersByType(%s) returned nil slice", institutionType)
		}
	}

	provider, err := svc.GetPaymentServiceProvider(context.Background(), "  switching-and-processing-company-unified-payment-services-limited  ")
	if err != nil {
		t.Fatalf("GetPaymentServiceProvider() error = %v", err)
	}
	if provider.ID != "switching-and-processing-company-unified-payment-services-limited" {
		t.Fatalf("unexpected provider result: %#v", provider)
	}
	if stub.lastID != "switching-and-processing-company-unified-payment-services-limited" {
		t.Fatalf("repository received %q, want trimmed id", stub.lastID)
	}
}

func TestFinanceServiceNormalizesNilAndEmptyLists(t *testing.T) {
	t.Parallel()

	stub := &financeRepositoryStub{}
	svc, err := NewFinanceService(stub)
	if err != nil {
		t.Fatalf("NewFinanceService() error = %v", err)
	}

	listed, err := svc.ListPaymentServiceProviders(context.Background())
	if err != nil {
		t.Fatalf("ListPaymentServiceProviders() error = %v", err)
	}
	if listed == nil {
		t.Fatal("ListPaymentServiceProviders() returned nil slice")
	}

	filtered, err := svc.ListPaymentServiceProvidersByType(context.Background(), "mobile_money_operator")
	if err != nil {
		t.Fatalf("ListPaymentServiceProvidersByType() error = %v", err)
	}
	if filtered == nil {
		t.Fatal("ListPaymentServiceProvidersByType() returned nil slice")
	}
}

func TestFinanceServiceRejectsInvalidInputsBeforeRepositoryAccess(t *testing.T) {
	t.Parallel()

	stub := &financeRepositoryStub{}
	svc, err := NewFinanceService(stub)
	if err != nil {
		t.Fatalf("NewFinanceService() error = %v", err)
	}

	for _, providerID := range []string{"", "   ", "Paystack Payment Limited", "bad id", "PAYSTACK"} {
		if _, err := svc.GetPaymentServiceProvider(context.Background(), providerID); !errors.Is(err, ErrInvalidPaymentServiceProviderID) {
			t.Fatalf("GetPaymentServiceProvider(%q) error = %v, want ErrInvalidPaymentServiceProviderID", providerID, err)
		}
	}
	if stub.getCalls != 0 {
		t.Fatalf("repository was called for invalid ids: %d", stub.getCalls)
	}

	for _, institutionType := range []string{"", "   ", "Mobile Money Operators", "mobile-money-operator", "unknown_type"} {
		if _, err := svc.ListPaymentServiceProvidersByType(context.Background(), institutionType); !errors.Is(err, ErrInvalidPaymentServiceProviderType) {
			t.Fatalf("ListPaymentServiceProvidersByType(%q) error = %v, want ErrInvalidPaymentServiceProviderType", institutionType, err)
		}
	}
	if stub.listByTypeCall != 0 {
		t.Fatalf("repository was called for invalid types: %d", stub.listByTypeCall)
	}
}

func TestFinanceServiceErrorTranslationAndContextPreservation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(*FinanceService) error
	}{
		{
			name: "list all not found",
			call: func(svc *FinanceService) error {
				_, err := svc.ListPaymentServiceProviders(context.Background())
				return err
			},
		},
		{
			name: "list by type not found",
			call: func(svc *FinanceService) error {
				_, err := svc.ListPaymentServiceProvidersByType(context.Background(), "super_agent")
				return err
			},
		},
		{
			name: "get not found",
			call: func(svc *FinanceService) error {
				_, err := svc.GetPaymentServiceProvider(context.Background(), "super-agent-missing")
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &financeRepositoryStub{
				listAllErr:    interfaces.ErrInvalidDatasetFile,
				listByTypeErr: interfaces.ErrInvalidDatasetFile,
				getErr:        interfaces.ErrPaymentServiceProviderNotFound,
			}
			svc, err := NewFinanceService(stub)
			if err != nil {
				t.Fatalf("NewFinanceService() error = %v", err)
			}
			err = tc.call(svc)
			if err == nil {
				t.Fatal("expected error")
			}
			switch tc.name {
			case "get not found":
				if !errors.Is(err, ErrPaymentServiceProviderNotFound) {
					t.Fatalf("unexpected error: %v", err)
				}
			default:
				if !strings.Contains(err.Error(), "repository unavailable") {
					t.Fatalf("unexpected sanitized error: %v", err)
				}
			}
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stub := &financeRepositoryStub{listAllResult: []models.PaymentServiceProvider{{ID: "a", Name: "A", InstitutionType: "mobile_money_operator", CountryCode: "NG"}}}
	svc, err := NewFinanceService(stub)
	if err != nil {
		t.Fatalf("NewFinanceService() error = %v", err)
	}
	if _, err := svc.ListPaymentServiceProviders(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context error = %v, want context.Canceled", err)
	}
	if stub.listAllCalls != 0 {
		t.Fatalf("repository called for canceled context: %d", stub.listAllCalls)
	}

	deadlineCtx, cancelDeadline := context.WithTimeout(context.Background(), time.Nanosecond)
	time.Sleep(2 * time.Nanosecond)
	cancelDeadline()
	if _, err := svc.GetPaymentServiceProvider(deadlineCtx, "mobile-money-operator-a"); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline error = %v, want context cancellation/deadline", err)
	}
}

func TestFinanceServiceRepositoryNotCalledOnInvalidInput(t *testing.T) {
	t.Parallel()

	stub := &financeRepositoryStub{}
	svc, err := NewFinanceService(stub)
	if err != nil {
		t.Fatalf("NewFinanceService() error = %v", err)
	}

	if _, err := svc.GetPaymentServiceProvider(context.Background(), "invalid id"); !errors.Is(err, ErrInvalidPaymentServiceProviderID) {
		t.Fatalf("unexpected get error: %v", err)
	}
	if _, err := svc.ListPaymentServiceProvidersByType(context.Background(), "invalid type"); !errors.Is(err, ErrInvalidPaymentServiceProviderType) {
		t.Fatalf("unexpected list-by-type error: %v", err)
	}
	if stub.getCalls != 0 || stub.listByTypeCall != 0 {
		t.Fatalf("repository was called for invalid inputs: get=%d listByType=%d", stub.getCalls, stub.listByTypeCall)
	}
}

func TestFinanceServiceUnknownRepositoryFailuresAreSanitized(t *testing.T) {
	t.Parallel()

	stub := &financeRepositoryStub{
		listAllErr: fmt.Errorf("/private/tmp/finance/payment_service_providers.json: permission denied"),
	}
	svc, err := NewFinanceService(stub)
	if err != nil {
		t.Fatalf("NewFinanceService() error = %v", err)
	}

	_, err = svc.ListPaymentServiceProviders(context.Background())
	if err == nil {
		t.Fatal("expected sanitized repository failure")
	}
	if strings.Contains(err.Error(), "/private/tmp/finance/payment_service_providers.json") {
		t.Fatalf("error leaked filesystem path: %v", err)
	}
	if !strings.Contains(err.Error(), "repository unavailable") {
		t.Fatalf("unexpected sanitized error: %v", err)
	}
}

func TestFinanceServiceCurrenciesListGetAndValidation(t *testing.T) {
	t.Parallel()

	stub := &financeRepositoryStub{
		listCurrencyResult: []models.Currency{
			{ID: "afn", Name: "Afghani", AlphabeticCode: "AFN", NumericCode: "971", MinorUnit: 2, CountryAreaIDs: []string{"af"}},
			{ID: "usd", Name: "US Dollar", AlphabeticCode: "USD", NumericCode: "840", MinorUnit: 2, CountryAreaIDs: []string{"us"}},
		},
		getCurrencyResult: map[string]models.Currency{
			"usd": {ID: "usd", Name: "US Dollar", AlphabeticCode: "USD", NumericCode: "840", MinorUnit: 2, CountryAreaIDs: []string{"us"}},
		},
	}
	svc, err := NewFinanceService(stub)
	if err != nil {
		t.Fatalf("NewFinanceService() error = %v", err)
	}

	listed, err := svc.ListCurrencies(context.Background())
	if err != nil {
		t.Fatalf("ListCurrencies() error = %v", err)
	}
	if listed == nil {
		t.Fatal("ListCurrencies() returned nil slice")
	}
	if !reflect.DeepEqual(listed, stub.listCurrencyResult) {
		t.Fatalf("unexpected list result: %#v", listed)
	}
	listed[0].Name = "Changed"
	again, err := svc.ListCurrencies(context.Background())
	if err != nil {
		t.Fatalf("ListCurrencies() second call error = %v", err)
	}
	if again[0].Name != stub.listCurrencyResult[0].Name {
		t.Fatal("ListCurrencies() shared mutable slice state")
	}

	got, err := svc.GetCurrency(context.Background(), "  usd  ")
	if err != nil {
		t.Fatalf("GetCurrency() error = %v", err)
	}
	if got.ID != "usd" || stub.lastCurrencyID != "usd" {
		t.Fatalf("unexpected currency lookup result: %#v / last id %q", got, stub.lastCurrencyID)
	}

	for _, currencyID := range []string{"", "   ", "USD", "usd-1", "us d", "12"} {
		if _, err := svc.GetCurrency(context.Background(), currencyID); !errors.Is(err, ErrInvalidCurrencyID) {
			t.Fatalf("GetCurrency(%q) error = %v, want ErrInvalidCurrencyID", currencyID, err)
		}
	}
	if stub.getCurrencyCalls != 1 {
		t.Fatalf("repository called for invalid currency ids: %d", stub.getCurrencyCalls)
	}
}

func TestFinanceServiceCurrencyErrorTranslationAndContextPreservation(t *testing.T) {
	t.Parallel()

	stub := &financeRepositoryStub{
		listCurrencyErr: interfaces.ErrInvalidDatasetFile,
		getCurrencyErr:  interfaces.ErrCurrencyNotFound,
	}
	svc, err := NewFinanceService(stub)
	if err != nil {
		t.Fatalf("NewFinanceService() error = %v", err)
	}

	if _, err := svc.ListCurrencies(context.Background()); err == nil || !strings.Contains(err.Error(), "repository unavailable") {
		t.Fatalf("unexpected list error: %v", err)
	}
	if _, err := svc.GetCurrency(context.Background(), "usd"); !errors.Is(err, ErrCurrencyNotFound) {
		t.Fatalf("unexpected not found translation: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.ListCurrencies(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list error = %v, want context.Canceled", err)
	}
	deadlineCtx, cancelDeadline := context.WithTimeout(context.Background(), time.Nanosecond)
	time.Sleep(2 * time.Nanosecond)
	cancelDeadline()
	if _, err := svc.GetCurrency(deadlineCtx, "usd"); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline error = %v, want context cancellation/deadline", err)
	}
}
