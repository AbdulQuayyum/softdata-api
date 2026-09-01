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

type geographyJSONRepoStub struct {
	states    []models.State
	zones     []models.GeopoliticalZone
	units     []models.LocalGovernmentUnit
	timeZones []models.TimeZone
	countries []models.CountryOrArea
	decodeFn  func(context.Context, string, any) error

	calls     int
	pathCalls map[string]int
	lastPath  string
}

func (s *geographyJSONRepoStub) Decode(ctx context.Context, relativePath string, destination any) error {
	s.calls++
	s.lastPath = relativePath
	if s.pathCalls != nil {
		s.pathCalls[relativePath]++
	}
	if s.decodeFn != nil {
		return s.decodeFn(ctx, relativePath, destination)
	}

	dest, ok := destination.(*[]models.State)
	if ok {
		*dest = cloneTestStates(s.states)
		return nil
	}
	unitDest, ok := destination.(*[]models.LocalGovernmentUnit)
	if ok {
		*unitDest = cloneTestUnits(s.units)
		return nil
	}
	countryDest, ok := destination.(*[]models.CountryOrArea)
	if ok {
		*countryDest = cloneTestCountries(s.countries)
		return nil
	}
	timeZoneDest, ok := destination.(*[]models.TimeZone)
	if ok {
		*timeZoneDest = cloneTestTimeZones(s.timeZones)
		return nil
	}
	zoneDest, ok := destination.(*[]models.GeopoliticalZone)
	if !ok {
		return fmt.Errorf("unexpected destination %T", destination)
	}
	*zoneDest = cloneTestZones(s.zones)
	return nil
}

func TestNewGeographyRepositoryRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	if _, err := NewGeographyRepository(nil, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json"); err == nil {
		t.Fatal("expected nil json repository to be rejected")
	}
	if _, err := NewGeographyRepository(&geographyJSONRepoStub{}, "", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json"); err == nil {
		t.Fatal("expected empty states path to be rejected")
	}
	if _, err := NewGeographyRepository(&geographyJSONRepoStub{}, "geography/states.json", "   ", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json"); err == nil {
		t.Fatal("expected whitespace zones path to be rejected")
	}
	if _, err := NewGeographyRepository(&geographyJSONRepoStub{}, "geography/states.json", "geography/geopolitical_zones.json", "   ", "geography/time_zones.json", "geography/countries_and_areas.json"); err == nil {
		t.Fatal("expected whitespace lga path to be rejected")
	}
	if _, err := NewGeographyRepository(&geographyJSONRepoStub{}, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "   "); err == nil {
		t.Fatal("expected whitespace countries path to be rejected")
	}

	repo, err := NewGeographyRepository(&geographyJSONRepoStub{}, "  geography/states.json  ", " geography/geopolitical_zones.json ", " geography/lgas.json ", " geography/time_zones.json ", " geography/countries_and_areas.json ")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	if repo.statesPath != "geography/states.json" {
		t.Fatalf("unexpected stored path: %q", repo.statesPath)
	}
	if repo.zonesPath != "geography/geopolitical_zones.json" {
		t.Fatalf("unexpected stored zones path: %q", repo.zonesPath)
	}
	if repo.localGovernmentUnitsPath != "geography/lgas.json" {
		t.Fatalf("unexpected stored lga path: %q", repo.localGovernmentUnitsPath)
	}
	if repo.timeZonesPath != "geography/time_zones.json" {
		t.Fatalf("unexpected stored time zones path: %q", repo.timeZonesPath)
	}
}

