package services

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type geographyRepoStub struct {
	states            []models.State
	zones             []models.GeopoliticalZone
	units             []models.LocalGovernmentUnit
	timeZones         []models.TimeZone
	countries         []models.CountryOrArea
	listErr           error
	getErr            error
	zoneListErr       error
	zoneGetErr        error
	unitListErr       error
	unitGetErr        error
	unitByStateErr    error
	timeZoneListErr   error
	timeZoneGetErr    error
	countryListErr    error
	countryGetErr     error
	listCalls         int
	getCalls          int
	zoneListCalls     int
	zoneGetCalls      int
	unitListCalls     int
	unitGetCalls      int
	unitByStateCalls  int
	timeZoneListCalls int
	timeZoneGetCalls  int
	countryListCalls  int
	countryGetCalls   int
	lastID            string
	lastZoneID        string
	lastStateID       string
	lastUnitID        string
	lastCountryID     string
	lastTimeZoneID    string
}

func (s *geographyRepoStub) ListStates(context.Context) ([]models.State, error) {
	s.listCalls++
	if s.listErr != nil {
		return nil, s.listErr
	}
	return cloneServiceStates(s.states), nil
}

func (s *geographyRepoStub) GetStateByID(_ context.Context, stateID string) (models.State, error) {
	s.getCalls++
	s.lastID = stateID
	if s.getErr != nil {
		return models.State{}, s.getErr
	}
	for _, state := range s.states {
		if state.ID == stateID {
			return state, nil
		}
	}
	return models.State{}, interfaces.ErrStateNotFound
}

func (s *geographyRepoStub) ListGeopoliticalZones(context.Context) ([]models.GeopoliticalZone, error) {
	s.zoneListCalls++
	if s.zoneListErr != nil {
		return nil, s.zoneListErr
	}
	return cloneServiceZones(s.zones), nil
}

func (s *geographyRepoStub) GetGeopoliticalZone(_ context.Context, zoneID string) (models.GeopoliticalZone, error) {
	s.zoneGetCalls++
	s.lastZoneID = zoneID
	if s.zoneGetErr != nil {
		return models.GeopoliticalZone{}, s.zoneGetErr
	}
	for _, zone := range s.zones {
		if zone.ID == zoneID {
			return zone, nil
		}
	}
	return models.GeopoliticalZone{}, interfaces.ErrGeopoliticalZoneNotFound
}

func (s *geographyRepoStub) ListLocalGovernmentUnits(context.Context) ([]models.LocalGovernmentUnit, error) {
	s.unitListCalls++
	if s.unitListErr != nil {
		return nil, s.unitListErr
	}
	return cloneServiceLocalGovernmentUnits(s.units), nil
}

func (s *geographyRepoStub) ListLocalGovernmentUnitsByStateID(_ context.Context, stateID string) ([]models.LocalGovernmentUnit, error) {
	s.unitByStateCalls++
	s.lastStateID = stateID
	if s.unitByStateErr != nil {
		return nil, s.unitByStateErr
	}
	if stateID == "" {
		return nil, interfaces.ErrStateNotFound
	}
	filtered := make([]models.LocalGovernmentUnit, 0)
	for _, unit := range s.units {
		if unit.StateID == stateID {
			filtered = append(filtered, unit)
		}
	}
	if len(filtered) == 0 {
		return nil, interfaces.ErrStateNotFound
	}
	return cloneServiceLocalGovernmentUnits(filtered), nil
}

func (s *geographyRepoStub) GetLocalGovernmentUnit(_ context.Context, unitID string) (models.LocalGovernmentUnit, error) {
	s.unitGetCalls++
	s.lastUnitID = unitID
	if s.unitGetErr != nil {
		return models.LocalGovernmentUnit{}, s.unitGetErr
	}
	for _, unit := range s.units {
		if unit.ID == unitID {
			return unit, nil
		}
	}
	return models.LocalGovernmentUnit{}, interfaces.ErrLocalGovernmentUnitNotFound
}

