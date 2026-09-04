package services

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
)

type countryProfileStub struct {
	countryFn  func(context.Context, string) (models.CountryOrArea, error)
	currencyFn func(context.Context, CurrencyListInput) ([]models.Currency, error)
	timeZoneFn func(context.Context, TimeZoneListInput) ([]models.TimeZone, error)
	languageFn func(context.Context, CountryLanguageListInput) ([]models.CountryLanguage, error)

	countryCalls  int
	currencyCalls int
	timeZoneCalls int
	languageCalls int
	lastCountryID string
	lastCurrency  CurrencyListInput
	lastTimeZone  TimeZoneListInput
	lastLanguage  CountryLanguageListInput
}

func (s *countryProfileStub) GetCountryOrArea(ctx context.Context, countryID string) (models.CountryOrArea, error) {
	s.countryCalls++
	s.lastCountryID = countryID
	if s.countryFn != nil {
		return s.countryFn(ctx, countryID)
	}
	return models.CountryOrArea{}, nil
}

func (s *countryProfileStub) ListCurrencies(ctx context.Context, input CurrencyListInput) ([]models.Currency, error) {
	s.currencyCalls++
	s.lastCurrency = input
	if s.currencyFn != nil {
		return s.currencyFn(ctx, input)
	}
	return nil, nil
}

func (s *countryProfileStub) ListTimeZones(ctx context.Context, input TimeZoneListInput) ([]models.TimeZone, error) {
	s.timeZoneCalls++
	s.lastTimeZone = input
	if s.timeZoneFn != nil {
		return s.timeZoneFn(ctx, input)
	}
	return nil, nil
}

func (s *countryProfileStub) ListCountryLanguages(ctx context.Context, input CountryLanguageListInput) ([]models.CountryLanguage, error) {
	s.languageCalls++
	s.lastLanguage = input
	if s.languageFn != nil {
		return s.languageFn(ctx, input)
	}
	return nil, nil
}

func TestCountryProfileServiceGetCountryProfile(t *testing.T) {
	t.Parallel()

	stub := &countryProfileStub{
		countryFn: func(context.Context, string) (models.CountryOrArea, error) {
			return models.CountryOrArea{
				ID:                     "ng",
				Name:                   "Nigeria",
				Alpha2Code:             "NG",
				Alpha3Code:             "NGA",
				NumericCode:            "566",
				CallingCodes:           []string{"+234"},
				FlagEmoji:              "🇳🇬",
				FlagSVGURL:             "/v1/assets/flags/ng.svg",
				RegionCode:             "002",
				RegionName:             "Africa",
				SubregionCode:          "015",
				SubregionName:          "Western Africa",
				IntermediateRegionCode: "011",
				IntermediateRegionName: "Western Africa",
			}, nil
		},
		currencyFn: func(context.Context, CurrencyListInput) ([]models.Currency, error) {
			return []models.Currency{
				{ID: "usd"},
				{ID: "ngn"},
				{ID: "ngn"},
				{ID: "jpy"},
			}, nil
		},
		timeZoneFn: func(context.Context, TimeZoneListInput) ([]models.TimeZone, error) {
			return []models.TimeZone{
				{ID: "Africa/Lagos"},
				{ID: "Africa/Abidjan"},
				{ID: "Africa/Lagos"},
			}, nil
		},
		languageFn: func(context.Context, CountryLanguageListInput) ([]models.CountryLanguage, error) {
			return []models.CountryLanguage{
				{LanguageID: "yo", Status: "official"},
				{LanguageID: "en", Status: "used"},
				{LanguageID: "en", Status: "de_facto_official"},
				{LanguageID: "ha", Status: "official_regional"},
			}, nil
		},
	}
	svc, err := NewCountryProfileService(stub, stub, stub, stub)
	if err != nil {
		t.Fatalf("NewCountryProfileService() error = %v", err)
	}

	profile, err := svc.GetCountryProfile(context.Background(), " ng ")
	if err != nil {
		t.Fatalf("GetCountryProfile() error = %v", err)
	}
	if stub.countryCalls != 1 || stub.currencyCalls != 1 || stub.timeZoneCalls != 1 || stub.languageCalls != 1 {
		t.Fatalf("unexpected call counts: country=%d currency=%d time_zone=%d language=%d", stub.countryCalls, stub.currencyCalls, stub.timeZoneCalls, stub.languageCalls)
	}
	if stub.lastCountryID != "ng" || stub.lastCurrency.CountryAreaID != "ng" || stub.lastTimeZone.CountryAreaID != "ng" || stub.lastLanguage.CountryAreaID != "ng" {
		t.Fatalf("unexpected lookup inputs: country=%q currency=%#v time_zone=%#v language=%#v", stub.lastCountryID, stub.lastCurrency, stub.lastTimeZone, stub.lastLanguage)
	}
	if profile.ID != "ng" || profile.Name != "Nigeria" {
		t.Fatalf("unexpected profile: %#v", profile)
	}
	if len(profile.CurrencyIDs) != 3 || profile.CurrencyIDs[0] != "jpy" || profile.CurrencyIDs[1] != "ngn" || profile.CurrencyIDs[2] != "usd" {
		t.Fatalf("unexpected currency ids: %#v", profile.CurrencyIDs)
	}
	if len(profile.TimeZoneIDs) != 2 || profile.TimeZoneIDs[0] != "Africa/Abidjan" || profile.TimeZoneIDs[1] != "Africa/Lagos" {
		t.Fatalf("unexpected time zone ids: %#v", profile.TimeZoneIDs)
	}
	wantLanguageIDs := []string{"en", "ha", "yo"}
	if !reflect.DeepEqual(profile.LanguageIDs, wantLanguageIDs) {
		t.Fatalf("unexpected language ids: got %#v want %#v", profile.LanguageIDs, wantLanguageIDs)
	}
	if profile.CallingCodes == nil || len(profile.CallingCodes) != 1 || profile.CallingCodes[0] != "+234" {
		t.Fatalf("unexpected calling codes: %#v", profile.CallingCodes)
	}
}

