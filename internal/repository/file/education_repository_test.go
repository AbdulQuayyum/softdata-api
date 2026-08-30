package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type educationJSONRepoStub struct {
	decodeFn func(context.Context, string, any) error

	calls     int
	pathCalls map[string]int
	lastPath  string
}

func (s *educationJSONRepoStub) Decode(ctx context.Context, relativePath string, destination any) error {
	s.calls++
	s.lastPath = relativePath
	if s.pathCalls != nil {
		s.pathCalls[relativePath]++
	}
	if s.decodeFn != nil {
		return s.decodeFn(ctx, relativePath, destination)
	}

	switch dest := destination.(type) {
	case *[]models.University:
		*dest = nil
		return nil
	case *[]models.CollegeOfEducation:
		*dest = nil
		return nil
	default:
		return fmt.Errorf("unexpected destination %T", destination)
	}
}

func TestNewEducationRepositoryRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	if _, err := NewEducationRepository(nil, "education/universities.json", "education/colleges_of_education.json"); err == nil {
		t.Fatal("expected nil json repository to be rejected")
	}
	if _, err := NewEducationRepository(&educationJSONRepoStub{}, "", "education/colleges_of_education.json"); err == nil {
		t.Fatal("expected empty universities path to be rejected")
	}
	if _, err := NewEducationRepository(&educationJSONRepoStub{}, "   ", "education/colleges_of_education.json"); err == nil {
		t.Fatal("expected whitespace universities path to be rejected")
	}
	if _, err := NewEducationRepository(&educationJSONRepoStub{}, "/tmp/education/universities.json", "education/colleges_of_education.json"); err == nil {
		t.Fatal("expected absolute universities path to be rejected")
	}
	if _, err := NewEducationRepository(&educationJSONRepoStub{}, "education/universities.json", ""); err == nil {
		t.Fatal("expected empty colleges path to be rejected")
	}
	if _, err := NewEducationRepository(&educationJSONRepoStub{}, "education/universities.json", "   "); err == nil {
		t.Fatal("expected whitespace colleges path to be rejected")
	}
	if _, err := NewEducationRepository(&educationJSONRepoStub{}, "education/universities.json", "/tmp/education/colleges_of_education.json"); err == nil {
		t.Fatal("expected absolute colleges path to be rejected")
	}

	stub := &educationJSONRepoStub{decodeFn: func(context.Context, string, any) error {
		t.Fatal("constructor should not decode dataset files")
		return nil
	}}
	repo, err := NewEducationRepository(stub, "  education/universities.json  ", "  education/colleges_of_education.json  ")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}
	if repo.universitiesPath != "education/universities.json" {
		t.Fatalf("unexpected stored path: %q", repo.universitiesPath)
	}
	if repo.collegesOfEducationPath != "education/colleges_of_education.json" {
		t.Fatalf("unexpected stored college path: %q", repo.collegesOfEducationPath)
	}
	if stub.calls != 0 {
		t.Fatalf("constructor decoded dataset file %d times", stub.calls)
	}
}

