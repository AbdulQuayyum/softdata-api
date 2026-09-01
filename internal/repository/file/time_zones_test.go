package file

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestGeographyRepositoryTimeZonesConstructorAndReadIsolation(t *testing.T) {
	t.Parallel()

	if _, err := NewGeographyRepository(nil, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json"); err == nil {
		t.Fatal("expected nil json repository to be rejected")
	}
	if _, err := NewGeographyRepository(&geographyJSONRepoStub{}, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "   ", "geography/countries_and_areas.json"); err == nil {
		t.Fatal("expected whitespace time zone path to be rejected")
	}

	stub := &geographyJSONRepoStub{decodeFn: func(context.Context, string, any) error {
		t.Fatal("constructor should not decode data")
		return nil
	}}
	repo, err := NewGeographyRepository(stub, " geography/states.json ", " geography/geopolitical_zones.json ", " geography/lgas.json ", " geography/time_zones.json ", " geography/countries_and_areas.json ")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	if repo.timeZonesPath != "geography/time_zones.json" {
		t.Fatalf("unexpected time zone path: %q", repo.timeZonesPath)
	}
}

func TestGeographyRepositoryTimeZonesListGetAndFiltering(t *testing.T) {
	t.Parallel()

	states := loadApprovedStateFixture(t)
	zones := loadApprovedZoneFixture(t)
	timeZones := loadApprovedTimeZoneFixture(t)
	countries := loadApprovedCountryFixture(t)
	jsonRepo := &geographyJSONRepoStub{
		states:    states,
		zones:     zones,
		timeZones: timeZones,
		countries: countries,
		pathCalls: map[string]int{},
	}
	repo, err := NewGeographyRepository(jsonRepo, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}

	loaded, err := repo.ListTimeZones(context.Background(), interfaces.TimeZoneFilter{})
	if err != nil {
		t.Fatalf("ListTimeZones() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("ListTimeZones() returned nil slice")
	}
	if len(loaded) != 312 {
		t.Fatalf("unexpected time zone count: %d", len(loaded))
	}
	if loaded[0].ID != "Africa/Abidjan" || loaded[len(loaded)-1].ID != "Pacific/Tongatapu" {
		t.Fatalf("unexpected ordering: first=%q last=%q", loaded[0].ID, loaded[len(loaded)-1].ID)
	}
	if jsonRepo.pathCalls["geography/time_zones.json"] != 1 || jsonRepo.pathCalls["geography/countries_and_areas.json"] != 1 {
		t.Fatalf("unexpected decode counts: %#v", jsonRepo.pathCalls)
	}
	if jsonRepo.pathCalls["geography/states.json"] != 0 || jsonRepo.pathCalls["geography/geopolitical_zones.json"] != 0 || jsonRepo.pathCalls["geography/lgas.json"] != 0 {
		t.Fatalf("time zone list decoded unrelated geography files: %#v", jsonRepo.pathCalls)
	}

	loaded[0].CountryAreaIDs[0] = "xx"
	again, err := repo.ListTimeZones(context.Background(), interfaces.TimeZoneFilter{})
	if err != nil {
		t.Fatalf("ListTimeZones() second call error = %v", err)
	}
	if again[0].CountryAreaIDs[0] == "xx" {
		t.Fatal("ListTimeZones() shared mutable slice state")
	}

	filtered, err := repo.ListTimeZones(context.Background(), interfaces.TimeZoneFilter{CountryAreaID: "ng"})
	if err != nil {
		t.Fatalf("ListTimeZones(filter) error = %v", err)
	}
	if len(filtered) == 0 {
		t.Fatal("expected Nigeria to map to at least one time zone")
	}
	for _, timeZone := range filtered {
		if !containsCountryAreaID(timeZone.CountryAreaIDs, "ng") {
			t.Fatalf("filtered time zone does not contain ng: %#v", timeZone)
		}
	}

	empty, err := repo.ListTimeZones(context.Background(), interfaces.TimeZoneFilter{CountryAreaID: "bv"})
	if err != nil {
		t.Fatalf("ListTimeZones(bv) error = %v", err)
	}
	if empty == nil {
		t.Fatal("ListTimeZones(bv) returned nil slice")
	}
	if len(empty) != 0 {
		t.Fatalf("unexpected bv result count: %d", len(empty))
	}

	zone, err := repo.GetTimeZone(context.Background(), "Africa/Lagos")
	if err != nil {
		t.Fatalf("GetTimeZone(Africa/Lagos) error = %v", err)
	}
	if zone.ID != "Africa/Lagos" {
		t.Fatalf("unexpected zone response: %#v", zone)
	}
	if _, err := repo.GetTimeZone(context.Background(), "Etc/UTC"); !errors.Is(err, interfaces.ErrTimeZoneNotFound) {
		t.Fatalf("alias lookup error = %v, want ErrTimeZoneNotFound", err)
	}
}

func TestGeographyRepositoryTimeZonesContextAndValidationErrors(t *testing.T) {
	t.Parallel()

	states := loadApprovedStateFixture(t)
	zones := loadApprovedZoneFixture(t)
	timeZones := loadApprovedTimeZoneFixture(t)
	countries := loadApprovedCountryFixture(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo, err := NewGeographyRepository(&geographyJSONRepoStub{states: states, zones: zones, timeZones: timeZones, countries: countries}, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	if _, err := repo.ListTimeZones(ctx, interfaces.TimeZoneFilter{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("ListTimeZones() error = %v, want context.Canceled", err)
	}
	if _, err := repo.GetTimeZone(ctx, "Africa/Lagos"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetTimeZone() error = %v, want context.Canceled", err)
	}

	expiredCtx, cancelExpired := context.WithDeadline(context.Background(), time.Now().Add(-time.Minute))
	defer cancelExpired()
	if _, err := repo.ListTimeZones(expiredCtx, interfaces.TimeZoneFilter{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ListTimeZones() error = %v, want context.DeadlineExceeded", err)
	}
	if _, err := repo.GetTimeZone(expiredCtx, "Africa/Lagos"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetTimeZone() error = %v, want context.DeadlineExceeded", err)
	}

	badRepo := &geographyJSONRepoStub{decodeFn: func(context.Context, string, any) error { return interfaces.ErrDatasetFileUnavailable }}
	repo, err = NewGeographyRepository(badRepo, "geography/states.json", "geography/geopolitical_zones.json", "geography/lgas.json", "geography/time_zones.json", "geography/countries_and_areas.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	if _, err := repo.ListTimeZones(context.Background(), interfaces.TimeZoneFilter{}); !errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		t.Fatalf("ListTimeZones() error = %v, want ErrDatasetFileUnavailable", err)
	}
	if _, err := repo.GetTimeZone(context.Background(), "Africa/Lagos"); !errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		t.Fatalf("GetTimeZone() error = %v, want ErrDatasetFileUnavailable", err)
	}
}

func loadApprovedTimeZoneFixture(t *testing.T) []models.TimeZone {
	t.Helper()

	path := filepath.Clean("../../../datasets/geography/time_zones.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read time zones fixture: %v", err)
	}

	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()

	var timeZones []models.TimeZone
	if err := dec.Decode(&timeZones); err != nil {
		t.Fatalf("decode time zones fixture: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("time zone fixture contains trailing json")
	}
	return timeZones
}