func (s *geographyRepoStub) ListTimeZones(_ context.Context, filter interfaces.TimeZoneFilter) ([]models.TimeZone, error) {
	s.timeZoneListCalls++
	if s.timeZoneListErr != nil {
		return nil, s.timeZoneListErr
	}
	if s.timeZones == nil {
		return nil, nil
	}
	filtered := make([]models.TimeZone, 0, len(s.timeZones))
	for _, timeZone := range s.timeZones {
		if filter.CountryAreaID != "" && !containsString(timeZone.CountryAreaIDs, filter.CountryAreaID) {
			continue
		}
		filtered = append(filtered, cloneServiceTimeZone(timeZone))
	}
	return filtered, nil
}

func (s *geographyRepoStub) GetTimeZone(_ context.Context, timeZoneID string) (models.TimeZone, error) {
	s.timeZoneGetCalls++
	s.lastTimeZoneID = timeZoneID
	if s.timeZoneGetErr != nil {
		return models.TimeZone{}, s.timeZoneGetErr
	}
	for _, timeZone := range s.timeZones {
		if timeZone.ID == timeZoneID {
			return cloneServiceTimeZone(timeZone), nil
		}
	}
	return models.TimeZone{}, interfaces.ErrTimeZoneNotFound
}

func (s *geographyRepoStub) ListCountriesAndAreas(_ context.Context, filter interfaces.CountryOrAreaFilter) ([]models.CountryOrArea, error) {
	s.countryListCalls++
	if s.countryListErr != nil {
		return nil, s.countryListErr
	}
	if s.countries == nil {
		return nil, nil
	}
	return cloneServiceCountries(s.filterCountries(filter)), nil
}

func (s *geographyRepoStub) GetCountryOrArea(_ context.Context, countryOrAreaID string) (models.CountryOrArea, error) {
	s.countryGetCalls++
	s.lastCountryID = countryOrAreaID
	if s.countryGetErr != nil {
		return models.CountryOrArea{}, s.countryGetErr
	}
	for _, country := range s.countries {
		if country.ID == countryOrAreaID {
			return country, nil
		}
	}
	return models.CountryOrArea{}, interfaces.ErrCountryOrAreaNotFound
}