func TestEducationRepositoryListFilterAndLookup(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedUniversityFixture(t)
	repo := mustNewEducationRepositoryFromRecords(t, fixture)

	all, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{})
	if err != nil {
		t.Fatalf("ListUniversities() error = %v", err)
	}
	if all == nil {
		t.Fatal("ListUniversities() returned nil slice")
	}
	if got := len(all); got != 328 {
		t.Fatalf("unexpected university count: got %d want 328", got)
	}
	all[0].Name = "Changed"
	again, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{})
	if err != nil {
		t.Fatalf("ListUniversities() second call error = %v", err)
	}
	if again[0].Name != fixture[0].Name {
		t.Fatal("ListUniversities() shared mutable slice state")
	}

	wantOwnershipCounts := map[string]int{
		"federal": 77,
		"state":   69,
		"private": 182,
	}
	for ownershipType, want := range wantOwnershipCounts {
		filtered, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{OwnershipType: ownershipType})
		if err != nil {
			t.Fatalf("ListUniversities(%s) error = %v", ownershipType, err)
		}
		if filtered == nil {
			t.Fatalf("ListUniversities(%s) returned nil slice", ownershipType)
		}
		if got := len(filtered); got != want {
			t.Fatalf("unexpected count for %s: got %d want %d", ownershipType, got, want)
		}
		for _, university := range filtered {
			if university.OwnershipType != ownershipType {
				t.Fatalf("filtered slice contains wrong ownership type: %#v", university)
			}
		}
	}

	for _, stateID := range []string{"taraba", "fct"} {
		filtered, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{StateID: stateID})
		if err != nil {
			t.Fatalf("ListUniversities(%s) error = %v", stateID, err)
		}
		if filtered == nil {
			t.Fatalf("ListUniversities(%s) returned nil slice", stateID)
		}
		for _, university := range filtered {
			if university.StateID != stateID {
				t.Fatalf("filtered slice contains wrong state id: %#v", university)
			}
		}
	}

	taraba, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{StateID: "taraba"})
	if err != nil {
		t.Fatalf("ListUniversities(taraba) error = %v", err)
	}
	if got := len(taraba); got != 5 {
		t.Fatalf("unexpected Taraba count: got %d want 5", got)
	}
	fct, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{StateID: "fct"})
	if err != nil {
		t.Fatalf("ListUniversities(fct) error = %v", err)
	}
	if got := len(fct); got != 21 {
		t.Fatalf("unexpected FCT count: got %d want 21", got)
	}

	combined, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{OwnershipType: "private", StateID: "lagos"})
	if err != nil {
		t.Fatalf("ListUniversities(private+lagos) error = %v", err)
	}
	if combined == nil {
		t.Fatal("ListUniversities(private+lagos) returned nil slice")
	}
	if len(combined) == 0 {
		t.Fatal("expected combined filter to return records")
	}
	for _, university := range combined {
		if university.OwnershipType != "private" || university.StateID != "lagos" {
			t.Fatalf("combined filter returned unexpected record: %#v", university)
		}
	}

	empty, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{OwnershipType: "state", StateID: "fct"})
	if err != nil {
		t.Fatalf("ListUniversities(state+fct) error = %v", err)
	}
	if empty == nil {
		t.Fatal("ListUniversities(state+fct) returned nil slice")
	}
	if len(empty) != 0 {
		t.Fatalf("expected zero matches, got %d", len(empty))
	}

	for _, id := range []string{
		"federal-university-of-science-and-technology-epe",
		"gombe-state-university-gombe",
		"baze-university",
	} {
		university, err := repo.GetUniversityByID(context.Background(), id)
		if err != nil {
			t.Fatalf("GetUniversityByID(%s) error = %v", id, err)
		}
		if university.ID != id {
			t.Fatalf("unexpected university id for %s: %#v", id, university)
		}
	}

	if _, err := repo.GetUniversityByID(context.Background(), "Abubakar Tafawa Balewa University, Bauchi"); !errors.Is(err, interfaces.ErrUniversityNotFound) {
		t.Fatalf("name lookup error = %v, want ErrUniversityNotFound", err)
	}
	if _, err := repo.GetUniversityByID(context.Background(), "123e4567-e89b-12d3-a456-426614174000"); !errors.Is(err, interfaces.ErrUniversityNotFound) {
		t.Fatalf("uuid lookup error = %v, want ErrUniversityNotFound", err)
	}
	if _, err := repo.GetUniversityByID(context.Background(), "missing"); !errors.Is(err, interfaces.ErrUniversityNotFound) {
		t.Fatalf("missing lookup error = %v, want ErrUniversityNotFound", err)
	}
}

func TestEducationRepositoryRejectsInvalidFixtures(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedUniversityFixture(t)
	tests := []struct {
		name string
		mut  func([]models.University) []models.University
	}{
		{
			name: "nil slice",
			mut: func([]models.University) []models.University {
				return nil
			},
		},
		{
			name: "empty slice",
			mut: func([]models.University) []models.University {
				return make([]models.University, 0)
			},
		},
		{
			name: "327 records",
			mut: func(universities []models.University) []models.University {
				return append([]models.University(nil), universities[:327]...)
			},
		},
		{
			name: "329 records",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				return append(out, universities[0])
			},
		},
		{
			name: "incorrect ownership count",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0].OwnershipType = "private"
				return out
			},
		},
		{
			name: "incorrect state count",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0].StateID = "lagos"
				return out
			},
		},
		{
			name: "invalid ownership type",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0].OwnershipType = "unknown"
				return out
			},
		},
		{
			name: "invalid state id",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0].StateID = "invalid"
				return out
			},
		},
		{
			name: "invalid country code",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0].CountryCode = "GH"
				return out
			},
		},
		{
			name: "missing fields",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0].Name = ""
				return out
			},
		},
		{
			name: "duplicate id",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[1].ID = out[0].ID
				return out
			},
		},
		{
			name: "duplicate name",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[1].Name = out[0].Name
				out[1].ID = slugifyUniversityName(out[1].Name)
				return out
			},
		},
		{
			name: "invalid id syntax",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0].ID = "bad"
				return out
			},
		},
		{
			name: "id name mismatch",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0].ID = "wrong-" + slugifyUniversityName(out[0].Name)
				return out
			},
		},
		{
			name: "incorrect ordering",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0], out[1] = out[1], out[0]
				return out
			},
		},
		{
			name: "raw typo form rejection",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0].Name = "Gombe State Univeristy, Gombe"
				return out
			},
		},
		{
			name: "former-name record rejection",
			mut: func(universities []models.University) []models.University {
				out := append([]models.University(nil), universities...)
				out[0].Name = "Rev. Fr. Moses Orshio Adasu (Formerly, Benue State University), Makurdi"
				return out
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mustNewEducationRepositoryFromRecords(t, tc.mut(fixture))
			_, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{})
			if err == nil {
				t.Fatalf("expected %s to fail", tc.name)
			}
			if !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("unexpected error for %s: %v", tc.name, err)
			}
		})
	}
}

