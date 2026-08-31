package file

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

func TestGeographyRepositoryCountriesAndAreasConstructor(t *testing.T) {
	t.Parallel()

	if _, err := NewGeographyRepository(nil, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/countries_and_areas.json"); err == nil {
		t.Fatal("expected nil json repository to be rejected")
	}
	if _, err := NewGeographyRepository(&geographyJSONRepoStub{}, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", ""); err == nil {
		t.Fatal("expected empty countries path to be rejected")
	}

	calls := 0
	stub := &geographyJSONRepoStub{decodeFn: func(context.Context, string, any) error {
		calls++
		return nil
	}}
	repo, err := NewGeographyRepository(stub, " geography/states.json ", " geography/geopolitical_zones.json ", " geography/lgas.json ", " geography/countries_and_areas.json ")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	if repo.countriesAndAreasPath != "geography/countries_and_areas.json" {
		t.Fatalf("unexpected countries path: %q", repo.countriesAndAreasPath)
	}
	if calls != 0 {
		t.Fatalf("constructor performed unexpected decode calls: %d", calls)
	}
}

func TestGeographyRepositoryCountriesAndAreasListGetAndFiltering(t *testing.T) {
	t.Parallel()

	states := loadApprovedStateFixture(t)
	zones := loadApprovedZoneFixture(t)
	units := loadApprovedLocalGovernmentUnitFixture(t)
	countries := loadApprovedCountryFixture(t)
	jsonRepo := &geographyJSONRepoStub{
		states:    states,
		zones:     zones,
		units:     units,
		countries: countries,
		pathCalls: map[string]int{},
	}
	repo, err := NewGeographyRepository(jsonRepo, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}

	loaded, err := repo.ListCountriesAndAreas(context.Background(), interfaces.CountryOrAreaFilter{})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("ListCountriesAndAreas() returned nil slice")
	}
	if len(loaded) != 248 {
		t.Fatalf("unexpected country count: %d", len(loaded))
	}
	if !reflect.DeepEqual(loaded, countries) {
		t.Fatal("ListCountriesAndAreas() returned unexpected records")
	}
	if jsonRepo.pathCalls["geography/countries_and_areas.json"] != 1 {
		t.Fatalf("unexpected countries decode count: %#v", jsonRepo.pathCalls)
	}
	if jsonRepo.pathCalls["geography/states.json"] != 0 || jsonRepo.pathCalls["geography/geopolitical_zones.json"] != 0 || jsonRepo.pathCalls["geography/lgas.json"] != 0 {
		t.Fatalf("country list decoded unrelated geography files: %#v", jsonRepo.pathCalls)
	}

	loaded[0].Name = "Changed"
	again, err := repo.ListCountriesAndAreas(context.Background(), interfaces.CountryOrAreaFilter{})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas() second call error = %v", err)
	}
	if again[0].Name != countries[0].Name {
		t.Fatal("ListCountriesAndAreas() shared mutable slice state")
	}

	regionMatches := filterCountries(countries, interfaces.CountryOrAreaFilter{RegionCode: "002"})
	gotRegion, err := repo.ListCountriesAndAreas(context.Background(), interfaces.CountryOrAreaFilter{RegionCode: "002"})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas(region) error = %v", err)
	}
	if !reflect.DeepEqual(gotRegion, regionMatches) {
		t.Fatalf("unexpected region-filtered results: %#v", gotRegion)
	}

	subregionMatches := filterCountries(countries, interfaces.CountryOrAreaFilter{SubregionCode: "202"})
	gotSubregion, err := repo.ListCountriesAndAreas(context.Background(), interfaces.CountryOrAreaFilter{SubregionCode: "202"})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas(subregion) error = %v", err)
	}
	if !reflect.DeepEqual(gotSubregion, subregionMatches) {
		t.Fatalf("unexpected subregion-filtered results: %#v", gotSubregion)
	}

	combinedMatches := filterCountries(countries, interfaces.CountryOrAreaFilter{RegionCode: "002", SubregionCode: "202"})
	gotCombined, err := repo.ListCountriesAndAreas(context.Background(), interfaces.CountryOrAreaFilter{RegionCode: "002", SubregionCode: "202"})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas(combined) error = %v", err)
	}
	if !reflect.DeepEqual(gotCombined, combinedMatches) {
		t.Fatalf("unexpected combined-filter results: %#v", gotCombined)
	}

	empty, err := repo.ListCountriesAndAreas(context.Background(), interfaces.CountryOrAreaFilter{RegionCode: "999"})
	if err != nil {
		t.Fatalf("ListCountriesAndAreas(unmatched) error = %v", err)
	}
	if empty == nil {
		t.Fatal("ListCountriesAndAreas(unmatched) returned nil slice")
	}
	if len(empty) != 0 {
		t.Fatalf("unexpected unmatched result count: %d", len(empty))
	}

	nigeria, err := repo.GetCountryOrArea(context.Background(), "ng")
	if err != nil {
		t.Fatalf("GetCountryOrArea(ng) error = %v", err)
	}
	if nigeria.Name != "Nigeria" || nigeria.Alpha2Code != "NG" || nigeria.Alpha3Code != "NGA" || nigeria.NumericCode != "566" {
		t.Fatalf("unexpected Nigeria record: %#v", nigeria)
	}

	if _, err := repo.GetCountryOrArea(context.Background(), "zz"); !errors.Is(err, interfaces.ErrCountryOrAreaNotFound) {
		t.Fatalf("missing lookup error = %v, want ErrCountryOrAreaNotFound", err)
	}
}

