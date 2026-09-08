package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	healthFacilityTestPath = "healthcare/health_facilities.json"
	healthStatesTestPath   = "geography/states.json"
	healthLGAsTestPath     = "geography/lgas.json"
)

func TestHealthFacilityRepositoryRealDataset(t *testing.T) {
	repository := newRealHealthFacilityRepository(t)

	result, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{})
	if err != nil {
		t.Fatalf("ListHealthFacilities() error = %v", err)
	}
	if result.Page != 1 || result.PageSize != 50 || result.Total != 50654 || result.TotalPages != 1014 {
		t.Fatalf("unexpected default result metadata: %#v", result)
	}
	if len(result.Facilities) != 50 || result.Facilities == nil {
		t.Fatalf("unexpected default result size: %d", len(result.Facilities))
	}

	last, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{Page: 1014, PageSize: 50})
	if err != nil || len(last.Facilities) != 4 {
		t.Fatalf("unexpected final page: result=%#v err=%v", last, err)
	}
	beyond, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{Page: 1015, PageSize: 50})
	if err != nil || beyond.Facilities == nil || len(beyond.Facilities) != 0 || beyond.Total != 50654 || beyond.TotalPages != 1014 {
		t.Fatalf("unexpected beyond-final page: result=%#v err=%v", beyond, err)
	}

	all, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{Page: 1, PageSize: 100, FacilityType: "clinic"})
	if err != nil || all.Total != 13819 || len(all.Facilities) != 100 {
		t.Fatalf("unexpected clinic page: result=%#v err=%v", all, err)
	}
	first, middle, lastRecord := allFacilityRecords(t)
	for _, want := range []models.HealthFacility{first, middle, lastRecord} {
		got, err := repository.GetHealthFacility(context.Background(), want.ID)
		if err != nil || got.ID != want.ID {
			t.Fatalf("GetHealthFacility(%q) = %#v, %v", want.ID, got, err)
		}
	}
	if _, err := repository.GetHealthFacility(context.Background(), "valid-but-unknown"); !errors.Is(err, interfaces.ErrHealthFacilityNotFound) {
		t.Fatalf("unknown lookup error = %v", err)
	}

	state := first.StateID
	stateResult, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{PageSize: 100, StateID: state})
	if err != nil || stateResult.Total == 0 {
		t.Fatalf("state filter failed: result=%#v err=%v", stateResult, err)
	}
	lgaResult, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{PageSize: 100, LGAID: first.LGAID})
	if err != nil || lgaResult.Total == 0 {
		t.Fatalf("LGA filter failed: result=%#v err=%v", lgaResult, err)
	}
	combined, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{PageSize: 100, StateID: first.StateID, LGAID: first.LGAID, Search: first.Name})
	if err != nil || combined.Total == 0 {
		t.Fatalf("combined filter failed: result=%#v err=%v", combined, err)
	}
	level, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{PageSize: 100, FacilityLevel: "primary"})
	if err != nil || level.Total != 44582 {
		t.Fatalf("facility level filter failed: result=%#v err=%v", level, err)
	}
	ownership, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{PageSize: 100, OwnershipType: "private"})
	if err != nil || ownership.Total != 11652 {
		t.Fatalf("ownership filter failed: result=%#v err=%v", ownership, err)
	}
	if _, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{StateID: "abia", LGAID: "lagos-ikeja"}); !errors.Is(err, interfaces.ErrInvalidHealthFacilityStateLGA) {
		t.Fatalf("invalid state/LGA error = %v", err)
	}
	search, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{PageSize: 100, Search: "  " + first.Name[:3] + "  "})
	if err != nil || search.Total == 0 {
		t.Fatalf("search failed: result=%#v err=%v", search, err)
	}

	result.Facilities[0].Name = "mutated"
	result.Facilities[0].Latitude = nil
	again, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{})
	if err != nil || again.Facilities[0].Name == "mutated" || again.Facilities[0].Latitude == nil {
		t.Fatalf("result mutation affected cache: result=%#v err=%v", again, err)
	}
}

func TestHealthFacilityRepositoryValidationAndContext(t *testing.T) {
	repository := newRealHealthFacilityRepository(t)
	tests := []struct {
		name  string
		query interfaces.HealthFacilityQuery
		want  error
	}{
		{"bad page", interfaces.HealthFacilityQuery{Page: -1}, interfaces.ErrInvalidHealthFacilityQuery},
		{"bad size", interfaces.HealthFacilityQuery{PageSize: 101}, interfaces.ErrInvalidHealthFacilityQuery},
		{"bad state", interfaces.HealthFacilityQuery{StateID: "not-a-state"}, interfaces.ErrInvalidHealthFacilityStateFilter},
		{"bad lga", interfaces.HealthFacilityQuery{LGAID: "not-an-lga"}, interfaces.ErrInvalidHealthFacilityLGAFilter},
		{"bad type", interfaces.HealthFacilityQuery{FacilityType: "hospital"}, interfaces.ErrInvalidHealthFacilityType},
		{"bad level", interfaces.HealthFacilityQuery{FacilityLevel: "quaternary"}, interfaces.ErrInvalidHealthFacilityLevel},
		{"bad ownership", interfaces.HealthFacilityQuery{OwnershipType: "unknown"}, interfaces.ErrInvalidHealthFacilityOwnership},
		{"bad search", interfaces.HealthFacilityQuery{Search: string(make([]byte, 101))}, interfaces.ErrInvalidHealthFacilitySearch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := repository.ListHealthFacilities(context.Background(), test.query); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repository.ListHealthFacilities(ctx, interfaces.HealthFacilityQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list error = %v", err)
	}
}

func TestHealthFacilityRepositoryConcurrentFirstLoad(t *testing.T) {
	repository := newRealHealthFacilityRepository(t)
	const workers = 8
	errs := make(chan error, workers)
	var wait sync.WaitGroup
	for i := 0; i < workers; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{PageSize: 100})
			if err == nil && len(result.Facilities) != 100 {
				err = fmt.Errorf("got %d facilities", len(result.Facilities))
			}
			errs <- err
		}()
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestHealthFacilityRepositoryFailedLoadCanRetry(t *testing.T) {
	fixture := loadHealthFacilitiesFixture(t)
	states := loadHealthStatesFixture(t)
	lgas := loadHealthLGAsFixture(t)
	stub := &healthFacilityJSONStub{records: fixture, states: states, lgas: lgas, failHealthOnce: true}
	repository := mustNewHealthFacilityRepository(t, stub)
	if _, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("first load error = %v", err)
	}
	if _, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{}); err != nil {
		t.Fatalf("retry error = %v", err)
	}
	if stub.calls[healthFacilityTestPath] != 2 {
		t.Fatalf("health decode calls = %d, want 2", stub.calls[healthFacilityTestPath])
	}
}

