package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

type countryProfileStub struct {
	countryFn  func(context.Context, string) (models.CountryOrArea, error)
	currencyFn func(context.Context, CurrencyListInput) ([]models.Currency, error)
	timeZoneFn func(context.Context, TimeZoneListInput) ([]models.TimeZone, error)

	countryCalls  int
	currencyCalls int
	timeZoneCalls int
	lastCountryID string
	lastCurrency  CurrencyListInput
	lastTimeZone  TimeZoneListInput
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
	}
	svc, err := NewCountryProfileService(stub, stub, stub)
	if err != nil {
		t.Fatalf("NewCountryProfileService() error = %v", err)
	}

	profile, err := svc.GetCountryProfile(context.Background(), " ng ")
	if err != nil {
		t.Fatalf("GetCountryProfile() error = %v", err)
	}
	if stub.countryCalls != 1 || stub.currencyCalls != 1 || stub.timeZoneCalls != 1 {
		t.Fatalf("unexpected call counts: country=%d currency=%d time_zone=%d", stub.countryCalls, stub.currencyCalls, stub.timeZoneCalls)
	}
	if stub.lastCountryID != "ng" || stub.lastCurrency.CountryAreaID != "ng" || stub.lastTimeZone.CountryAreaID != "ng" {
		t.Fatalf("unexpected lookup inputs: country=%q currency=%#v time_zone=%#v", stub.lastCountryID, stub.lastCurrency, stub.lastTimeZone)
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
	if profile.CallingCodes == nil || len(profile.CallingCodes) != 1 || profile.CallingCodes[0] != "+234" {
		t.Fatalf("unexpected calling codes: %#v", profile.CallingCodes)
	}
}

func TestCountryProfileServiceValidationAndErrors(t *testing.T) {
	t.Parallel()

	stub := &countryProfileStub{}
	svc, err := NewCountryProfileService(stub, stub, stub)
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
			if stub.countryCalls != 0 || stub.currencyCalls != 0 || stub.timeZoneCalls != 0 {
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
	svc, err := NewCountryProfileService(stub, stub, stub)
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
