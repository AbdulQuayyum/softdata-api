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

func TestGeographyServiceTimeZonesListGetAndFiltering(t *testing.T) {
	t.Parallel()

	timeZones := loadServiceTimeZoneFixture(t)
	stub := &geographyRepoStub{timeZones: timeZones}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	loaded, err := svc.ListTimeZones(context.Background(), TimeZoneListInput{})
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
	if stub.timeZoneListCalls != 1 {
		t.Fatalf("unexpected list call count: %d", stub.timeZoneListCalls)
	}

	loaded[0].CountryAreaIDs[0] = "xx"
	again, err := svc.ListTimeZones(context.Background(), TimeZoneListInput{})
	if err != nil {
		t.Fatalf("ListTimeZones() second call error = %v", err)
	}
	if again[0].CountryAreaIDs[0] == "xx" {
		t.Fatal("ListTimeZones() exposed shared mutable slice state")
	}

	byCountry, err := svc.ListTimeZones(context.Background(), TimeZoneListInput{CountryAreaID: "ng"})
	if err != nil {
		t.Fatalf("ListTimeZones(filter) error = %v", err)
	}
	if len(byCountry) == 0 {
		t.Fatal("expected Nigeria to map to at least one time zone")
	}
	for _, timeZone := range byCountry {
		if !containsString(timeZone.CountryAreaIDs, "ng") {
			t.Fatalf("filtered time zone does not contain ng: %#v", timeZone)
		}
	}

	empty, err := svc.ListTimeZones(context.Background(), TimeZoneListInput{CountryAreaID: "bv"})
	if err != nil {
		t.Fatalf("ListTimeZones(bv) error = %v", err)
	}
	if empty == nil {
		t.Fatal("ListTimeZones(bv) returned nil slice")
	}
	if len(empty) != 0 {
		t.Fatalf("unexpected bv result count: %d", len(empty))
	}

	zone, err := svc.GetTimeZone(context.Background(), "  Africa/Lagos  ")
	if err != nil {
		t.Fatalf("GetTimeZone() error = %v", err)
	}
	if stub.lastTimeZoneID != "Africa/Lagos" {
		t.Fatalf("unexpected lookup id: %q", stub.lastTimeZoneID)
	}
	if zone.ID != "Africa/Lagos" {
		t.Fatalf("unexpected time zone response: %#v", zone)
	}
	if stub.timeZoneGetCalls != 1 {
		t.Fatalf("unexpected get call count: %d", stub.timeZoneGetCalls)
	}
}

func TestGeographyServiceTimeZonesValidationAndErrors(t *testing.T) {
	t.Parallel()

	timeZones := loadServiceTimeZoneFixture(t)
	stub := &geographyRepoStub{timeZones: timeZones}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	for _, tc := range []struct {
		name  string
		input TimeZoneListInput
		want  error
	}{
		{name: "empty filter", input: TimeZoneListInput{}, want: nil},
		{name: "uppercase filter", input: TimeZoneListInput{CountryAreaID: "NG"}, want: ErrInvalidTimeZoneCountryAreaID},
		{name: "spaces", input: TimeZoneListInput{CountryAreaID: "n g"}, want: ErrInvalidTimeZoneCountryAreaID},
		{name: "underscore", input: TimeZoneListInput{CountryAreaID: "n_g"}, want: ErrInvalidTimeZoneCountryAreaID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.timeZoneListCalls = 0
			loaded, err := svc.ListTimeZones(context.Background(), tc.input)
			if tc.want != nil {
				if !errors.Is(err, tc.want) {
					t.Fatalf("ListTimeZones() error = %v, want %v", err, tc.want)
				}
				if stub.timeZoneListCalls != 0 {
					t.Fatalf("repository should not have been called, got %d", stub.timeZoneListCalls)
				}
				return
			}
			if err != nil {
				t.Fatalf("ListTimeZones() error = %v", err)
			}
			if loaded == nil {
				t.Fatal("expected non-nil slice")
			}
		})
	}

	for _, tc := range []struct {
		name string
		id   string
		want error
	}{
		{name: "empty", id: "", want: ErrInvalidTimeZoneID},
		{name: "space", id: "  ", want: ErrInvalidTimeZoneID},
		{name: "slash only", id: "/", want: ErrInvalidTimeZoneID},
		{name: "leading slash", id: "/Africa/Lagos", want: ErrInvalidTimeZoneID},
		{name: "trailing slash", id: "Africa/Lagos/", want: ErrInvalidTimeZoneID},
		{name: "double slash", id: "Africa//Lagos", want: ErrInvalidTimeZoneID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.timeZoneGetCalls = 0
			if _, err := svc.GetTimeZone(context.Background(), tc.id); !errors.Is(err, tc.want) {
				t.Fatalf("GetTimeZone() error = %v, want %v", err, tc.want)
			}
			if stub.timeZoneGetCalls != 0 {
				t.Fatalf("repository should not have been called, got %d", stub.timeZoneGetCalls)
			}
		})
	}

	stub.timeZoneGetErr = interfaces.ErrTimeZoneNotFound
	if _, err := svc.GetTimeZone(context.Background(), "Africa/Nowhere"); !errors.Is(err, ErrTimeZoneNotFound) {
		t.Fatalf("missing lookup error = %v, want ErrTimeZoneNotFound", err)
	}

	stub.timeZoneGetErr = errors.New("explode")
	if _, err := svc.GetTimeZone(context.Background(), "Africa/Lagos"); err == nil || strings.Contains(err.Error(), "explode") {
		t.Fatalf("unexpected get error was not sanitized: %v", err)
	}

	stub.timeZoneListErr = errors.New("boom")
	if _, err := svc.ListTimeZones(context.Background(), TimeZoneListInput{}); err == nil || strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected list error was not sanitized: %v", err)
	}
}

func TestGeographyServiceTimeZonesContextCancellationAndDeadline(t *testing.T) {
	t.Parallel()

	stub := &geographyRepoStub{timeZones: loadServiceTimeZoneFixture(t)}
	svc, err := NewGeographyService(stub)
	if err != nil {
		t.Fatalf("NewGeographyService() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.ListTimeZones(ctx, TimeZoneListInput{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("ListTimeZones() error = %v, want context.Canceled", err)
	}
	if _, err := svc.GetTimeZone(ctx, "Africa/Lagos"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetTimeZone() error = %v, want context.Canceled", err)
	}

	deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Minute))
	defer deadlineCancel()
	if _, err := svc.ListTimeZones(deadlineCtx, TimeZoneListInput{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ListTimeZones() error = %v, want context.DeadlineExceeded", err)
	}
	if _, err := svc.GetTimeZone(deadlineCtx, "Africa/Lagos"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetTimeZone() error = %v, want context.DeadlineExceeded", err)
	}
}

func loadServiceTimeZoneFixture(t *testing.T) []models.TimeZone {
	t.Helper()

	data, err := os.ReadFile("../../datasets/geography/time_zones.json")
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
