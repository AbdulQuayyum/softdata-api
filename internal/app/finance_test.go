package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type financeServiceStub struct {
	providers  []models.PaymentServiceProvider
	operators  []models.InternationalMoneyTransferOperator
	currencies []models.Currency
	err        error
	calls      int
	imtoCalls  int
	currCalls  int
}

func (s *financeServiceStub) ListPaymentServiceProviders(context.Context) ([]models.PaymentServiceProvider, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return append([]models.PaymentServiceProvider(nil), s.providers...), nil
}

func (s *financeServiceStub) ListPaymentServiceProvidersByType(context.Context, string) ([]models.PaymentServiceProvider, error) {
	return nil, nil
}

func (s *financeServiceStub) GetPaymentServiceProvider(context.Context, string) (models.PaymentServiceProvider, error) {
	return models.PaymentServiceProvider{}, nil
}

func (s *financeServiceStub) ListInternationalMoneyTransferOperators(context.Context) ([]models.InternationalMoneyTransferOperator, error) {
	s.imtoCalls++
	return append([]models.InternationalMoneyTransferOperator(nil), s.operators...), nil
}

func (s *financeServiceStub) GetInternationalMoneyTransferOperator(context.Context, string) (models.InternationalMoneyTransferOperator, error) {
	return models.InternationalMoneyTransferOperator{}, nil
}

func (s *financeServiceStub) ListCurrencies(context.Context, services.CurrencyListInput) ([]models.Currency, error) {
	s.currCalls++
	return append([]models.Currency(nil), s.currencies...), nil
}

func (s *financeServiceStub) GetCurrency(context.Context, string) (models.Currency, error) {
	return models.Currency{}, nil
}

type financeRepositoryStub struct{}

func (s *financeRepositoryStub) ListPaymentServiceProviders(context.Context) ([]models.PaymentServiceProvider, error) {
	return nil, nil
}

func (s *financeRepositoryStub) ListPaymentServiceProvidersByType(context.Context, string) ([]models.PaymentServiceProvider, error) {
	return nil, nil
}

func (s *financeRepositoryStub) GetPaymentServiceProvider(context.Context, string) (models.PaymentServiceProvider, error) {
	return models.PaymentServiceProvider{}, nil
}

func (s *financeRepositoryStub) ListInternationalMoneyTransferOperators(context.Context) ([]models.InternationalMoneyTransferOperator, error) {
	return nil, nil
}

func (s *financeRepositoryStub) GetInternationalMoneyTransferOperator(context.Context, string) (models.InternationalMoneyTransferOperator, error) {
	return models.InternationalMoneyTransferOperator{}, nil
}

func (s *financeRepositoryStub) ListCurrencies(context.Context, interfaces.CurrencyFilter) ([]models.Currency, error) {
	return nil, nil
}

func (s *financeRepositoryStub) GetCurrency(context.Context, string) (models.Currency, error) {
	return models.Currency{}, nil
}

type financeJSONRepoStub struct{}

func (s *financeJSONRepoStub) Decode(context.Context, string, any) error {
	return nil
}

