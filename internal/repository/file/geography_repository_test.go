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
	states   []models.State
	decodeFn func(context.Context, string, any) error

	calls    int
	lastPath string
}

func (s *geographyJSONRepoStub) Decode(ctx context.Context, relativePath string, destination any) error {
	s.calls++
	s.lastPath = relativePath
	if s.decodeFn != nil {
		return s.decodeFn(ctx, relativePath, destination)
	}

	dest, ok := destination.(*[]models.State)
	if !ok {
		return fmt.Errorf("unexpected destination %T", destination)
	}
	*dest = cloneTestStates(s.states)
	return nil
}

func TestNewGeographyRepositoryRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	if _, err := NewGeographyRepository(nil, "geography/states.json"); err == nil {
		t.Fatal("expected nil json repository to be rejected")
	}
	if _, err := NewGeographyRepository(&geographyJSONRepoStub{}, ""); err == nil {
		t.Fatal("expected empty path to be rejected")
	}
	if _, err := NewGeographyRepository(&geographyJSONRepoStub{}, "   "); err == nil {
		t.Fatal("expected whitespace path to be rejected")
	}

	repo, err := NewGeographyRepository(&geographyJSONRepoStub{}, "  geography/states.json  ")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	if repo.statesPath != "geography/states.json" {
		t.Fatalf("unexpected stored path: %q", repo.statesPath)
	}
}

func TestGeographyRepositoryListStatesAndGetStateByID(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedStateFixture(t)
	jsonRepo := &geographyJSONRepoStub{states: fixture}
	repo, err := NewGeographyRepository(jsonRepo, "geography/states.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}

	states, err := repo.ListStates(context.Background())
	if err != nil {
		t.Fatalf("ListStates() error = %v", err)
	}
	if jsonRepo.calls != 1 {
		t.Fatalf("unexpected decode call count: %d", jsonRepo.calls)
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

func TestGeographyRepositoryRejectsNilAndEmptyDecodedSlices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		repo *GeographyFileRepository
	}{
		{
			name: "nil decoded slice",
			repo: newGeographyRepoWithDecodeFn(t, func(_ context.Context, _ string, destination any) error {
				dest := destination.(*[]models.State)
				*dest = nil
				return nil
			}),
		},
		{
			name: "empty decoded slice",
			repo: newGeographyRepoWithDecodeFn(t, func(_ context.Context, _ string, destination any) error {
				dest := destination.(*[]models.State)
				*dest = make([]models.State, 0)
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

	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		stub := &geographyJSONRepoStub{
			states: fixture,
			decodeFn: func(context.Context, string, any) error {
				t.Fatal("decode should not be called for canceled context")
				return nil
			},
		}
		repo, err := NewGeographyRepository(stub, "geography/states.json")
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
			decodeFn: func(context.Context, string, any) error {
				t.Fatal("decode should not be called for expired deadline")
				return nil
			},
		}
		repo, err := NewGeographyRepository(stub, "geography/states.json")
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
}

func newGeographyRepoForError(t *testing.T, decodeErr error) *GeographyFileRepository {
	t.Helper()

	repo, err := NewGeographyRepository(&geographyJSONRepoStub{decodeFn: func(context.Context, string, any) error { return decodeErr }}, "geography/states.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	return repo
}

func newGeographyRepoWithStates(t *testing.T, states []models.State) *GeographyFileRepository {
	t.Helper()

	repo, err := NewGeographyRepository(&geographyJSONRepoStub{states: states}, "geography/states.json")
	if err != nil {
		t.Fatalf("NewGeographyRepository() error = %v", err)
	}
	return repo
}

func newGeographyRepoWithDecodeFn(t *testing.T, decodeFn func(context.Context, string, any) error) *GeographyFileRepository {
	t.Helper()

	repo, err := NewGeographyRepository(&geographyJSONRepoStub{decodeFn: decodeFn}, "geography/states.json")
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

func cloneTestStates(states []models.State) []models.State {
	if len(states) == 0 {
		return make([]models.State, 0)
	}
	cloned := make([]models.State, len(states))
	copy(cloned, states)
	return cloned
}