func TestGeographyRepositoryListStatesAndGetStateByID(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedStateFixture(t)
	zones := loadApprovedZoneFixture(t)
	jsonRepo := &geographyJSONRepoStub{states: fixture, zones: zones, pathCalls: map[string]int{}}
	repo, err := NewGeographyRepository(jsonRepo, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}

	states, err := repo.ListStates(context.Background())
	if err != nil {
		t.Fatalf("ListStates() error = %v", err)
	}
	if jsonRepo.calls != 2 {
		t.Fatalf("unexpected decode call count: %d", jsonRepo.calls)
	}
	if jsonRepo.pathCalls["geography/lgas.json"] != 0 {
		t.Fatalf("unexpected lga decode count for state lookup path: %#v", jsonRepo.pathCalls)
	}
	if states == nil {
		t.Fatal("ListStates() returned nil slice")
	}
	if !reflect.DeepEqual(states, fixture) {
		t.Fatalf("unexpected state list: %#v", states)
	}

	states[0].Name = "Changed"
	statesAgain, err := repo.ListStates(context.Background())
	if err != nil {
		t.Fatalf("ListStates() second call error = %v", err)
	}
	if statesAgain[0].Name != fixture[0].Name {
		t.Fatal("ListStates() shared mutable slice state")
	}

	abia, err := repo.GetStateByID(context.Background(), "abia")
	if err != nil {
		t.Fatalf("GetStateByID(abia) error = %v", err)
	}
	if abia.ID != "abia" || abia.Name != "Abia" {
		t.Fatalf("unexpected Abia record: %#v", abia)
	}

	akwa, err := repo.GetStateByID(context.Background(), "akwa-ibom")
	if err != nil {
		t.Fatalf("GetStateByID(akwa-ibom) error = %v", err)
	}
	if akwa.ID != "akwa-ibom" {
		t.Fatalf("unexpected Akwa Ibom record: %#v", akwa)
	}

	fct, err := repo.GetStateByID(context.Background(), "fct")
	if err != nil {
		t.Fatalf("GetStateByID(fct) error = %v", err)
	}
	if fct.ID != "fct" || fct.Capital != "Abuja" {
		t.Fatalf("unexpected FCT record: %#v", fct)
	}

	if _, err := repo.GetStateByID(context.Background(), "Abia"); !errors.Is(err, interfaces.ErrStateNotFound) {
		t.Fatalf("uppercase identifier error = %v, want ErrStateNotFound", err)
	}
	if _, err := repo.GetStateByID(context.Background(), "  abia  "); err != nil {
		t.Fatalf("trimmed identifier lookup failed: %v", err)
	}
	if _, err := repo.GetStateByID(context.Background(), ""); !errors.Is(err, interfaces.ErrStateNotFound) {
		t.Fatalf("empty identifier error = %v, want ErrStateNotFound", err)
	}
	if _, err := repo.GetStateByID(context.Background(), "../abia"); !errors.Is(err, interfaces.ErrStateNotFound) {
		t.Fatalf("malformed identifier error = %v, want ErrStateNotFound", err)
	}
	if _, err := repo.GetStateByID(context.Background(), "missing"); !errors.Is(err, interfaces.ErrStateNotFound) {
		t.Fatalf("missing identifier error = %v, want ErrStateNotFound", err)
	}
}

func TestGeographyRepositoryListGeopoliticalZonesAndGetGeopoliticalZone(t *testing.T) {
	t.Parallel()

	states := loadApprovedStateFixture(t)
	zones := loadApprovedZoneFixture(t)
	jsonRepo := &geographyJSONRepoStub{states: states, zones: zones, pathCalls: map[string]int{}}
	repo, err := NewGeographyRepository(jsonRepo, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}

	loaded, err := repo.ListGeopoliticalZones(context.Background())
	if err != nil {
		t.Fatalf("ListGeopoliticalZones() error = %v", err)
	}
	if len(loaded) != 6 {
		t.Fatalf("unexpected zone count: %d", len(loaded))
	}
	if !reflect.DeepEqual(loaded, zones) {
		t.Fatalf("unexpected zone list: %#v", loaded)
	}
	if jsonRepo.calls != 2 {
		t.Fatalf("unexpected decode call count: %d", jsonRepo.calls)
	}
	if jsonRepo.pathCalls["geography/lgas.json"] != 0 {
		t.Fatalf("unexpected lga decode count for zone lookup path: %#v", jsonRepo.pathCalls)
	}

	loaded[0].Name = "Changed"
	again, err := repo.ListGeopoliticalZones(context.Background())
	if err != nil {
		t.Fatalf("ListGeopoliticalZones() second call error = %v", err)
	}
	if again[0].Name != zones[0].Name {
		t.Fatal("ListGeopoliticalZones() shared mutable slice state")
	}

	northCentral, err := repo.GetGeopoliticalZone(context.Background(), "north-central")
	if err != nil {
		t.Fatalf("GetGeopoliticalZone(north-central) error = %v", err)
	}
	if northCentral.ID != "north-central" || northCentral.Name != "North Central" {
		t.Fatalf("unexpected North Central record: %#v", northCentral)
	}

	if _, err := repo.GetGeopoliticalZone(context.Background(), "North Central"); !errors.Is(err, interfaces.ErrGeopoliticalZoneNotFound) {
		t.Fatalf("uppercase zone identifier error = %v, want ErrGeopoliticalZoneNotFound", err)
	}
	if _, err := repo.GetGeopoliticalZone(context.Background(), "missing"); !errors.Is(err, interfaces.ErrGeopoliticalZoneNotFound) {
		t.Fatalf("missing zone identifier error = %v, want ErrGeopoliticalZoneNotFound", err)
	}
}

