package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type financeJSONRepoStub struct {
	decodeFn func(context.Context, string, any) error

	calls     int
	pathCalls map[string]int
	lastPath  string
}

func (s *financeJSONRepoStub) Decode(ctx context.Context, relativePath string, destination any) error {
	s.calls++
	s.lastPath = relativePath
	if s.pathCalls != nil {
		s.pathCalls[relativePath]++
	}
	if s.decodeFn != nil {
		return s.decodeFn(ctx, relativePath, destination)
	}

	dest, ok := destination.(*[]models.PaymentServiceProvider)
	if !ok {
		return fmt.Errorf("unexpected destination %T", destination)
	}
	*dest = nil
	return nil
}

func TestNewFinanceRepositoryRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	if _, err := NewFinanceRepository(nil, "finance/payment_service_providers.json"); err == nil {
		t.Fatal("expected nil json repository to be rejected")
	}
	if _, err := NewFinanceRepository(&financeJSONRepoStub{}, ""); err == nil {
		t.Fatal("expected empty payment service providers path to be rejected")
	}
	if _, err := NewFinanceRepository(&financeJSONRepoStub{}, "   "); err == nil {
		t.Fatal("expected whitespace payment service providers path to be rejected")
	}
	if _, err := NewFinanceRepository(&financeJSONRepoStub{}, "/tmp/finance/payment_service_providers.json"); err == nil {
		t.Fatal("expected absolute payment service providers path to be rejected")
	}

	repo, err := NewFinanceRepository(&financeJSONRepoStub{}, "  finance/payment_service_providers.json  ")
	if err != nil {
		t.Fatalf("NewFinanceRepository() error = %v", err)
	}
	if repo.paymentServiceProvidersPath != "finance/payment_service_providers.json" {
		t.Fatalf("unexpected stored path: %q", repo.paymentServiceProvidersPath)
	}
}