func containsString(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func cloneServiceTimeZone(timeZone models.TimeZone) models.TimeZone {
	cloned := timeZone
	cloned.CountryAreaIDs = append([]string(nil), timeZone.CountryAreaIDs...)
	return cloned
}

func (s *geographyRepoStub) filterCountries(filter interfaces.CountryOrAreaFilter) []models.CountryOrArea {
	filtered := make([]models.CountryOrArea, 0, len(s.countries))
	for _, country := range s.countries {
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

func TestNewGeographyServiceRejectsNilRepository(t *testing.T) {
	t.Parallel()

	if _, err := NewGeographyService(nil); err == nil {
		t.Fatal("expected nil repository to be rejected")
	}
}

func TestGeographyServiceListStates(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{states: loadServiceStateFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	states, err := svc.ListStates(context.Background())
	if err != nil {
		t.Fatalf("ListStates() error = %v", err)
	}
	if stub.listCalls != 1 {
		t.Fatalf("unexpected repository call count: %d", stub.listCalls)
	}
	if states == nil {
		t.Fatal("ListStates() returned nil slice")
	}
	if len(states) == 0 {
		t.Fatal("expected non-empty state list")
	}
	if states[0].ID != "abia" || states[len(states)-1].ID != "zamfara" {
		t.Fatalf("unexpected ordering: first=%q last=%q", states[0].ID, states[len(states)-1].ID)
	}

	states[0].Name = "Changed"
	again, err := svc.ListStates(context.Background())
	if err != nil {
		t.Fatalf("ListStates() second call error = %v", err)
	}
	if again[0].Name != "Abia" {
		t.Fatal("ListStates() exposed shared mutable slice state")
	}
}

func TestGeographyServiceListStatesNilRepositorySliceBecomesEmpty(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{states: nil}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	states, err := svc.ListStates(context.Background())
	if err != nil {
		t.Fatalf("ListStates() error = %v", err)
	}
	if states == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(states) != 0 {
		t.Fatalf("unexpected length: %d", len(states))
	}
}

func TestGeographyServiceGetState(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{states: loadServiceStateFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	state, err := svc.GetState(context.Background(), "  abia  ")
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if stub.getCalls != 1 {
		t.Fatalf("unexpected repository call count: %d", stub.getCalls)
	}
	if stub.lastID != "abia" {
		t.Fatalf("unexpected repository lookup id: %q", stub.lastID)
	}
	if state.ID != "abia" || state.Name != "Abia" {
		t.Fatalf("unexpected state response: %#v", state)
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(fields) != 7 {
		t.Fatalf("unexpected json field count: %#v", fields)
	}
	for _, key := range []string{"id", "name", "official_name", "administrative_type", "capital", "geopolitical_zone_id", "country_code"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("missing json field %q", key)
		}
	}

	state, err = svc.GetState(context.Background(), "fct")
	if err != nil {
		t.Fatalf("GetState(fct) error = %v", err)
	}
	if state.ID != "fct" || state.Capital != "Abuja" {
		t.Fatalf("unexpected FCT state: %#v", state)
	}
}

func TestGeographyServiceListGeopoliticalZones(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{zones: loadServiceGeopoliticalZoneFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	zones, err := svc.ListGeopoliticalZones(context.Background())
	if err != nil {
		t.Fatalf("ListGeopoliticalZones() error = %v", err)
	}
	if stub.zoneListCalls != 1 {
		t.Fatalf("unexpected zone list call count: %d", stub.zoneListCalls)
	}
	if len(zones) != 6 {
		t.Fatalf("unexpected zone count: %d", len(zones))
	}
	zones[0].Name = "Changed"
	again, err := svc.ListGeopoliticalZones(context.Background())
	if err != nil {
		t.Fatalf("ListGeopoliticalZones() second call error = %v", err)
	}
	if again[0].Name != "North Central" {
		t.Fatal("ListGeopoliticalZones() exposed shared mutable slice state")
	}
}

func TestGeographyServiceGetGeopoliticalZone(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{zones: loadServiceGeopoliticalZoneFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	zone, err := svc.GetGeopoliticalZone(context.Background(), "  north-central  ")
	if err != nil {
		t.Fatalf("GetGeopoliticalZone() error = %v", err)
	}
	if stub.zoneGetCalls != 1 {
		t.Fatalf("unexpected zone get call count: %d", stub.zoneGetCalls)
	}
	if stub.lastZoneID != "north-central" {
		t.Fatalf("unexpected zone lookup id: %q", stub.lastZoneID)
	}
	if zone.ID != "north-central" || zone.Name != "North Central" {
		t.Fatalf("unexpected zone response: %#v", zone)
	}
}

func TestGeographyServiceRejectsInvalidGeopoliticalZoneIdentifiers(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{zones: loadServiceGeopoliticalZoneFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "uppercase", value: "North Central"},
		{name: "malformed", value: "../north-central"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.zoneGetCalls = 0
			_, err := svc.GetGeopoliticalZone(context.Background(), tc.value)
			if !errors.Is(err, ErrInvalidGeopoliticalZoneID) {
				t.Fatalf("GetGeopoliticalZone() error = %v, want ErrInvalidGeopoliticalZoneID", err)
			}
			if stub.zoneGetCalls != 0 {
				t.Fatalf("repository should not have been called, got %d", stub.zoneGetCalls)
			}
		})
	}
}

func TestGeographyServiceMissingGeopoliticalZoneAndUnexpectedErrors(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{zones: loadServiceGeopoliticalZoneFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	stub.zoneGetErr = interfaces.ErrGeopoliticalZoneNotFound
	if _, err := svc.GetGeopoliticalZone(context.Background(), "missing"); !errors.Is(err, ErrGeopoliticalZoneNotFound) {
		t.Fatalf("missing zone error = %v, want ErrGeopoliticalZoneNotFound", err)
	}

	stub.zoneGetErr = errors.New("boom")
	if _, err := svc.GetGeopoliticalZone(context.Background(), "north-central"); err == nil || strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected repository error was not sanitized: %v", err)
	}

	stub.zoneListErr = errors.New("explode")
	if _, err := svc.ListGeopoliticalZones(context.Background()); err == nil || strings.Contains(err.Error(), "explode") {
		t.Fatalf("unexpected zone list error was not sanitized: %v", err)
	}
}

func TestGeographyServiceLocalGovernmentUnits(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{
		states: loadServiceStateFixture(t),
		zones:  loadServiceGeopoliticalZoneFixture(t),
		units:  loadServiceLocalGovernmentUnitFixture(t),
	}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	units, err := svc.ListLocalGovernmentUnits(context.Background())
	if err != nil {
		t.Fatalf("ListLocalGovernmentUnits() error = %v", err)
	}
	if stub.unitListCalls != 1 {
		t.Fatalf("unexpected unit list call count: %d", stub.unitListCalls)
	}
	if units == nil {
		t.Fatal("ListLocalGovernmentUnits() returned nil slice")
	}
	if len(units) != 774 {
		t.Fatalf("unexpected unit count: %d", len(units))
	}
	units[0].Name = "Changed"
	again, err := svc.ListLocalGovernmentUnits(context.Background())
	if err != nil {
		t.Fatalf("ListLocalGovernmentUnits() second call error = %v", err)
	}
	if again[0].Name == "Changed" {
		t.Fatal("ListLocalGovernmentUnits() exposed shared mutable slice state")
	}

	byState, err := svc.ListLocalGovernmentUnitsByState(context.Background(), "  fct  ")
	if err != nil {
		t.Fatalf("ListLocalGovernmentUnitsByState() error = %v", err)
	}
	if stub.unitByStateCalls != 1 {
		t.Fatalf("unexpected unit by state call count: %d", stub.unitByStateCalls)
	}
	if stub.lastStateID != "fct" {
		t.Fatalf("unexpected state lookup id: %q", stub.lastStateID)
	}
	if len(byState) != 6 {
		t.Fatalf("unexpected FCT unit count: %d", len(byState))
	}
	if byState[0].ID != "fct-abaji" {
		t.Fatalf("unexpected FCT ordering: %#v", byState)
	}
	byState[0].Name = "Changed"
	againByState, err := svc.ListLocalGovernmentUnitsByState(context.Background(), "fct")
	if err != nil {
		t.Fatalf("ListLocalGovernmentUnitsByState() second call error = %v", err)
	}
	if againByState[0].Name == "Changed" {
		t.Fatal("ListLocalGovernmentUnitsByState() exposed shared mutable slice state")
	}

	unit, err := svc.GetLocalGovernmentUnit(context.Background(), "  fct-abuja-municipal  ")
	if err != nil {
		t.Fatalf("GetLocalGovernmentUnit() error = %v", err)
	}
	if stub.unitGetCalls != 1 {
		t.Fatalf("unexpected unit get call count: %d", stub.unitGetCalls)
	}
	if stub.lastUnitID != "fct-abuja-municipal" {
		t.Fatalf("unexpected unit lookup id: %q", stub.lastUnitID)
	}
	if unit.ID != "fct-abuja-municipal" || unit.Name != "Abuja Municipal" {
		t.Fatalf("unexpected unit response: %#v", unit)
	}

	data, err := json.Marshal(unit)
	if err != nil {
		t.Fatalf("json.Marshal(unit): %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("json.Unmarshal(unit): %v", err)
	}
	if len(fields) != 5 {
		t.Fatalf("unexpected unit field count: %#v", fields)
	}
	for _, key := range []string{"id", "name", "state_id", "country_code", "administrative_type"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("missing unit field %q", key)
		}
	}
}

func TestGeographyServiceRejectsInvalidLocalGovernmentUnitIdentifiers(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{
		states: loadServiceStateFixture(t),
		zones:  loadServiceGeopoliticalZoneFixture(t),
		units:  loadServiceLocalGovernmentUnitFixture(t),
	}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "uppercase", value: "FCT-ABUJA-MUNICIPAL"},
		{name: "name", value: "Abuja Municipal"},
		{name: "underscore", value: "fct_abuja_municipal"},
		{name: "slash", value: "fct/abuja-municipal"},
		{name: "malformed", value: "../fct-abuja-municipal"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.unitGetCalls = 0
			_, err := svc.GetLocalGovernmentUnit(context.Background(), tc.value)
			if !errors.Is(err, ErrInvalidLocalGovernmentUnitID) {
				t.Fatalf("GetLocalGovernmentUnit() error = %v, want ErrInvalidLocalGovernmentUnitID", err)
			}
			if stub.unitGetCalls != 0 {
				t.Fatalf("repository should not have been called, got %d", stub.unitGetCalls)
			}
		})
	}
}

func TestGeographyServiceMissingLocalGovernmentUnitAndUnexpectedErrors(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{
		states: loadServiceStateFixture(t),
		zones:  loadServiceGeopoliticalZoneFixture(t),
		units:  loadServiceLocalGovernmentUnitFixture(t),
	}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	stub.unitGetErr = interfaces.ErrLocalGovernmentUnitNotFound
	if _, err := svc.GetLocalGovernmentUnit(context.Background(), "fct-missing"); !errors.Is(err, ErrLocalGovernmentUnitNotFound) {
		t.Fatalf("missing unit error = %v, want ErrLocalGovernmentUnitNotFound", err)
	}

	stub.unitGetErr = errors.New("boom")
	if _, err := svc.GetLocalGovernmentUnit(context.Background(), "fct-abaji"); err == nil || strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected unit error was not sanitized: %v", err)
	}

	stub.unitListErr = errors.New("explode")
	if _, err := svc.ListLocalGovernmentUnits(context.Background()); err == nil || strings.Contains(err.Error(), "explode") {
		t.Fatalf("unexpected unit list error was not sanitized: %v", err)
	}

	stub.unitByStateErr = errors.New("zap")
	if _, err := svc.ListLocalGovernmentUnitsByState(context.Background(), "fct"); err == nil || strings.Contains(err.Error(), "zap") {
		t.Fatalf("unexpected unit-by-state error was not sanitized: %v", err)
	}
}

func TestGeographyServiceLocalGovernmentUnitContextCancellationAndDeadline(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{
		states: loadServiceStateFixture(t),
		zones:  loadServiceGeopoliticalZoneFixture(t),
		units:  loadServiceLocalGovernmentUnitFixture(t),
	}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.ListLocalGovernmentUnits(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("ListLocalGovernmentUnits() error = %v, want context.Canceled", err)
	}
	if _, err := svc.ListLocalGovernmentUnitsByState(ctx, "fct"); !errors.Is(err, context.Canceled) {
		t.Fatalf("ListLocalGovernmentUnitsByState() error = %v, want context.Canceled", err)
	}
	if _, err := svc.GetLocalGovernmentUnit(ctx, "fct-abaji"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetLocalGovernmentUnit() error = %v, want context.Canceled", err)
	}

	deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Minute))
	defer deadlineCancel()
	if _, err := svc.ListLocalGovernmentUnits(deadlineCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ListLocalGovernmentUnits() error = %v, want context.DeadlineExceeded", err)
	}
	if _, err := svc.ListLocalGovernmentUnitsByState(deadlineCtx, "fct"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ListLocalGovernmentUnitsByState() error = %v, want context.DeadlineExceeded", err)
	}
	if _, err := svc.GetLocalGovernmentUnit(deadlineCtx, "fct-abaji"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetLocalGovernmentUnit() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestGeographyServiceRejectsInvalidIdentifiers(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{states: loadServiceStateFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "uppercase", value: "Abia"},
		{name: "malformed", value: "../abia"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.getCalls = 0
			_, err := svc.GetState(context.Background(), tc.value)
			if !errors.Is(err, ErrInvalidStateID) {
				t.Fatalf("GetState() error = %v, want ErrInvalidStateID", err)
			}
			if stub.getCalls != 0 {
				t.Fatalf("repository should not have been called, got %d", stub.getCalls)
			}
		})
	}
}