func TestGeographyRepositoryListLocalGovernmentUnitsAndLookups(t *testing.T) {
	t.Parallel()

	states := loadApprovedStateFixture(t)
	zones := loadApprovedZoneFixture(t)
	units := loadApprovedLocalGovernmentUnitFixture(t)
	jsonRepo := &geographyJSONRepoStub{states: states, zones: zones, units: units, pathCalls: map[string]int{}}
	repo, err := NewGeographyRepository(jsonRepo, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}

	loaded, err := repo.ListLocalGovernmentUnits(context.Background())
	if err != nil {
		t.Fatalf("ListLocalGovernmentUnits() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("ListLocalGovernmentUnits() returned nil slice")
	}
	if len(loaded) != 774 {
		t.Fatalf("unexpected unit count: %d", len(loaded))
	}
	if !reflect.DeepEqual(loaded, units) {
		t.Fatalf("unexpected unit list: %#v", loaded)
	}
	if jsonRepo.pathCalls["geography/states.json"] != 1 || jsonRepo.pathCalls["geography/geopolitical_zones.json"] != 1 || jsonRepo.pathCalls["geography/lgas.json"] != 1 {
		t.Fatalf("unexpected per-path decode counts: %#v", jsonRepo.pathCalls)
	}

	loaded[0].Name = "Changed"
	again, err := repo.ListLocalGovernmentUnits(context.Background())
	if err != nil {
		t.Fatalf("ListLocalGovernmentUnits() second call error = %v", err)
	}
	if again[0].Name != units[0].Name {
		t.Fatal("ListLocalGovernmentUnits() shared mutable slice state")
	}

	fctUnits, err := repo.ListLocalGovernmentUnitsByStateID(context.Background(), "  fct  ")
	if err != nil {
		t.Fatalf("ListLocalGovernmentUnitsByStateID(fct) error = %v", err)
	}
	if len(fctUnits) != 6 {
		t.Fatalf("unexpected FCT unit count: %d", len(fctUnits))
	}
	if fctUnits[0].ID != "fct-abaji" || fctUnits[1].ID != "fct-abuja-municipal" {
		t.Fatalf("unexpected FCT ordering: %#v", fctUnits)
	}

	abiaUnits, err := repo.ListLocalGovernmentUnitsByStateID(context.Background(), "abia")
	if err != nil {
		t.Fatalf("ListLocalGovernmentUnitsByStateID(abia) error = %v", err)
	}
	if len(abiaUnits) != 17 {
		t.Fatalf("unexpected Abia unit count: %d", len(abiaUnits))
	}

	if _, err := repo.ListLocalGovernmentUnitsByStateID(context.Background(), "missing"); !errors.Is(err, interfaces.ErrStateNotFound) {
		t.Fatalf("missing state error = %v, want ErrStateNotFound", err)
	}
	if _, err := repo.ListLocalGovernmentUnitsByStateID(context.Background(), "Abia"); !errors.Is(err, interfaces.ErrStateNotFound) {
		t.Fatalf("invalid state error = %v, want ErrStateNotFound", err)
	}

	abujaMunicipal, err := repo.GetLocalGovernmentUnit(context.Background(), "fct-abuja-municipal")
	if err != nil {
		t.Fatalf("GetLocalGovernmentUnit(fct-abuja-municipal) error = %v", err)
	}
	if abujaMunicipal.ID != "fct-abuja-municipal" || abujaMunicipal.Name != "Abuja Municipal" {
		t.Fatalf("unexpected Abuja Municipal record: %#v", abujaMunicipal)
	}

	jamaare, err := repo.GetLocalGovernmentUnit(context.Background(), "bauchi-jama-are")
	if err != nil {
		t.Fatalf("GetLocalGovernmentUnit(bauchi-jama-are) error = %v", err)
	}
	if jamaare.Name != "Jama'are" {
		t.Fatalf("unexpected Jama'are record: %#v", jamaare)
	}

	if _, err := repo.GetLocalGovernmentUnit(context.Background(), "Abuja Municipal"); !errors.Is(err, interfaces.ErrLocalGovernmentUnitNotFound) {
		t.Fatalf("name lookup error = %v, want ErrLocalGovernmentUnitNotFound", err)
	}
	if _, err := repo.GetLocalGovernmentUnit(context.Background(), "123e4567-e89b-12d3-a456-426614174000"); !errors.Is(err, interfaces.ErrLocalGovernmentUnitNotFound) {
		t.Fatalf("uuid lookup error = %v, want ErrLocalGovernmentUnitNotFound", err)
	}
	if _, err := repo.GetLocalGovernmentUnit(context.Background(), "missing"); !errors.Is(err, interfaces.ErrLocalGovernmentUnitNotFound) {
		t.Fatalf("missing unit error = %v, want ErrLocalGovernmentUnitNotFound", err)
	}
}

func TestGeographyRepositoryLocalGovernmentUnitDecodeCounts(t *testing.T) {
	t.Parallel()

	fixtureStates := loadApprovedStateFixture(t)
	fixtureZones := loadApprovedZoneFixture(t)
	fixtureUnits := loadApprovedLocalGovernmentUnitFixture(t)

	tests := []struct {
		name string
		call func(*GeographyFileRepository) error
	}{
		{
			name: "list all",
			call: func(repo *GeographyFileRepository) error {
				_, err := repo.ListLocalGovernmentUnits(context.Background())
				return err
			},
		},
		{
			name: "list by state",
			call: func(repo *GeographyFileRepository) error {
				_, err := repo.ListLocalGovernmentUnitsByStateID(context.Background(), "fct")
				return err
			},
		},
		{
			name: "get by id",
			call: func(repo *GeographyFileRepository) error {
				_, err := repo.GetLocalGovernmentUnit(context.Background(), "fct-abuja-municipal")
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonRepo := &geographyJSONRepoStub{states: fixtureStates, zones: fixtureZones, units: fixtureUnits, pathCalls: map[string]int{}}
			repo, err := NewGeographyRepository(jsonRepo, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
			if err != nil {
				t.Fatalf("NewGeographyRepository() error = %v", err)
			}
			if err := tc.call(repo); err != nil {
				t.Fatalf("%s() error = %v", tc.name, err)
			}
			if jsonRepo.pathCalls["geography/states.json"] != 1 || jsonRepo.pathCalls["geography/geopolitical_zones.json"] != 1 || jsonRepo.pathCalls["geography/lgas.json"] != 1 {
				t.Fatalf("unexpected decode counts: %#v", jsonRepo.pathCalls)
			}
		})
	}
}

func TestGeographyRepositoryRejectsInvalidLocalGovernmentUnitFixtures(t *testing.T) {
	t.Parallel()

	fixtureStates := loadApprovedStateFixture(t)
	fixtureZones := loadApprovedZoneFixture(t)
	fixtureUnits := loadApprovedLocalGovernmentUnitFixture(t)

	type mutation func([]models.LocalGovernmentUnit) []models.LocalGovernmentUnit

	tests := []struct {
		name string
		mut  mutation
	}{
		{
			name: "nil slice",
			mut: func([]models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				return nil
			},
		},
		{
			name: "empty slice",
			mut: func([]models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				return make([]models.LocalGovernmentUnit, 0)
			},
		},
		{
			name: "773 records",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				return append([]models.LocalGovernmentUnit(nil), units[:773]...)
			},
		},
		{
			name: "775 records",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				return append(out, units[0])
			},
		},
		{
			name: "duplicate id",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[1].ID = out[0].ID
				return out
			},
		},
		{
			name: "duplicate state name pair",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[1].Name = out[0].Name
				out[1].ID = out[1].StateID + "-" + slugifyLocalGovernmentUnitName(out[1].Name)
				return out
			},
		},
		{
			name: "invalid state id",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[0].StateID = "invalid"
				out[0].ID = "invalid-" + slugifyLocalGovernmentUnitName(out[0].Name)
				return out
			},
		},
		{
			name: "missing parent state",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[0].StateID = "missing-state"
				out[0].ID = "missing-state-" + slugifyLocalGovernmentUnitName(out[0].Name)
				return out
			},
		},
		{
			name: "malformed id",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[0].ID = "bad"
				return out
			},
		},
		{
			name: "wrong state prefix",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[0].ID = "lagos-" + slugifyLocalGovernmentUnitName(out[0].Name)
				return out
			},
		},
		{
			name: "wrong slug",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[0].ID = out[0].StateID + "-wrong"
				return out
			},
		},
		{
			name: "invalid country code",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[0].CountryCode = "GH"
				return out
			},
		},
		{
			name: "unsupported administrative type",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[0].AdministrativeType = "district"
				return out
			},
		},
		{
			name: "fct lga type",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				for i := range out {
					if out[i].StateID == "fct" {
						out[i].AdministrativeType = "local_government_area"
						break
					}
				}
				return out
			},
		},
		{
			name: "non fct area council type",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[0].AdministrativeType = "area_council"
				return out
			},
		},
		{
			name: "missing canonical fct record",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				for i := range out {
					if out[i].ID == "fct-abuja-municipal" {
						out[i].Name = "AMAC"
						out[i].ID = "fct-amac"
						break
					}
				}
				return out
			},
		},
		{
			name: "unexpected fct record",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[0].StateID = "fct"
				out[0].AdministrativeType = "area_council"
				out[0].ID = "fct-extra"
				out[0].Name = "Extra"
				return out
			},
		},
		{
			name: "incorrect per state count",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				for i := range out {
					if out[i].StateID == "abia" {
						out[i].StateID = "lagos"
						out[i].ID = "lagos-" + slugifyLocalGovernmentUnitName(out[i].Name)
						break
					}
				}
				return out
			},
		},
		{
			name: "incorrect zone-derived count",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				for i := range out {
					if out[i].StateID == "lagos" {
						out[i].StateID = "benue"
						out[i].ID = "benue-" + slugifyLocalGovernmentUnitName(out[i].Name)
						break
					}
				}
				return out
			},
		},
		{
			name: "invalid ordering",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				out[0], out[1] = out[1], out[0]
				return out
			},
		},
		{
			name: "unreferenced state",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := make([]models.LocalGovernmentUnit, 0, len(units))
				for _, unit := range units {
					if unit.StateID == "yobe" {
						continue
					}
					out = append(out, unit)
				}
				for len(out) < len(units) {
					out = append(out, units[0])
				}
				return out[:len(units)]
			},
		},
		{
			name: "unreferenced zone",
			mut: func(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
				out := append([]models.LocalGovernmentUnit(nil), units...)
				for i := range out {
					if out[i].StateID == "kebbi" {
						out[i].StateID = "lagos"
						out[i].ID = "lagos-" + slugifyLocalGovernmentUnitName(out[i].Name)
						break
					}
				}
				return out
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newGeographyRepoWithLocalGovernmentUnits(t, tc.mut(cloneTestUnits(fixtureUnits)), fixtureStates, fixtureZones)
			if _, err := repo.ListLocalGovernmentUnits(context.Background()); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("ListLocalGovernmentUnits() error = %v, want ErrInvalidDatasetFile", err)
			}
		})
	}
}

