package services

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestGeographyServiceCountriesAndAreasListFilterAndGet(t *testing.T) {
	t.Parallel()

	countries := loadServiceCountryFixture(t)
	stub := &geographyRepoStub{countries: countries}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	loaded, err := svc.ListCountriesAndAreas(context.Background(), CountryOrAreaListInput{})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("ListCountriesAndAreas() returned nil slice")
	}
	if len(loaded) != 248 {
		t.Fatalf("unexpected country count: %d", len(loaded))
	}
	if loaded[0].ID != countries[0].ID || loaded[len(loaded)-1].ID != countries[len(countries)-1].ID {
		t.Fatalf("unexpected ordering: first=%q last=%q", loaded[0].ID, loaded[len(loaded)-1].ID)
	}

	region := "002"
	byRegion, err := svc.ListCountriesAndAreas(context.Background(), CountryOrAreaListInput{RegionCode: region})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas(region) error = %v", err)
	}
	if !countrySliceEqual(byRegion, filterServiceCountries(countries, CountryOrAreaListInput{RegionCode: region})) {
		t.Fatalf("unexpected region-filtered results: %#v", byRegion)
	}

	subregion := "202"
	bySubregion, err := svc.ListCountriesAndAreas(context.Background(), CountryOrAreaListInput{SubregionCode: subregion})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas(subregion) error = %v", err)
	}
	if !countrySliceEqual(bySubregion, filterServiceCountries(countries, CountryOrAreaListInput{SubregionCode: subregion})) {
		t.Fatalf("unexpected subregion-filtered results: %#v", bySubregion)
	}

	combined, err := svc.ListCountriesAndAreas(context.Background(), CountryOrAreaListInput{RegionCode: region, SubregionCode: subregion})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas(combined) error = %v", err)
	}
	if !countrySliceEqual(combined, filterServiceCountries(countries, CountryOrAreaListInput{RegionCode: region, SubregionCode: subregion})) {
		t.Fatalf("unexpected combined-filter results: %#v", combined)
	}

	unit, err := svc.GetCountryOrArea(context.Background(), "  ng  ")
	if err != nil {
		t.Fatalf("GetCountryOrArea() error = %v", err)
	}
	if stub.lastCountryID != "ng" {
		t.Fatalf("unexpected lookup id: %q", stub.lastCountryID)
	}
	if unit.Name != "Nigeria" || unit.Alpha2Code != "NG" || unit.Alpha3Code != "NGA" || unit.NumericCode != "566" {
		t.Fatalf("unexpected Nigeria record: %#v", unit)
	}
}

func TestGeographyServiceCountriesAndAreasNilResultNormalizesToEmptySlice(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	countries, err := svc.ListCountriesAndAreas(context.Background(), CountryOrAreaListInput{})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas() error = %v", err)
	}
	if countries == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(countries) != 0 {
		t.Fatalf("unexpected length: %d", len(countries))
	}
}

func TestGeographyServiceCountriesAndAreasValidationAndErrors(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{countries: loadServiceCountryFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	for _, tc := range []struct {
		name string
		id   string
		want error
	}{
		{name: "empty", id: "", want: ErrInvalidCountryOrAreaID},
		{name: "uppercase", id: "NG", want: ErrInvalidCountryOrAreaID},
		{name: "space", id: "n g", want: ErrInvalidCountryOrAreaID},
		{name: "underscore", id: "n_g", want: ErrInvalidCountryOrAreaID},
		{name: "hyphen", id: "n-g", want: ErrInvalidCountryOrAreaID},
		{name: "numeric", id: "56", want: ErrInvalidCountryOrAreaID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.countryGetCalls = 0
			if _, err := svc.GetCountryOrArea(context.Background(), tc.id); !errors.Is(err, tc.want) {
				t.Fatalf("GetCountryOrArea() error = %v, want %v", err, tc.want)
			}
			if stub.countryGetCalls != 0 {
				t.Fatalf("repository should not have been called, got %d", stub.countryGetCalls)
			}
		})
	}

	for _, tc := range []struct {
		name   string
		filter CountryOrAreaListInput
		want   error
	}{
		{name: "invalid region", filter: CountryOrAreaListInput{RegionCode: "12a"}, want: ErrInvalidCountryOrAreaRegionCode},
		{name: "invalid subregion", filter: CountryOrAreaListInput{SubregionCode: "1"}, want: ErrInvalidCountryOrAreaSubregionCode},
		{name: "invalid both", filter: CountryOrAreaListInput{RegionCode: "12a", SubregionCode: "1"}, want: ErrInvalidCountryOrAreaRegionCode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.countryListCalls = 0
			if _, err := svc.ListCountriesAndAreas(context.Background(), tc.filter); !errors.Is(err, tc.want) {
				t.Fatalf("ListCountriesAndAreas() error = %v, want %v", err, tc.want)
			}
			if stub.countryListCalls != 0 {
				t.Fatalf("repository should not have been called, got %d", stub.countryListCalls)
			}
		})
	}

	stub.countryGetErr = interfaces.ErrCountryOrAreaNotFound
	if _, err := svc.GetCountryOrArea(context.Background(), "zz"); !errors.Is(err, ErrCountryOrAreaNotFound) {
		t.Fatalf("missing country error = %v, want ErrCountryOrAreaNotFound", err)
	}

	stub.countryGetErr = errors.New("boom")
	if _, err := svc.GetCountryOrArea(context.Background(), "ng"); err == nil || strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected country error was not sanitized: %v", err)
	}

	stub.countryListErr = errors.New("explode")
	if _, err := svc.ListCountriesAndAreas(context.Background(), CountryOrAreaListInput{}); err == nil || strings.Contains(err.Error(), "explode") {
		t.Fatalf("unexpected country list error was not sanitized: %v", err)
	}
}

func TestGeographyServiceCountriesAndAreasContextCancellationAndDeadline(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{countries: loadServiceCountryFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.ListCountriesAndAreas(ctx, CountryOrAreaListInput{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("ListCountriesAndAreas() error = %v, want context.Canceled", err)
	}
	if _, err := svc.GetCountryOrArea(ctx, "ng"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetCountryOrArea() error = %v, want context.Canceled", err)
	}

	deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Minute))
	defer deadlineCancel()
	if _, err := svc.ListCountriesAndAreas(deadlineCtx, CountryOrAreaListInput{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ListCountriesAndAreas() error = %v, want context.DeadlineExceeded", err)
	}
	if _, err := svc.GetCountryOrArea(deadlineCtx, "ng"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetCountryOrArea() error = %v, want context.DeadlineExceeded", err)
	}
}

func filterServiceCountries(countries []models.CountryOrArea, input CountryOrAreaListInput) []models.CountryOrArea {
	filtered := make([]models.CountryOrArea, 0, len(countries))
	for _, country := range countries {
		if input.RegionCode != "" && country.RegionCode != input.RegionCode {
			continue
		}
		if input.SubregionCode != "" && country.SubregionCode != input.SubregionCode {
			continue
		}
		filtered = append(filtered, country)
	}
	return filtered
}

func countrySliceEqual(a, b []models.CountryOrArea) bool {
	return reflect.DeepEqual(a, b)
}