func loadApprovedFinanceProviders(t *testing.T) []models.PaymentServiceProvider {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean("../../datasets/finance/payment_service_providers.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var providers []models.PaymentServiceProvider
	if err := json.Unmarshal(data, &providers); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return providers
}

func loadApprovedIMTOOperators(t *testing.T) []models.InternationalMoneyTransferOperator {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean("../../datasets/finance/international_money_transfer_operators.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var operators []models.InternationalMoneyTransferOperator
	if err := json.Unmarshal(data, &operators); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return operators
}

func writeFinanceFixture(path string, providers []models.PaymentServiceProvider) error {
	data, err := json.Marshal(providers)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func writeIMTOFixture(path string, operators []models.InternationalMoneyTransferOperator) error {
	data, err := json.Marshal(operators)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func TestBuildFinanceHandlerPassesConfiguredDatasetArgs(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         "/tmp/custom-datasets",
			JSONMaxBytes: 9876,
		},
	}

	var gotRoot string
	var gotMaxBytes int64
	var gotPath string
	handler, err := buildFinanceHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			gotRoot = root
			gotMaxBytes = maxBytes
			return &financeJSONRepoStub{}, nil
		},
		func(repository interfaces.JSONFileRepository, paymentServiceProvidersPath string) (interfaces.FinanceRepository, error) {
			gotPath = paymentServiceProvidersPath
			return &financeRepositoryStub{}, nil
		},
		func(repository interfaces.FinanceRepository) (financeService, error) {
			return &financeServiceStub{
				providers: loadApprovedFinanceProviders(t),
				operators: loadApprovedIMTOOperators(t),
			}, nil
		},
		func(service financeService) (*handlers.FinanceHandler, error) {
			return handlers.NewFinanceHandler(service)
		},
	)
	if err != nil {
		t.Fatalf("buildFinanceHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("buildFinanceHandler() returned nil handler")
	}
	if gotRoot != cfg.Datasets.Path {
		t.Fatalf("unexpected dataset root: %q", gotRoot)
	}
	if gotMaxBytes != cfg.Datasets.JSONMaxBytes {
		t.Fatalf("unexpected JSON max bytes: %d", gotMaxBytes)
	}
	if gotPath != financePaymentServiceProvidersRelativePath {
		t.Fatalf("unexpected finance dataset path: %q", gotPath)
	}
}

func TestBuildFinanceHandlerValidFixturePassesStartupVerification(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "finance"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Clean("../../datasets/finance/payment_service_providers.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "finance", "payment_service_providers.json"), data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	imtoData, err := os.ReadFile(filepath.Clean("../../datasets/finance/international_money_transfer_operators.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "finance", "international_money_transfer_operators.json"), imtoData, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	maxBytes := int64(len(data))
	if imtoLen := int64(len(imtoData)); imtoLen > maxBytes {
		maxBytes = imtoLen
	}

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         root,
			JSONMaxBytes: maxBytes + 1024,
		},
	}

	handler, err := buildFinanceHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return fileRepo.NewJSONRepository(root, maxBytes)
		},
		func(repository interfaces.JSONFileRepository, paymentServiceProvidersPath string) (interfaces.FinanceRepository, error) {
			return fileRepo.NewFinanceRepository(repository, paymentServiceProvidersPath, financeInternationalMoneyTransferOperatorsRelativePath)
		},
		func(repository interfaces.FinanceRepository) (financeService, error) {
			return services.NewFinanceService(repository)
		},
		func(service financeService) (*handlers.FinanceHandler, error) {
			return handlers.NewFinanceHandler(service)
		},
	)
	if err != nil {
		t.Fatalf("buildFinanceHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("buildFinanceHandler() returned nil handler")
	}
}

