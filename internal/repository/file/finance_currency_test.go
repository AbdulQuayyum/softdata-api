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

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type currencyJSONRepoStub struct {
	decodeFn func(context.Context, string, any) error

	calls    map[string]int
	lastPath string
}

func (s *currencyJSONRepoStub) Decode(ctx context.Context, relativePath string, destination any) error {
	if s.calls == nil {
		s.calls = make(map[string]int)
	}
	s.calls[relativePath]++
	s.lastPath = relativePath
	if s.decodeFn != nil {
		return s.decodeFn(ctx, relativePath, destination)
	}
	return fmt.Errorf("unexpected destination %T", destination)
}

func TestFinanceRepositoryCurrenciesListGetAndOwnership(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedCurrencies(t)
	countries := readFinanceDatasetBytes(t, financeDatasetPath("geography", "countries_and_areas.json"))
	repo := mustNewCurrencyRepositoryFromRecords(t, fixture, countries)

	listed, err := repo.ListCurrencies(context.Background())
	if err != nil {
		t.Fatalf("ListCurrencies() error = %v", err)
	}
	if listed == nil {
		t.Fatal("ListCurrencies() returned nil slice")
	}
	if !reflect.DeepEqual(listed, fixture) {
		t.Fatalf("unexpected currency list: %#v", listed)
	}

	listed[0].Name = "Changed"
	again, err := repo.ListCurrencies(context.Background())
	if err != nil {
		t.Fatalf("ListCurrencies() second call error = %v", err)
	}
	if again[0].Name != fixture[0].Name {
		t.Fatal("ListCurrencies() shared mutable slice state")
	}

	got, err := repo.GetCurrency(context.Background(), " usd ")
	if err != nil {
		t.Fatalf("GetCurrency() error = %v", err)
	}
	if got.ID != "usd" || got.AlphabeticCode != "USD" || got.NumericCode != "840" {
		t.Fatalf("unexpected currency lookup result: %#v", got)
	}

	twd, err := repo.GetCurrency(context.Background(), "twd")
	if err != nil {
		t.Fatalf("GetCurrency(twd) error = %v", err)
	}
	if !reflect.DeepEqual(twd.CountryAreaIDs, []string{}) {
		t.Fatalf("unexpected TWD country_area_ids: %#v", twd.CountryAreaIDs)
	}

	if _, err := repo.GetCurrency(context.Background(), "USD"); !errors.Is(err, interfaces.ErrCurrencyNotFound) {
		t.Fatalf("unexpected uppercase lookup error: %v", err)
	}
	if _, err := repo.GetCurrency(context.Background(), "zzz"); !errors.Is(err, interfaces.ErrCurrencyNotFound) {
		t.Fatalf("unexpected missing lookup error: %v", err)
	}
}