func TestGeographyServiceMissingStateAndUnexpectedErrors(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{states: loadServiceStateFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	stub.getErr = interfaces.ErrStateNotFound
	if _, err := svc.GetState(context.Background(), "missing"); !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("missing state error = %v, want ErrStateNotFound", err)
	}

	stub.getErr = errors.New("boom")
	if _, err := svc.GetState(context.Background(), "abia"); err == nil || strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected repository error was not sanitized: %v", err)
	}

	stub.listErr = errors.New("explode")
	if _, err := svc.ListStates(context.Background()); err == nil || strings.Contains(err.Error(), "explode") {
		t.Fatalf("unexpected list error was not sanitized: %v", err)
	}
}

func TestGeographyServiceContextCancellationAndDeadline(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{states: loadServiceStateFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.ListStates(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("ListStates() error = %v, want context.Canceled", err)
	}
	if _, err := svc.GetState(ctx, "abia"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetState() error = %v, want context.Canceled", err)
	}

	deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Minute))
	defer deadlineCancel()
	if _, err := svc.ListStates(deadlineCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ListStates() error = %v, want context.DeadlineExceeded", err)
	}
	if _, err := svc.GetState(deadlineCtx, "abia"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetState() error = %v, want context.DeadlineExceeded", err)
	}
}