func TestGeographyRepositoryRejectsNilAndEmptyDecodedSlices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		repo *GeographyFileRepository
	}{
		{
			name: "nil decoded slice",
			repo: newGeographyRepoWithDecodeFn(t, func(_ context.Context, _ string, destination any) error {
				switch dest := destination.(type) {
				case *[]models.State:
					*dest = nil
				case *[]models.GeopoliticalZone:
					*dest = nil
				default:
					return fmt.Errorf("unexpected destination %T", destination)
				}
				return nil
			}),
		},
		{
			name: "empty decoded slice",
			repo: newGeographyRepoWithDecodeFn(t, func(_ context.Context, _ string, destination any) error {
				switch dest := destination.(type) {
				case *[]models.State:
					*dest = make([]models.State, 0)
				case *[]models.GeopoliticalZone:
					*dest = make([]models.GeopoliticalZone, 0)
				default:
					return fmt.Errorf("unexpected destination %T", destination)
				}
				return nil
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			states, err := tc.repo.ListStates(context.Background())
			if err == nil {
				t.Fatalf("ListStates() states = %#v, want error", states)
			}
			if !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("ListStates() error = %v, want ErrInvalidDatasetFile", err)
			}
		})
	}
}

func TestGeographyRepositoryContextAndDecodeErrors(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedStateFixture(t)
	zones := loadApprovedZoneFixture(t)

	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		stub := &geographyJSONRepoStub{
			states: fixture,
			zones:  zones,
			decodeFn: func(context.Context, string, any) error {
				t.Fatal("decode should not be called for canceled context")
				return nil
			},
		}
		repo, err := NewGeographyRepository(stub, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
		if err != nil {
			t.Fatalf("NewGeographyRepository() error = %v", err)
		}
		if _, err := repo.ListStates(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("ListStates() error = %v, want context.Canceled", err)
		}
	})

	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Minute))
		defer cancel()
		stub := &geographyJSONRepoStub{
			states: fixture,
			zones:  zones,
			decodeFn: func(context.Context, string, any) error {
				t.Fatal("decode should not be called for expired deadline")
				return nil
			},
		}
		repo, err := NewGeographyRepository(stub, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
		if err != nil {
			t.Fatalf("NewGeographyRepository() error = %v", err)
		}
		if _, err := repo.ListStates(ctx); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("ListStates() error = %v, want context.DeadlineExceeded", err)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		repo := newGeographyRepoForError(t, interfaces.ErrDatasetFileNotFound)
		if _, err := repo.ListStates(context.Background()); !errors.Is(err, interfaces.ErrDatasetFileNotFound) {
			t.Fatalf("ListStates() error = %v, want ErrDatasetFileNotFound", err)
		}
	})

	t.Run("unavailable file", func(t *testing.T) {
		repo := newGeographyRepoForError(t, interfaces.ErrDatasetFileUnavailable)
		if _, err := repo.ListStates(context.Background()); !errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
			t.Fatalf("ListStates() error = %v, want ErrDatasetFileUnavailable", err)
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		repo := newGeographyRepoForError(t, interfaces.ErrInvalidDatasetFile)
		if _, err := repo.ListStates(context.Background()); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
			t.Fatalf("ListStates() error = %v, want ErrInvalidDatasetFile", err)
		}
	})

	t.Run("lgas", func(t *testing.T) {
		repo := newGeographyRepoForError(t, interfaces.ErrDatasetFileUnavailable)
		if _, err := repo.ListLocalGovernmentUnits(context.Background()); !errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
			t.Fatalf("ListLocalGovernmentUnits() error = %v, want ErrDatasetFileUnavailable", err)
		}
	})
}