func TestCountryProfileServiceValidationAndErrors(t *testing.T) {
	t.Parallel()

	stub := &countryProfileStub{}
	svc, err := NewCountryProfileService(stub, stub, stub, stub)
	if err != nil {
		t.Fatalf("NewCountryProfileService() error = %v", err)
	}

	for _, tc := range []struct {
		name string
		id   string
		want error
	}{
		{name: "empty", id: "", want: ErrInvalidCountryOrAreaID},
		{name: "uppercase", id: "NG", want: ErrInvalidCountryOrAreaID},
		{name: "space", id: "n g", want: ErrInvalidCountryOrAreaID},
		{name: "uuid", id: "550e8400-e29b-41d4-a716-446655440000", want: ErrInvalidCountryOrAreaID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.countryCalls = 0
			if _, err := svc.GetCountryProfile(context.Background(), tc.id); !errors.Is(err, tc.want) {
				t.Fatalf("GetCountryProfile() error = %v, want %v", err, tc.want)
			}
			if stub.countryCalls != 0 || stub.currencyCalls != 0 || stub.timeZoneCalls != 0 || stub.languageCalls != 0 {
				t.Fatalf("service should not have been called for invalid id")
			}
		})
	}

	stub.countryFn = func(context.Context, string) (models.CountryOrArea, error) {
		return models.CountryOrArea{}, ErrCountryOrAreaNotFound
	}
	if _, err := svc.GetCountryProfile(context.Background(), "ng"); !errors.Is(err, ErrCountryOrAreaNotFound) {
		t.Fatalf("GetCountryProfile() error = %v, want ErrCountryOrAreaNotFound", err)
	}

	stub.countryFn = func(context.Context, string) (models.CountryOrArea, error) {
		return models.CountryOrArea{}, errors.New("explode")
	}
	if _, err := svc.GetCountryProfile(context.Background(), "ng"); err == nil || strings.Contains(err.Error(), "explode") {
		t.Fatalf("unexpected country error was not sanitized: %v", err)
	}

	stub.countryFn = func(context.Context, string) (models.CountryOrArea, error) {
		return models.CountryOrArea{ID: "ng", Name: "Nigeria", Alpha2Code: "NG", Alpha3Code: "NGA", NumericCode: "566"}, nil
	}
	stub.currencyFn = func(context.Context, CurrencyListInput) ([]models.Currency, error) {
		return nil, errors.New("boom")
	}
	if _, err := svc.GetCountryProfile(context.Background(), "ng"); err == nil || strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected currency error was not sanitized: %v", err)
	}

	stub.currencyFn = func(context.Context, CurrencyListInput) ([]models.Currency, error) {
		return []models.Currency{}, nil
	}
	stub.timeZoneFn = func(context.Context, TimeZoneListInput) ([]models.TimeZone, error) {
		return nil, errors.New("boom")
	}
	if _, err := svc.GetCountryProfile(context.Background(), "ng"); err == nil || strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected time zone error was not sanitized: %v", err)
	}

	stub.timeZoneFn = func(context.Context, TimeZoneListInput) ([]models.TimeZone, error) { return []models.TimeZone{}, nil }
	stub.languageFn = func(context.Context, CountryLanguageListInput) ([]models.CountryLanguage, error) {
		return nil, errors.New("language backend down")
	}
	if _, err := svc.GetCountryProfile(context.Background(), "ng"); err == nil || strings.Contains(err.Error(), "language backend down") {
		t.Fatalf("unexpected language error was not sanitized: %v", err)
	}
}