func TestEducationRepositoryContextAndSanitizedErrors(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedUniversityFixture(t)
	repo := mustNewEducationRepositoryFromRecords(t, fixture)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.ListUniversities(ctx, interfaces.UniversityFilter{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context error = %v, want context.Canceled", err)
	}

	deadlineCtx, cancelDeadline := context.WithTimeout(context.Background(), time.Nanosecond)
	time.Sleep(2 * time.Nanosecond)
	cancelDeadline()
	if _, err := repo.GetUniversityByID(deadlineCtx, "baze-university"); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline context error = %v, want context cancellation/deadline", err)
	}

	secretErr := errors.New("/private/tmp/education/universities.json: permission denied")
	stub := &educationJSONRepoStub{decodeFn: func(context.Context, string, any) error { return secretErr }}
	sanitizedRepo, err := NewEducationRepository(stub, "education/universities.json", "education/colleges_of_education.json")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}
	_, err = sanitizedRepo.ListUniversities(context.Background(), interfaces.UniversityFilter{})
	if err == nil {
		t.Fatal("expected sanitized decode error")
	}
	if strings.Contains(err.Error(), "/private/tmp/education/universities.json") {
		t.Fatalf("error leaked filesystem path: %v", err)
	}
	if !errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		t.Fatalf("unexpected sanitized error: %v", err)
	}
}

func TestEducationRepositoryDoesNotMutateReturnedSlice(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedUniversityFixture(t)
	repo := mustNewEducationRepositoryFromRecords(t, fixture)

	loaded, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{OwnershipType: "private"})
	if err != nil {
		t.Fatalf("ListUniversities() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("ListUniversities() returned nil slice")
	}
	if len(loaded) == 0 {
		t.Fatal("expected non-empty slice")
	}
	originalFirst := loaded[0]

	loaded[0].Name = "Changed"
	again, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{OwnershipType: "private"})
	if err != nil {
		t.Fatalf("ListUniversities() second call error = %v", err)
	}
	if again[0] != originalFirst {
		t.Fatal("ListUniversities() shared mutable slice state")
	}
}

func TestEducationRepositoryDecodeCounts(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedUniversityFixture(t)
	stub := &educationJSONRepoStub{decodeFn: func(ctx context.Context, relativePath string, destination any) error {
		switch dest := destination.(type) {
		case *[]models.University:
			*dest = cloneUniversityList(fixture)
			return nil
		default:
			return fmt.Errorf("unexpected destination %T", destination)
		}
	}, pathCalls: map[string]int{}}
	repo, err := NewEducationRepository(stub, "education/universities.json", "education/colleges_of_education.json")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}

	for _, call := range []func() error{
		func() error {
			_, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{})
			return err
		},
		func() error {
			_, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{OwnershipType: "federal"})
			return err
		},
		func() error {
			_, err := repo.GetUniversityByID(context.Background(), "baze-university")
			return err
		},
	} {
		stub.calls = 0
		stub.pathCalls = map[string]int{}
		if err := call(); err != nil {
			t.Fatalf("repository call error = %v", err)
		}
		if stub.calls != 1 {
			t.Fatalf("unexpected decode call count: %d", stub.calls)
		}
		if stub.pathCalls["education/universities.json"] != 1 {
			t.Fatalf("unexpected decode path counts: %#v", stub.pathCalls)
		}
	}
}

func loadApprovedUniversityFixture(t *testing.T) []models.University {
	t.Helper()

	var universities []models.University
	dec := json.NewDecoder(bytes.NewReader(readEducationDatasetBytes(t, educationDatasetPath("education", "universities.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&universities); err != nil {
		t.Fatalf("decode approved universities dataset: %v", err)
	}
	return universities
}

func mustNewEducationRepositoryFromRecords(t *testing.T, universities []models.University) *EducationFileRepository {
	t.Helper()

	root := t.TempDir()
	fixturePath := filepath.Join(root, "education", "universities.json")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}
	encoded, err := json.MarshalIndent(universities, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, append(encoded, '\n'), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	jsonRepo, err := NewJSONRepository(root, 16<<20)
	if err != nil {
		t.Fatalf("NewJSONRepository() error = %v", err)
	}
	repo, err := NewEducationRepository(jsonRepo, "education/universities.json", "education/colleges_of_education.json")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}
	return repo
}

func readEducationDatasetBytes(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func educationDatasetPath(parts ...string) string {
	elems := append([]string{"..", "..", "..", "datasets"}, parts...)
	return filepath.Join(elems...)
}