func TestGeographyRepositoryRuntimeValidation(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedStateFixture(t)
	tests := []struct {
		name string
		mut  func([]models.State) []models.State
	}{
		{
			name: "duplicate ids",
			mut: func(states []models.State) []models.State {
				states[1].ID = states[0].ID
				return states
			},
		},
		{
			name: "duplicate names",
			mut: func(states []models.State) []models.State {
				states[1].Name = states[0].Name
				return states
			},
		},
		{
			name: "missing required fields",
			mut: func(states []models.State) []models.State {
				states[0].Capital = ""
				return states
			},
		},
		{
			name: "invalid administrative type",
			mut: func(states []models.State) []models.State {
				states[0].AdministrativeType = "district"
				return states
			},
		},
		{
			name: "invalid zone",
			mut: func(states []models.State) []models.State {
				states[0].GeopoliticalZoneID = "central"
				return states
			},
		},
		{
			name: "invalid country code",
			mut: func(states []models.State) []models.State {
				states[0].CountryCode = "GH"
				return states
			},
		},
		{
			name: "missing fct",
			mut: func(states []models.State) []models.State {
				out := make([]models.State, 0, len(states)-1)
				for _, state := range states {
					if state.ID != "fct" {
						out = append(out, state)
					}
				}
				return out
			},
		},
		{
			name: "multiple fct records",
			mut: func(states []models.State) []models.State {
				states = append(states, states[14])
				return states
			},
		},
		{
			name: "incorrect fct id",
			mut: func(states []models.State) []models.State {
				states[14].ID = "abuja"
				return states
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newGeographyRepoWithStates(t, tc.mut(cloneTestStates(fixture)))
			if _, err := repo.ListStates(context.Background()); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("ListStates() error = %v, want ErrInvalidDatasetFile", err)
			}
		})
	}
}