func TestCountryProfileServiceContextCancellationAndDeadline(t *testing.T) {
	t.Parallel()

	stub := &countryProfileStub{
		countryFn: func(context.Context, string) (models.CountryOrArea, error) {
			return models.CountryOrArea{ID: "ng", Name: "Nigeria", Alpha2Code: "NG", Alpha3Code: "NGA", NumericCode: "566"}, nil
		},
		currencyFn: func(context.Context, CurrencyListInput) ([]models.Currency, error) {
			return []models.Currency{}, nil
		},
		timeZoneFn: func(context.Context, TimeZoneListInput) ([]models.TimeZone, error) {
			return []models.TimeZone{}, nil
		},
	}
	svc, err := NewCountryProfileService(stub, stub, stub, stub)
	if err != nil {
		t.Fatalf("NewCountryProfileService() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.GetCountryProfile(ctx, "ng"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetCountryProfile() error = %v, want context.Canceled", err)
	}

	deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Minute))
	defer deadlineCancel()
	if _, err := svc.GetCountryProfile(deadlineCtx, "ng"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetCountryProfile() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestCountryProfileServiceUsesPublishedLanguageRelationships(t *testing.T) {
	jsonRepository, err := fileRepo.NewJSONRepository("../../datasets", 10<<20)
	if err != nil {
		t.Fatalf("NewJSONRepository() error = %v", err)
	}
	geographyRepository, err := fileRepo.NewGeographyRepository(
		jsonRepository,
		"geography/states.json",
		"geography/geopolitical_zones.json",
		"geography/lgas.json",
		"geography/time_zones.json",
		"geography/countries_and_areas.json",
		"geography/languages.json",
		"geography/country_languages.json",
	)
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	geographyService, err := NewGeographyService(geographyRepository)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}
	financeRepository, err := fileRepo.NewFinanceRepository(jsonRepository, "finance/payment_service_providers.json")
	if err != nil {
		t.Fatalf("NewFinanceRepository() error = %v", err)
	}
	financeService, err := NewFinanceService(financeRepository)
	if err != nil {
		t.Fatalf("NewFinanceService() error = %v", err)
	}
	profileService, err := NewCountryProfileService(geographyService, financeService, geographyService, geographyService)
	if err != nil {
		t.Fatalf("NewCountryProfileService() error = %v", err)
	}

	wantCounts := map[string]int{"za": 13, "in": 46, "ch": 9, "ca": 58, "us": 29, "gb": 24, "hk": 3, "mo": 6, "ps": 1, "aq": 1}
	for countryID, wantCount := range wantCounts {
		profile, err := profileService.GetCountryProfile(context.Background(), countryID)
		if err != nil {
			t.Fatalf("GetCountryProfile(%q) error = %v", countryID, err)
		}
		if len(profile.LanguageIDs) != wantCount {
			t.Fatalf("%s language count = %d, want %d", countryID, len(profile.LanguageIDs), wantCount)
		}
		if !sort.StringsAreSorted(profile.LanguageIDs) {
			t.Fatalf("%s language IDs are not sorted: %#v", countryID, profile.LanguageIDs)
		}
		for i := 1; i < len(profile.LanguageIDs); i++ {
			if profile.LanguageIDs[i-1] == profile.LanguageIDs[i] {
				t.Fatalf("%s contains duplicate language ID %q", countryID, profile.LanguageIDs[i])
			}
		}
	}

	nigeria, err := profileService.GetCountryProfile(context.Background(), "ng")
	if err != nil {
		t.Fatalf("GetCountryProfile(ng) error = %v", err)
	}
	wantNigeria := []string{"ann", "ar", "bin", "cch", "efi", "en", "ff", "ha", "ibb", "ig", "kaj", "kcg", "pcm", "tiv", "yo"}
	if !reflect.DeepEqual(nigeria.LanguageIDs, wantNigeria) {
		t.Fatalf("Nigeria language IDs = %#v, want %#v", nigeria.LanguageIDs, wantNigeria)
	}

	for _, anchor := range []struct {
		countryID string
		language  string
	}{
		{"ps", "ar"},
		{"aq", "und"},
	} {
		profile, err := profileService.GetCountryProfile(context.Background(), anchor.countryID)
		if err != nil {
			t.Fatalf("GetCountryProfile(%q) error = %v", anchor.countryID, err)
		}
		if !profileContainsString(profile.LanguageIDs, anchor.language) {
			t.Fatalf("%s language IDs %#v do not contain %q", anchor.countryID, profile.LanguageIDs, anchor.language)
		}
	}
}

func TestCountryProfileServiceNormalizesEmptyLanguageRelationships(t *testing.T) {
	stub := &countryProfileStub{
		countryFn: func(context.Context, string) (models.CountryOrArea, error) {
			return models.CountryOrArea{ID: "aq"}, nil
		},
		currencyFn: func(context.Context, CurrencyListInput) ([]models.Currency, error) { return nil, nil },
		timeZoneFn: func(context.Context, TimeZoneListInput) ([]models.TimeZone, error) { return nil, nil },
		languageFn: func(context.Context, CountryLanguageListInput) ([]models.CountryLanguage, error) { return nil, nil },
	}
	svc, err := NewCountryProfileService(stub, stub, stub, stub)
	if err != nil {
		t.Fatalf("NewCountryProfileService() error = %v", err)
	}
	profile, err := svc.GetCountryProfile(context.Background(), "aq")
	if err != nil {
		t.Fatalf("GetCountryProfile() error = %v", err)
	}
	if profile.LanguageIDs == nil {
		t.Fatal("LanguageIDs must be non-nil")
	}
	encoded, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !strings.Contains(string(encoded), `"language_ids":[]`) {
		t.Fatalf("empty language IDs did not serialize as []: %s", encoded)
	}
}

func profileContainsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
