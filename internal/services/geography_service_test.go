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
	states    []models.State
	listErr   error
	getErr    error
	listCalls int
	getCalls  int
	lastID    string
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