func TestGeographyRepositoryRejectsInvalidRecordCountsAndFCTComposition(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedStateFixture(t)
	tests := []struct {
		name string
		mut  func([]models.State) []models.State
	}{
		{
			name: "36 total records",
			mut: func(states []models.State) []models.State {
				return states[:36]
			},
		},
		{
			name: "38 total records",
			mut: func(states []models.State) []models.State {
				return append(states, states[0])
			},
		},
		{
			name: "37 records with only 35 state records",
			mut: func(states []models.State) []models.State {
				out := make([]models.State, 0, 37)
				stateRemoved := false
				for _, state := range states {
					if state.AdministrativeType == "state" && !stateRemoved {
						stateRemoved = true
						continue
					}
					out = append(out, state)
				}
				out = append(out, models.State{
					ID:                 "fct-alt",
					Name:               "Federal Capital Territory Alt",
					OfficialName:       "Federal Capital Territory Alt",
					AdministrativeType: "federal_capital_territory",
					Capital:            "Abuja",
					GeopoliticalZoneID: "north-central",
					CountryCode:        "NG",
				})
				return out
			},
		},
		{
			name: "37 records with multiple fct records",
			mut: func(states []models.State) []models.State {
				out := make([]models.State, 0, 37)
				removed := 0
				for _, state := range states {
					if state.AdministrativeType == "state" && removed < 2 {
						removed++
						continue
					}
					out = append(out, state)
				}
				out = append(out,
					models.State{
						ID:                 "fct-alt-1",
						Name:               "Federal Capital Territory Alt 1",
						OfficialName:       "Federal Capital Territory Alt 1",
						AdministrativeType: "federal_capital_territory",
						Capital:            "Abuja",
						GeopoliticalZoneID: "north-central",
						CountryCode:        "NG",
					},
					models.State{
						ID:                 "fct-alt-2",
						Name:               "Federal Capital Territory Alt 2",
						OfficialName:       "Federal Capital Territory Alt 2",
						AdministrativeType: "federal_capital_territory",
						Capital:            "Abuja",
						GeopoliticalZoneID: "north-central",
						CountryCode:        "NG",
					},
					models.State{
						ID:                 "fct-alt-3",
						Name:               "Federal Capital Territory Alt 3",
						OfficialName:       "Federal Capital Territory Alt 3",
						AdministrativeType: "federal_capital_territory",
						Capital:            "Abuja",
						GeopoliticalZoneID: "north-central",
						CountryCode:        "NG",
					},
				)
				return out[:37]
			},
		},
		{
			name: "37 records with no fct",
			mut: func(states []models.State) []models.State {
				out := make([]models.State, 0, 37)
				for _, state := range states {
					if state.ID == "fct" {
						continue
					}
					out = append(out, state)
				}
				out = append(out, states[0])
				return out[:37]
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newGeographyRepoWithStates(t, tc.mut(cloneTestStates(fixture)))
			if _, err := repo.ListStates(context.Background()); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("ListStates() error = %v, want ErrInvalidDatasetFile", err)
			}
		})
	}
}