func TestFinanceRepositoryCurrenciesRejectInvalidFixtures(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedCurrencies(t)
	countries := readFinanceDatasetBytes(t, financeDatasetPath("geography", "countries_and_areas.json"))

	tests := []struct {
		name       string
		mutate     func([]models.Currency) []models.Currency
		countryRaw []byte
		wantErr    error
	}{
		{
			name: "nil slice",
			mutate: func([]models.Currency) []models.Currency {
				return nil
			},
			countryRaw: countries,
			wantErr:    interfaces.ErrInvalidDatasetFile,
		},
		{
			name: "empty slice",
			mutate: func([]models.Currency) []models.Currency {
				return []models.Currency{}
			},
			countryRaw: countries,
			wantErr:    interfaces.ErrInvalidDatasetFile,
		},
		{
			name: "wrong record count",
			mutate: func(records []models.Currency) []models.Currency {
				return append([]models.Currency(nil), records[:154]...)
			},
			countryRaw: countries,
			wantErr:    interfaces.ErrInvalidDatasetFile,
		},
		{
			name: "duplicate id",
			mutate: func(records []models.Currency) []models.Currency {
				out := append([]models.Currency(nil), records...)
				out[1].ID = out[0].ID
				return out
			},
			countryRaw: countries,
			wantErr:    interfaces.ErrInvalidDatasetFile,
		},
		{
			name: "invalid country mapping",
			mutate: func(records []models.Currency) []models.Currency {
				out := append([]models.Currency(nil), records...)
				out[0].CountryAreaIDs = []string{"zz"}
				return out
			},
			countryRaw: countries,
			wantErr:    interfaces.ErrInvalidDatasetFile,
		},
		{
			name: "wrong zero mapping",
			mutate: func(records []models.Currency) []models.Currency {
				out := append([]models.Currency(nil), records...)
				for i := range out {
					if out[i].AlphabeticCode == "TWD" {
						out[i].CountryAreaIDs = []string{"us"}
						break
					}
				}
				return out
			},
			countryRaw: countries,
			wantErr:    interfaces.ErrInvalidDatasetFile,
		},
		{
			name: "incorrect ordering",
			mutate: func(records []models.Currency) []models.Currency {
				out := append([]models.Currency(nil), records...)
				out[0], out[1] = out[1], out[0]
				return out
			},
			countryRaw: countries,
			wantErr:    interfaces.ErrInvalidDatasetFile,
		},
		{
			name: "malformed countries file",
			mutate: func(records []models.Currency) []models.Currency {
				return append([]models.Currency(nil), records...)
			},
			countryRaw: []byte("{bad"),
			wantErr:    interfaces.ErrInvalidDatasetFile,
		},
		{
			name: "missing countries file",
			mutate: func(records []models.Currency) []models.Currency {
				return append([]models.Currency(nil), records...)
			},
			countryRaw: nil,
			wantErr:    interfaces.ErrDatasetFileNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mustNewCurrencyRepositoryFromRecords(t, tc.mutate(fixture), tc.countryRaw)
			_, err := repo.ListCurrencies(context.Background())
			if err == nil {
				t.Fatal("expected failure")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestFinanceRepositoryCurrencyDecodeCountsAndSanitizedErrors(t *testing.T) {
	t.Parallel()

	currencies := loadApprovedCurrencies(t)
	countries := loadApprovedCountryOrAreas(t)
	stub := &currencyJSONRepoStub{decodeFn: func(ctx context.Context, relativePath string, destination any) error {
		switch dest := destination.(type) {
		case *[]models.Currency:
			*dest = append([]models.Currency(nil), currencies...)
			return nil
		case *[]models.CountryOrArea:
			*dest = append([]models.CountryOrArea(nil), countries...)
			return nil
		default:
			return fmt.Errorf("unexpected destination %T", destination)
		}
	}}
	repo, err := NewFinanceRepository(stub, "finance/payment_service_providers.json")
	if err != nil {
		t.Fatalf("NewFinanceRepository() error = %v", err)
	}

	for _, call := range []struct {
		name string
		fn   func() error
	}{
		{
			name: "list",
			fn: func() error {
				_, err := repo.ListCurrencies(context.Background())
				return err
			},
		},
		{
			name: "get",
			fn: func() error {
				_, err := repo.GetCurrency(context.Background(), "usd")
				return err
			},
		},
	} {
		stub.calls = map[string]int{}
		if err := call.fn(); err != nil {
			t.Fatalf("%s call error = %v", call.name, err)
		}
		if stub.calls[financeCurrenciesRelativePath] != 1 {
			t.Fatalf("%s call decoded currencies %d times", call.name, stub.calls[financeCurrenciesRelativePath])
		}
		if stub.calls[financeCountriesAndAreasRelativePath] != 1 {
			t.Fatalf("%s call decoded countries %d times", call.name, stub.calls[financeCountriesAndAreasRelativePath])
		}
	}

	secretErr := errors.New("/private/tmp/finance/currencies.json: permission denied")
	sanitizedRepo, err := NewFinanceRepository(&currencyJSONRepoStub{decodeFn: func(context.Context, string, any) error {
		return secretErr
	}}, "finance/payment_service_providers.json")
	if err != nil {
		t.Fatalf("NewFinanceRepository() error = %v", err)
	}
	_, err = sanitizedRepo.ListCurrencies(context.Background())
	if err == nil {
		t.Fatal("expected sanitized error")
	}
	if strings.Contains(err.Error(), "/private/tmp/finance/currencies.json") {
		t.Fatalf("error leaked filesystem path: %v", err)
	}
	if !errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		t.Fatalf("unexpected sanitized error: %v", err)
	}
}

func loadApprovedCurrencies(t *testing.T) []models.Currency {
	t.Helper()

	var currencies []models.Currency
	dec := json.NewDecoder(bytes.NewReader(readFinanceDatasetBytes(t, financeDatasetPath("finance", "currencies.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&currencies); err != nil {
		t.Fatalf("decode approved currencies dataset: %v", err)
	}
	return currencies
}

func loadApprovedCountryOrAreas(t *testing.T) []models.CountryOrArea {
	t.Helper()

	var countries []models.CountryOrArea
	dec := json.NewDecoder(bytes.NewReader(readFinanceDatasetBytes(t, financeDatasetPath("geography", "countries_and_areas.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&countries); err != nil {
		t.Fatalf("decode approved countries dataset: %v", err)
	}
	return countries
}

func mustNewCurrencyRepositoryFromRecords(t *testing.T, currencies []models.Currency, countryRaw []byte) *FinanceFileRepository {
	t.Helper()

	root := t.TempDir()
	currencyPath := filepath.Join(root, "finance", "currencies.json")
	if err := os.MkdirAll(filepath.Dir(currencyPath), 0o755); err != nil {
		t.Fatalf("mkdir currency fixture dir: %v", err)
	}
	currencyData, err := json.MarshalIndent(currencies, "", "  ")
	if err != nil {
		t.Fatalf("marshal currency fixture: %v", err)
	}
	if err := os.WriteFile(currencyPath, append(currencyData, '\n'), 0o600); err != nil {
		t.Fatalf("write currency fixture: %v", err)
	}

	countryPath := filepath.Join(root, "geography", "countries_and_areas.json")
	if countryRaw != nil {
		if err := os.MkdirAll(filepath.Dir(countryPath), 0o755); err != nil {
			t.Fatalf("mkdir country fixture dir: %v", err)
		}
		if err := os.WriteFile(countryPath, append(append([]byte(nil), countryRaw...), '\n'), 0o600); err != nil {
			t.Fatalf("write country fixture: %v", err)
		}
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