func TestGeographyRepositoryCountriesAndAreasOperationalIndependence(t *testing.T) {
	t.Parallel()

	states := loadApprovedStateFixture(t)
	zones := loadApprovedZoneFixture(t)
	units := loadApprovedLocalGovernmentUnitFixture(t)
	countries := loadApprovedCountryFixture(t)

	t.Run("country operations only read countries file", func(t *testing.T) {
		jsonRepo := &geographyJSONRepoStub{states: states, zones: zones, units: units, countries: countries, pathCalls: map[string]int{}}
		repo, err := NewGeographyRepository(jsonRepo, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/countries_and_areas.json")
		if err != nil {
			t.Fatalf("NewGeographyRepository() error = %v", err)
		}
		if _, err := repo.ListCountriesAndAreas(context.Background(), interfaces.CountryOrAreaFilter{}); err != nil {
			t.Fatalf("ListCountriesAndAreas() error = %v", err)
		}
		if jsonRepo.pathCalls["geography/countries_and_areas.json"] != 1 {
			t.Fatalf("unexpected country decode count: %#v", jsonRepo.pathCalls)
		}
		if jsonRepo.pathCalls["geography/states.json"] != 0 || jsonRepo.pathCalls["geography/geopolitical_zones.json"] != 0 || jsonRepo.pathCalls["geography/lgas.json"] != 0 {
			t.Fatalf("country list decoded unrelated files: %#v", jsonRepo.pathCalls)
		}
	})

	t.Run("state zone and lga operations do not read country file", func(t *testing.T) {
		jsonRepo := &geographyJSONRepoStub{states: states, zones: zones, units: units, countries: countries, pathCalls: map[string]int{}}
		repo, err := NewGeographyRepository(jsonRepo, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/countries_and_areas.json")
		if err != nil {
			t.Fatalf("NewGeographyRepository() error = %v", err)
		}
		if _, err := repo.ListStates(context.Background()); err != nil {
			t.Fatalf("ListStates() error = %v", err)
		}
		if _, err := repo.ListGeopoliticalZones(context.Background()); err != nil {
			t.Fatalf("ListGeopoliticalZones() error = %v", err)
		}
		if _, err := repo.ListLocalGovernmentUnits(context.Background()); err != nil {
			t.Fatalf("ListLocalGovernmentUnits() error = %v", err)
		}
		if jsonRepo.pathCalls["geography/countries_and_areas.json"] != 0 {
			t.Fatalf("unrelated file was decoded: %#v", jsonRepo.pathCalls)
		}
	})
}

func TestGeographyRepositoryCountriesAndAreasValidationFailures(t *testing.T) {
	t.Parallel()

	states := loadApprovedStateFixture(t)
	zones := loadApprovedZoneFixture(t)
	units := loadApprovedLocalGovernmentUnitFixture(t)
	countries := loadApprovedCountryFixture(t)

	tests := []struct {
		name string
		mut  func([]models.CountryOrArea) []models.CountryOrArea
	}{
		{name: "nil slice", mut: func([]models.CountryOrArea) []models.CountryOrArea { return nil }},
		{name: "empty slice", mut: func([]models.CountryOrArea) []models.CountryOrArea { return []models.CountryOrArea{} }},
		{name: "247 records", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			return append([]models.CountryOrArea(nil), list[:247]...)
		}},
		{name: "249 records", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			return append(append([]models.CountryOrArea(nil), list...), list[0])
		}},
		{name: "duplicate id", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[1].ID = out[0].ID
			return out
		}},
		{name: "duplicate alpha2", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[1].Alpha2Code = out[0].Alpha2Code
			out[1].ID = strings.ToLower(out[1].Alpha2Code)
			return out
		}},
		{name: "duplicate alpha3", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[1].Alpha3Code = out[0].Alpha3Code
			return out
		}},
		{name: "duplicate numeric", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[1].NumericCode = out[0].NumericCode
			return out
		}},
		{name: "id alpha2 mismatch", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].ID = "zz"
			return out
		}},
		{name: "malformed id", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].ID = "bad"
			return out
		}},
		{name: "lowercase alpha2", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].Alpha2Code = strings.ToLower(out[0].Alpha2Code)
			return out
		}},
		{name: "lowercase alpha3", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].Alpha3Code = strings.ToLower(out[0].Alpha3Code)
			return out
		}},
		{name: "malformed numeric", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].NumericCode = "12"
			return out
		}},
		{name: "lost leading zero", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[1].NumericCode = "12"
			return out
		}},
		{name: "empty name", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].Name = ""
			return out
		}},
		{name: "broken hierarchy pair", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].RegionName = ""
			return out
		}},
		{name: "subregion without region", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].RegionCode = ""
			out[0].RegionName = ""
			return out
		}},
		{name: "intermediate without subregion", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].SubregionCode = ""
			out[0].SubregionName = ""
			out[0].IntermediateRegionCode = "001"
			out[0].IntermediateRegionName = "Middle"
			return out
		}},
		{name: "malformed hierarchy code", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].RegionCode = "1"
			return out
		}},
		{name: "incorrect ordering", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0], out[1] = out[1], out[0]
			return out
		}},
		{name: "missing nigeria", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			for i := range out {
				if out[i].ID == "ng" {
					out[i].Name = "Not Nigeria"
					break
				}
			}
			return out
		}},
		{name: "incorrect nigeria codes", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			for i := range out {
				if out[i].ID == "ng" {
					out[i].Alpha2Code = "NN"
					out[i].Alpha3Code = "NNN"
					out[i].NumericCode = "000"
					break
				}
			}
			return out
		}},
		{name: "missing sensitive record", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := make([]models.CountryOrArea, 0, len(list))
			for _, country := range list {
				if country.ID == "va" {
					continue
				}
				out = append(out, country)
			}
			return out
		}},
		{name: "unexpected kosovo record", mut: func(list []models.CountryOrArea) []models.CountryOrArea {
			out := append([]models.CountryOrArea(nil), list...)
			out[0].ID = "xk"
			out[0].Alpha2Code = "XK"
			out[0].Alpha3Code = "XKS"
			out[0].NumericCode = "999"
			out[0].Name = "Kosovo"
			return out
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newCountryRepositoryForTest(t, tc.mut(cloneTestCountries(countries)), states, zones, units)
			if _, err := repo.ListCountriesAndAreas(context.Background(), interfaces.CountryOrAreaFilter{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("ListCountriesAndAreas() error = %v, want ErrInvalidDatasetFile", err)
			}
		})
	}
}