func TestFinanceRepositoryListAllByTypeAndLookup(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedPaymentServiceProviders(t)
	repo := mustNewFinanceRepositoryFromRecords(t, fixture)

	loaded, err := repo.ListPaymentServiceProviders(context.Background())
	if err != nil {
		t.Fatalf("ListPaymentServiceProviders() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("ListPaymentServiceProviders() returned nil slice")
	}
	if !reflect.DeepEqual(loaded, fixture) {
		t.Fatalf("unexpected payment service provider list: %#v", loaded)
	}

	loaded[0].Name = "Changed"
	again, err := repo.ListPaymentServiceProviders(context.Background())
	if err != nil {
		t.Fatalf("ListPaymentServiceProviders() second call error = %v", err)
	}
	if again[0].Name != fixture[0].Name {
		t.Fatal("ListPaymentServiceProviders() shared mutable slice state")
	}

	wantCounts := map[string]int{
		"mobile_money_operator":               17,
		"switching_and_processing_company":    19,
		"payment_solution_service_provider":   108,
		"payment_terminal_service_provider":   47,
		"super_agent":                         61,
		"payment_service_holding_company":     1,
		"payment_terminal_service_aggregator": 2,
	}
	for institutionType, want := range wantCounts {
		filtered, err := repo.ListPaymentServiceProvidersByType(context.Background(), institutionType)
		if err != nil {
			t.Fatalf("ListPaymentServiceProvidersByType(%s) error = %v", institutionType, err)
		}
		if filtered == nil {
			t.Fatalf("ListPaymentServiceProvidersByType(%s) returned nil slice", institutionType)
		}
		if got := len(filtered); got != want {
			t.Fatalf("unexpected count for %s: got %d want %d", institutionType, got, want)
		}
		for _, provider := range filtered {
			if provider.InstitutionType != institutionType {
				t.Fatalf("filtered slice contains wrong type: %#v", provider)
			}
		}
	}

	if got := repoByName(t, fixture, "Unified Payment Services Limited"); len(got) != 2 {
		t.Fatalf("expected cross-category duplicate provider name to appear twice, got %d", len(got))
	}

	for _, id := range []string{
		"mobile-money-operator-kongapay-technologies-limited",
		"switching-and-processing-company-zone-payment-network-limited",
		"payment-solution-service-provider-cyberpay-limited",
		"payment-terminal-service-provider-funds-konnect-limited",
		"super-agent-crowd-force-limited",
		"payment-service-holding-company-interswitch-pshc-nigeria-limited",
		"payment-terminal-service-aggregator-unified-payment-services-limited",
	} {
		provider, err := repo.GetPaymentServiceProvider(context.Background(), id)
		if err != nil {
			t.Fatalf("GetPaymentServiceProvider(%s) error = %v", id, err)
		}
		if provider.ID != id {
			t.Fatalf("unexpected provider id for %s: %#v", id, provider)
		}
	}

	if _, err := repo.GetPaymentServiceProvider(context.Background(), "Paystack Payment Limited"); !errors.Is(err, interfaces.ErrPaymentServiceProviderNotFound) {
		t.Fatalf("name lookup error = %v, want ErrPaymentServiceProviderNotFound", err)
	}
	if _, err := repo.GetPaymentServiceProvider(context.Background(), "missing"); !errors.Is(err, interfaces.ErrPaymentServiceProviderNotFound) {
		t.Fatalf("missing lookup error = %v, want ErrPaymentServiceProviderNotFound", err)
	}
}

func TestFinanceRepositoryRejectsInvalidFixtures(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedPaymentServiceProviders(t)
	tests := []struct {
		name string
		mut  func([]models.PaymentServiceProvider) []models.PaymentServiceProvider
	}{
		{
			name: "nil slice",
			mut:  func([]models.PaymentServiceProvider) []models.PaymentServiceProvider { return nil },
		},
		{
			name: "empty slice",
			mut: func([]models.PaymentServiceProvider) []models.PaymentServiceProvider {
				return make([]models.PaymentServiceProvider, 0)
			},
		},
		{
			name: "254 records",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				return append([]models.PaymentServiceProvider(nil), records[:254]...)
			},
		},
		{
			name: "256 records",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				out := append([]models.PaymentServiceProvider(nil), records...)
				return append(out, records[0])
			},
		},
		{
			name: "incorrect category count",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				out := append([]models.PaymentServiceProvider(nil), records...)
				for i := range out {
					if out[i].InstitutionType == "payment_solution_service_provider" {
						out[i].InstitutionType = "mobile_money_operator"
						out[i].ID = "mobile-money-operator-" + slugifyFinancePaymentServiceProviderName(out[i].Name)
						break
					}
				}
				return out
			},
		},
		{
			name: "unknown institution type",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				out := append([]models.PaymentServiceProvider(nil), records...)
				out[0].InstitutionType = "unknown_type"
				out[0].ID = "unknown-type-" + slugifyFinancePaymentServiceProviderName(out[0].Name)
				return out
			},
		},
		{
			name: "missing required fields",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				out := append([]models.PaymentServiceProvider(nil), records...)
				out[0].Name = ""
				return out
			},
		},
		{
			name: "invalid country code",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				out := append([]models.PaymentServiceProvider(nil), records...)
				out[0].CountryCode = "GH"
				return out
			},
		},
		{
			name: "duplicate id",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				out := append([]models.PaymentServiceProvider(nil), records...)
				out[1].ID = out[0].ID
				return out
			},
		},
		{
			name: "duplicate type name pair",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				out := append([]models.PaymentServiceProvider(nil), records...)
				out[1].Name = out[0].Name
				out[1].ID = out[1].InstitutionType + "-" + slugifyFinancePaymentServiceProviderName(out[1].Name)
				return out
			},
		},
		{
			name: "invalid id pattern",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				out := append([]models.PaymentServiceProvider(nil), records...)
				out[0].ID = "bad"
				return out
			},
		},
		{
			name: "incorrect id prefix",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				out := append([]models.PaymentServiceProvider(nil), records...)
				out[0].ID = "switching-and-processing-company-" + slugifyFinancePaymentServiceProviderName(out[0].Name)
				return out
			},
		},
		{
			name: "incorrect ordering",
			mut: func(records []models.PaymentServiceProvider) []models.PaymentServiceProvider {
				out := append([]models.PaymentServiceProvider(nil), records...)
				out[0], out[1] = out[1], out[0]
				return out
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mustNewFinanceRepositoryFromRecords(t, tc.mut(fixture))
			_, err := repo.ListPaymentServiceProviders(context.Background())
			if err == nil {
				t.Fatalf("expected %s to fail", tc.name)
			}
			if !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("unexpected error for %s: %v", tc.name, err)
			}
		})
	}
}

func TestFinanceRepositoryContextAndSanitizedErrors(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedPaymentServiceProviders(t)
	repo := mustNewFinanceRepositoryFromRecords(t, fixture)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.ListPaymentServiceProviders(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context error = %v, want context.Canceled", err)
	}

	deadlineCtx, cancelDeadline := context.WithTimeout(context.Background(), time.Nanosecond)
	time.Sleep(2 * time.Nanosecond)
	cancelDeadline()
	if _, err := repo.ListPaymentServiceProviders(deadlineCtx); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline context error = %v, want context cancellation/deadline", err)
	}

	secretErr := errors.New("/private/tmp/finance/payment_service_providers.json: permission denied")
	stub := &financeJSONRepoStub{decodeFn: func(context.Context, string, any) error { return secretErr }}
	sanitizedRepo, err := NewFinanceRepository(stub, "finance/payment_service_providers.json")
	if err != nil {
		t.Fatalf("NewFinanceRepository() error = %v", err)
	}
	_, err = sanitizedRepo.ListPaymentServiceProviders(context.Background())
	if err == nil {
		t.Fatal("expected sanitized decode error")
	}
	if strings.Contains(err.Error(), "/private/tmp/finance/payment_service_providers.json") {
		t.Fatalf("error leaked filesystem path: %v", err)
	}
	if !errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		t.Fatalf("unexpected sanitized error: %v", err)
	}
}