func TestGeographyRepositoryErrorsDoNotExposePathOrRecordContents(t *testing.T) {
	t.Parallel()

	repo := newGeographyRepoWithStates(t, loadApprovedStateFixture(t))
	repo.statesPath = "../../../datasets/geography/states.json"

	states, err := repo.ListStates(context.Background())
	if err != nil {
		t.Fatalf("ListStates() error = %v", err)
	}
	if len(states) == 0 {
		t.Fatal("expected non-empty fixture")
	}

	badRepo := newGeographyRepoWithStates(t, []models.State{{ID: "bad", Name: "Abia", OfficialName: "Abia State", AdministrativeType: "state", Capital: "Umuahia", GeopoliticalZoneID: "south-east", CountryCode: "NG"}, {ID: "bad", Name: "Bad", OfficialName: "Bad State", AdministrativeType: "state", Capital: "Bad", GeopoliticalZoneID: "south-east", CountryCode: "NG"}})
	_, err = badRepo.ListStates(context.Background())
	if err == nil {
		t.Fatal("expected validation error")
	}
	if strings.Contains(err.Error(), "Abia") || strings.Contains(err.Error(), "states.json") || strings.Contains(err.Error(), "bad") {
		t.Fatalf("error exposed unsafe details: %v", err)
	}

	lgaUnits := loadApprovedLocalGovernmentUnitFixture(t)
	lgaUnits[0].CountryCode = "GH"
	lgaRepo := newGeographyRepoWithLocalGovernmentUnits(t, lgaUnits, loadApprovedStateFixture(t), loadApprovedZoneFixture(t))
	lgaRepo.localGovernmentUnitsPath = "../../../datasets/geography/lgas.json"
	_, err = lgaRepo.ListLocalGovernmentUnits(context.Background())
	if err == nil {
		t.Fatal("expected lga validation error")
	}
	if strings.Contains(err.Error(), "lgas.json") || strings.Contains(err.Error(), "Abaji") || strings.Contains(err.Error(), "../../../datasets") {
		t.Fatalf("error exposed unsafe details: %v", err)
	}
}