func TestGeographyRepositoryCountriesAndAreasContextCancellationAndDeadline(t *testing.T) {
	t.Parallel()

	repo := newCountryRepositoryForTest(t, loadApprovedCountryFixture(t), loadApprovedStateFixture(t), loadApprovedZoneFixture(t), loadApprovedLocalGovernmentUnitFixture(t))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.ListCountriesAndAreas(ctx, interfaces.CountryOrAreaFilter{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("ListCountriesAndAreas() error = %v, want context.Canceled", err)
	}
	if _, err := repo.GetCountryOrArea(ctx, "ng"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetCountryOrArea() error = %v, want context.Canceled", err)
	}

	deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Minute))
	defer deadlineCancel()
	if _, err := repo.ListCountriesAndAreas(deadlineCtx, interfaces.CountryOrAreaFilter{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ListCountriesAndAreas() error = %v, want context.DeadlineExceeded", err)
	}
	if _, err := repo.GetCountryOrArea(deadlineCtx, "ng"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetCountryOrArea() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestGeographyRepositoryCountriesAndAreasSafeErrorTranslation(t *testing.T) {
	t.Parallel()

	countries := loadApprovedCountryFixture(t)
	stub := &geographyJSONRepoStub{decodeFn: func(context.Context, string, any) error {
		return fmt.Errorf("wrapped: %w", interfaces.ErrDatasetFileNotFound)
	}}
	repo, err := NewGeographyRepository(stub, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	if _, err := repo.ListCountriesAndAreas(context.Background(), interfaces.CountryOrAreaFilter{}); !errors.Is(err, interfaces.ErrDatasetFileNotFound) {
		t.Fatalf("ListCountriesAndAreas() error = %v, want ErrDatasetFileNotFound", err)
	}
	if _, err := repo.GetCountryOrArea(context.Background(), "ng"); !errors.Is(err, interfaces.ErrDatasetFileNotFound) {
		t.Fatalf("GetCountryOrArea() error = %v, want ErrDatasetFileNotFound", err)
	}

	repo = newCountryRepositoryForTest(t, countries, loadApprovedStateFixture(t), loadApprovedZoneFixture(t), loadApprovedLocalGovernmentUnitFixture(t))
	if _, err := repo.GetCountryOrArea(context.Background(), "../ng"); !errors.Is(err, interfaces.ErrCountryOrAreaNotFound) {
		t.Fatalf("malformed lookup error = %v, want ErrCountryOrAreaNotFound", err)
	}
}

func newCountryRepositoryForTest(t *testing.T, countries []models.CountryOrArea, states []models.State, zones []models.GeopoliticalZone, units []models.LocalGovernmentUnit) *GeographyFileRepository {
	t.Helper()

	repo, err := NewGeographyRepository(
		&geographyJSONRepoStub{countries: countries, states: states, zones: zones, units: units, pathCalls: map[string]int{}},
		"geography/states.json",
		"geography/geopolitical_zones.json",
		"geography/lgas.json",
		"geography/countries_and_areas.json",
	)
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	return repo
}

func filterCountries(countries []models.CountryOrArea, filter interfaces.CountryOrAreaFilter) []models.CountryOrArea {
	filtered := make([]models.CountryOrArea, 0, len(countries))
	for _, country := range countries {
		if filter.RegionCode != "" && country.RegionCode != filter.RegionCode {
			continue
		}
		if filter.SubregionCode != "" && country.SubregionCode != filter.SubregionCode {
			continue
		}
		filtered = append(filtered, country)
	}
	return filtered
}