func TestFinanceRepositoryDoesNotMutateReturnedSlice(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedPaymentServiceProviders(t)
	repo := mustNewFinanceRepositoryFromRecords(t, fixture)

	loaded, err := repo.ListPaymentServiceProvidersByType(context.Background(), "mobile_money_operator")
	if err != nil {
		t.Fatalf("ListPaymentServiceProvidersByType() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("ListPaymentServiceProvidersByType() returned nil slice")
	}
	if len(loaded) == 0 {
		t.Fatal("expected non-empty slice")
	}

	loaded[0].Name = "Changed"
	again, err := repo.ListPaymentServiceProvidersByType(context.Background(), "mobile_money_operator")
	if err != nil {
		t.Fatalf("ListPaymentServiceProvidersByType() second call error = %v", err)
	}
	if again[0].Name != fixture[0].Name {
		t.Fatal("ListPaymentServiceProvidersByType() shared mutable slice state")
	}
}

func TestFinanceRepositoryDecodeCounts(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedPaymentServiceProviders(t)
	zoneRepo := mustNewFinanceRepositoryFromRecords(t, fixture)
	if _, err := zoneRepo.ListPaymentServiceProviders(context.Background()); err != nil {
		t.Fatalf("ListPaymentServiceProviders() error = %v", err)
	}

	stub := &financeJSONRepoStub{decodeFn: func(ctx context.Context, relativePath string, destination any) error {
		records := clonePaymentServiceProviderList(fixture)
		switch dest := destination.(type) {
		case *[]models.PaymentServiceProvider:
			*dest = records
			return nil
		default:
			return fmt.Errorf("unexpected destination %T", destination)
		}
	}}
	repo, err := NewFinanceRepository(stub, "finance/payment_service_providers.json")
	if err != nil {
		t.Fatalf("NewFinanceRepository() error = %v", err)
	}

	for _, call := range []func() error{
		func() error { _, err := repo.ListPaymentServiceProviders(context.Background()); return err },
		func() error {
			_, err := repo.ListPaymentServiceProvidersByType(context.Background(), "super_agent")
			return err
		},
		func() error {
			_, err := repo.GetPaymentServiceProvider(context.Background(), "super-agent-spout-payment-solutions")
			return err
		},
	} {
		stub.calls = 0
		stub.pathCalls = map[string]int{}
		if err := call(); err != nil {
			t.Fatalf("repository call error = %v", err)
		}
		if stub.calls != 1 {
			t.Fatalf("unexpected decode call count: %d", stub.calls)
		}
		if stub.pathCalls["finance/payment_service_providers.json"] != 1 {
			t.Fatalf("unexpected decode path counts: %#v", stub.pathCalls)
		}
	}
}

func loadApprovedPaymentServiceProviders(t *testing.T) []models.PaymentServiceProvider {
	t.Helper()

	var providers []models.PaymentServiceProvider
	dec := json.NewDecoder(bytes.NewReader(readFinanceDatasetBytes(t, financeDatasetPath("finance", "payment_service_providers.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&providers); err != nil {
		t.Fatalf("decode approved payment service providers dataset: %v", err)
	}
	return providers
}

func mustNewFinanceRepositoryFromRecords(t *testing.T, providers []models.PaymentServiceProvider) *FinanceFileRepository {
	t.Helper()

	root := t.TempDir()
	fixturePath := filepath.Join(root, "finance", "payment_service_providers.json")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}
	encoded, err := json.MarshalIndent(providers, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, append(encoded, '\n'), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	jsonRepo, err := NewJSONRepository(root, 16<<20)
	if err != nil {
		t.Fatalf("NewJSONRepository() error = %v", err)
	}
	repo, err := NewFinanceRepository(jsonRepo, "finance/payment_service_providers.json")
	if err != nil {
		t.Fatalf("NewFinanceRepository() error = %v", err)
	}
	return repo
}

func repoByName(t *testing.T, providers []models.PaymentServiceProvider, name string) []models.PaymentServiceProvider {
	t.Helper()

	matches := make([]models.PaymentServiceProvider, 0)
	for _, provider := range providers {
		if provider.Name == name {
			matches = append(matches, provider)
		}
	}
	return matches
}

func readFinanceDatasetBytes(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func financeDatasetPath(parts ...string) string {
	elems := append([]string{"..", "..", "..", "datasets"}, parts...)
	return filepath.Join(elems...)
}