func TestHealthFacilityRepositorySanitizesUnexpectedErrors(t *testing.T) {
	secret := errors.New("/machine/private/health_facilities.json: decoder secret")
	stub := &healthFacilityJSONStub{healthErr: secret}
	repository := mustNewHealthFacilityRepository(t, stub)
	_, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{})
	if err == nil || errors.Is(err, secret) || stringsContainsAny(err.Error(), []string{"/machine", "decoder secret"}) {
		t.Fatalf("unsanitized error: %v", err)
	}
}

func TestHealthFacilityRepositoryRejectsMalformedTrailingAndUnknownJSON(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "malformed", body: "["},
		{name: "trailing", body: "[] []"},
		{name: "unknown field", body: `[{"unexpected":true}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, healthFacilityTestPath)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(test.body), 0o644); err != nil {
				t.Fatal(err)
			}
			jsonRepository, err := NewJSONRepository(root, 1<<20)
			if err != nil {
				t.Fatal(err)
			}
			repository, err := NewHealthFacilityRepository(jsonRepository, healthFacilityTestPath, healthStatesTestPath, healthLGAsTestPath)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v, want invalid dataset", err)
			}
		})
	}
}

func newRealHealthFacilityRepository(t testing.TB) *HealthFacilityFileRepository {
	t.Helper()
	jsonRepository, err := NewJSONRepository("../../../datasets", 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewHealthFacilityRepository(jsonRepository, healthFacilityTestPath, healthStatesTestPath, healthLGAsTestPath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

func mustNewHealthFacilityRepository(t testing.TB, stub *healthFacilityJSONStub) *HealthFacilityFileRepository {
	t.Helper()
	repository, err := NewHealthFacilityRepository(stub, healthFacilityTestPath, healthStatesTestPath, healthLGAsTestPath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

func allFacilityRecords(t testing.TB) (models.HealthFacility, models.HealthFacility, models.HealthFacility) {
	t.Helper()
	items := loadHealthFacilitiesFixture(t)
	return items[0], items[len(items)/2], items[len(items)-1]
}

func loadHealthFacilitiesFixture(t testing.TB) []models.HealthFacility {
	t.Helper()
	var items []models.HealthFacility
	decodeTestJSON(t, "../../../datasets/healthcare/health_facilities.json", &items)
	return items
}

func loadHealthStatesFixture(t testing.TB) []models.State {
	t.Helper()
	var items []models.State
	decodeTestJSON(t, "../../../datasets/geography/states.json", &items)
	return items
}

func loadHealthLGAsFixture(t testing.TB) []models.LocalGovernmentUnit {
	t.Helper()
	var items []models.LocalGovernmentUnit
	decodeTestJSON(t, "../../../datasets/geography/lgas.json", &items)
	return items
}

func decodeTestJSON(t testing.TB, path string, destination any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, destination); err != nil {
		t.Fatal(err)
	}
}

type healthFacilityJSONStub struct {
	records        []models.HealthFacility
	states         []models.State
	lgas           []models.LocalGovernmentUnit
	healthErr      error
	failHealthOnce bool
	calls          map[string]int
}

func (s *healthFacilityJSONStub) Decode(_ context.Context, path string, destination any) error {
	if s.calls == nil {
		s.calls = map[string]int{}
	}
	s.calls[path]++
	if path == healthFacilityTestPath {
		if s.failHealthOnce {
			s.failHealthOnce = false
			return interfaces.ErrInvalidDatasetFile
		}
		if s.healthErr != nil {
			return s.healthErr
		}
		*(destination.(*[]models.HealthFacility)) = append([]models.HealthFacility(nil), s.records...)
		return nil
	}
	switch path {
	case healthStatesTestPath:
		*(destination.(*[]models.State)) = append([]models.State(nil), s.states...)
	case healthLGAsTestPath:
		*(destination.(*[]models.LocalGovernmentUnit)) = append([]models.LocalGovernmentUnit(nil), s.lgas...)
	default:
		return fmt.Errorf("unexpected path %s", path)
	}
	return nil
}

func stringsContainsAny(value string, candidates []string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func TestHealthFacilityRepositoryFixturePathsAreSafe(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{healthFacilityTestPath, healthStatesTestPath, healthLGAsTestPath} {
		if err := os.MkdirAll(filepath.Join(root, filepath.Dir(path)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NewHealthFacilityRepository(&healthFacilityJSONStub{}, "/absolute.json", healthStatesTestPath, healthLGAsTestPath); err == nil {
		t.Fatal("expected absolute health path rejection")
	}
}