func cloneServiceStates(states []models.State) []models.State {
	if len(states) == 0 {
		return make([]models.State, 0)
	}
	cloned := make([]models.State, len(states))
	copy(cloned, states)
	return cloned
}

func cloneServiceZones(zones []models.GeopoliticalZone) []models.GeopoliticalZone {
	if len(zones) == 0 {
		return make([]models.GeopoliticalZone, 0)
	}
	cloned := make([]models.GeopoliticalZone, len(zones))
	copy(cloned, zones)
	return cloned
}

func cloneServiceLocalGovernmentUnits(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
	if len(units) == 0 {
		return make([]models.LocalGovernmentUnit, 0)
	}
	cloned := make([]models.LocalGovernmentUnit, len(units))
	copy(cloned, units)
	return cloned
}

func cloneServiceCountries(countries []models.CountryOrArea) []models.CountryOrArea {
	if len(countries) == 0 {
		return make([]models.CountryOrArea, 0)
	}
	cloned := make([]models.CountryOrArea, len(countries))
	copy(cloned, countries)
	return cloned
}

func loadServiceStateFixture(t *testing.T) []models.State {
	t.Helper()

	data, err := os.ReadFile("../../datasets/geography/states.json")
	if err != nil {
		t.Fatalf("read states fixture: %v", err)
	}

	var states []models.State
	if err := json.Unmarshal(data, &states); err != nil {
		t.Fatalf("decode states fixture: %v", err)
	}
	return states
}

func loadServiceGeopoliticalZoneFixture(t *testing.T) []models.GeopoliticalZone {
	t.Helper()

	data, err := os.ReadFile("../../datasets/geography/geopolitical_zones.json")
	if err != nil {
		t.Fatalf("read geopolitical zones fixture: %v", err)
	}

	dec := json.NewDecoder(strings.NewReader(string(data)))
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

func loadServiceLocalGovernmentUnitFixture(t *testing.T) []models.LocalGovernmentUnit {
	t.Helper()

	data, err := os.ReadFile("../../datasets/geography/lgas.json")
	if err != nil {
		t.Fatalf("read local government units fixture: %v", err)
	}

	dec := json.NewDecoder(strings.NewReader(string(data)))
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

func loadServiceCountryFixture(t *testing.T) []models.CountryOrArea {
	t.Helper()

	data, err := os.ReadFile("../../datasets/geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("read countries and areas fixture: %v", err)
	}

	dec := json.NewDecoder(strings.NewReader(string(data)))
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