func TestBuildFinanceHandlerFailsSafelyForInvalidDatasets(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedFinanceProviders(t)
	imtoFixture := loadApprovedIMTOOperators(t)
	tests := []struct {
		name string
		set  func(root string) error
	}{
		{
			name: "missing file",
			set: func(root string) error {
				return os.MkdirAll(root, 0o755)
			},
		},
		{
			name: "malformed json",
			set: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "finance"), 0o755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(root, "finance", "payment_service_providers.json"), []byte("{bad"), 0o600)
			},
		},
		{
			name: "wrong record count",
			set: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "finance"), 0o755); err != nil {
					return err
				}
				return writeFinanceFixture(filepath.Join(root, "finance", "payment_service_providers.json"), fixture[:254])
			},
		},
		{
			name: "wrong category composition",
			set: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "finance"), 0o755); err != nil {
					return err
				}
				mutated := append([]models.PaymentServiceProvider(nil), fixture...)
				if mutated[0].InstitutionType == "mobile_money_operator" {
					mutated[0].InstitutionType = "switching_and_processing_company"
				} else {
					mutated[0].InstitutionType = "mobile_money_operator"
				}
				return writeFinanceFixture(filepath.Join(root, "finance", "payment_service_providers.json"), mutated)
			},
		},
		{
			name: "missing imto file",
			set: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "finance"), 0o755); err != nil {
					return err
				}
				return writeFinanceFixture(filepath.Join(root, "finance", "payment_service_providers.json"), fixture)
			},
		},
		{
			name: "wrong imto record count",
			set: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "finance"), 0o755); err != nil {
					return err
				}
				if err := writeFinanceFixture(filepath.Join(root, "finance", "payment_service_providers.json"), fixture); err != nil {
					return err
				}
				return writeIMTOFixture(filepath.Join(root, "finance", "international_money_transfer_operators.json"), imtoFixture[:107])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := tc.set(root); err != nil {
				t.Fatalf("setup() error = %v", err)
			}

			cfg := &config.Config{Datasets: config.DatasetConfig{Path: root, JSONMaxBytes: 2048}}
			_, err := buildFinanceHandler(context.Background(), cfg,
				func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
					return fileRepo.NewJSONRepository(root, maxBytes)
				},
				func(repository interfaces.JSONFileRepository, paymentServiceProvidersPath string) (interfaces.FinanceRepository, error) {
					return fileRepo.NewFinanceRepository(repository, paymentServiceProvidersPath, financeInternationalMoneyTransferOperatorsRelativePath)
				},
				func(repository interfaces.FinanceRepository) (financeService, error) {
					return services.NewFinanceService(repository)
				},
				func(service financeService) (*handlers.FinanceHandler, error) {
					return handlers.NewFinanceHandler(service)
				},
			)
			if err == nil {
				t.Fatal("buildFinanceHandler() error = nil, want failure")
			}
			if strings.Contains(err.Error(), root) {
				t.Fatalf("error leaked dataset root: %v", err)
			}
		})
	}
}

func TestBuildFinanceHandlerPropagatesContextCancellation(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         t.TempDir(),
			JSONMaxBytes: 1,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := buildFinanceHandler(ctx, cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return &financeJSONRepoStub{}, nil
		},
		func(repository interfaces.JSONFileRepository, paymentServiceProvidersPath string) (interfaces.FinanceRepository, error) {
			return &financeRepositoryStub{}, nil
		},
		func(repository interfaces.FinanceRepository) (financeService, error) {
			return &financeServiceStub{
				providers: loadApprovedFinanceProviders(t),
				operators: loadApprovedIMTOOperators(t),
			}, nil
		},
		func(service financeService) (*handlers.FinanceHandler, error) {
			return handlers.NewFinanceHandler(service)
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("buildFinanceHandler() error = %v, want context.Canceled", err)
	}
}

func TestBuildFinanceHandlerVerifiesThroughServiceAbstraction(t *testing.T) {
	t.Parallel()

	service := &financeServiceStub{
		providers: loadApprovedFinanceProviders(t),
		operators: loadApprovedIMTOOperators(t),
	}
	handler, err := buildFinanceHandlerFromJSONRepository(context.Background(), &financeJSONRepoStub{},
		func(repository interfaces.JSONFileRepository, paymentServiceProvidersPath string) (interfaces.FinanceRepository, error) {
			return &financeRepositoryStub{}, nil
		},
		func(repository interfaces.FinanceRepository) (financeService, error) {
			return service, nil
		},
		func(service financeService) (*handlers.FinanceHandler, error) {
			return handlers.NewFinanceHandler(service)
		},
	)
	if err != nil {
		t.Fatalf("buildFinanceHandlerFromJSONRepository() error = %v", err)
	}
	if handler == nil {
		t.Fatal("buildFinanceHandlerFromJSONRepository() returned nil handler")
	}
	if service.calls != 1 {
		t.Fatalf("expected startup verification to call list-all once, got %d", service.calls)
	}
	if service.imtoCalls != 1 {
		t.Fatalf("expected startup verification to call IMTO list once, got %d", service.imtoCalls)
	}
}