func newGeographyRepoForError(t *testing.T, decodeErr error) *GeographyFileRepository {
	t.Helper()

	repo, err := NewGeographyRepository(&geographyJSONRepoStub{decodeFn: func(context.Context, string, any) error { return decodeErr }}, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	return repo
}

func newGeographyRepoWithStates(t *testing.T, states []models.State) *GeographyFileRepository {
	t.Helper()

	repo, err := NewGeographyRepository(&geographyJSONRepoStub{states: states, zones: loadApprovedZoneFixture(t)}, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	return repo
}

func newGeographyRepoWithDecodeFn(t *testing.T, decodeFn func(context.Context, string, any) error) *GeographyFileRepository {
	t.Helper()

	repo, err := NewGeographyRepository(&geographyJSONRepoStub{decodeFn: decodeFn}, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	return repo
}

func newGeographyRepoWithLocalGovernmentUnits(t *testing.T, units []models.LocalGovernmentUnit, states []models.State, zones []models.GeopoliticalZone) *GeographyFileRepository {
	t.Helper()

	repo, err := NewGeographyRepository(&geographyJSONRepoStub{states: states, zones: zones, units: units}, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	return repo
}

func loadApprovedStateFixture(t *testing.T) []models.State {
	t.Helper()

	path := filepath.Clean("../../../datasets/geography/states.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read states fixture: %v", err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var states []models.State
	if err := dec.Decode(&states); err != nil {
		t.Fatalf("decode states fixture: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("fixture contains trailing json")
	}
	return states
}

func loadApprovedZoneFixture(t *testing.T) []models.GeopoliticalZone {
	t.Helper()

	path := filepath.Clean("../../../datasets/geography/geopolitical_zones.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read geopolitical zones fixture: %v", err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var zones []models.GeopoliticalZone
	if err := dec.Decode(&zones); err != nil {
		t.Fatalf("decode geopolitical zones fixture: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("zone fixture contains trailing json")
	}
	return zones
}

func loadApprovedLocalGovernmentUnitFixture(t *testing.T) []models.LocalGovernmentUnit {
	t.Helper()

	path := filepath.Clean("../../../datasets/geography/lgas.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read local government units fixture: %v", err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var units []models.LocalGovernmentUnit
	if err := dec.Decode(&units); err != nil {
		t.Fatalf("decode local government units fixture: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("local government unit fixture contains trailing json")
	}
	return units
}

func loadApprovedCountryFixture(t *testing.T) []models.CountryOrArea {
	t.Helper()

	path := filepath.Clean("../../../datasets/geography/countries_and_areas.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read countries and areas fixture: %v", err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var countries []models.CountryOrArea
	if err := dec.Decode(&countries); err != nil {
		t.Fatalf("decode countries and areas fixture: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("countries and areas fixture contains trailing json")
	}
	return countries
}

func cloneTestStates(states []models.State) []models.State {
	if len(states) == 0 {
		return make([]models.State, 0)
	}
	cloned := make([]models.State, len(states))
	copy(cloned, states)
	return cloned
}

func cloneTestZones(zones []models.GeopoliticalZone) []models.GeopoliticalZone {
	if len(zones) == 0 {
		return make([]models.GeopoliticalZone, 0)
	}
	cloned := make([]models.GeopoliticalZone, len(zones))
	copy(cloned, zones)
	return cloned
}

func cloneTestUnits(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
	if len(units) == 0 {
		return make([]models.LocalGovernmentUnit, 0)
	}
	cloned := make([]models.LocalGovernmentUnit, len(units))
	copy(cloned, units)
	return cloned
}

func cloneTestCountries(countries []models.CountryOrArea) []models.CountryOrArea {
	if len(countries) == 0 {
		return make([]models.CountryOrArea, 0)
	}
	cloned := make([]models.CountryOrArea, len(countries))
	copy(cloned, countries)
	return cloned
}

func cloneTestTimeZones(timeZones []models.TimeZone) []models.TimeZone {
	if len(timeZones) == 0 {
		return make([]models.TimeZone, 0)
	}
	cloned := make([]models.TimeZone, len(timeZones))
	for i, timeZone := range timeZones {
		cloned[i] = timeZone
		cloned[i].CountryAreaIDs = cloneTestCountryAreaIDs(timeZone.CountryAreaIDs)
	}
	return cloned
}

func cloneTestCountryAreaIDs(ids []string) []string {
	if len(ids) == 0 {
		return make([]string, 0)
	}
	cloned := make([]string, len(ids))
	copy(cloned, ids)
	return cloned
}
