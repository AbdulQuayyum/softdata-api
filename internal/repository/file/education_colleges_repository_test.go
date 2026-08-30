package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestEducationRepositoryListCollegesOfEducationAndGet(t *testing.T) {
	t.Parallel()

	universities := loadApprovedUniversityFixture(t)
	colleges := loadApprovedCollegeFixture(t)
	repo := mustNewEducationRepositoryFromEducationFixtures(t, universities, colleges)

	all, err := repo.ListCollegesOfEducation(context.Background(), interfaces.CollegeOfEducationFilter{})
	if err != nil {
		t.Fatalf("ListCollegesOfEducation() error = %v", err)
	}
	if all == nil {
		t.Fatal("ListCollegesOfEducation() returned nil slice")
	}
	if got := len(all); got != 244 {
		t.Fatalf("unexpected college count: got %d want 244", got)
	}
	if all[0].Name == "" || all[0].ID == "" {
		t.Fatalf("unexpected first college: %#v", all[0])
	}
	all[0].Name = "Changed"
	again, err := repo.ListCollegesOfEducation(context.Background(), interfaces.CollegeOfEducationFilter{})
	if err != nil {
		t.Fatalf("ListCollegesOfEducation() second call error = %v", err)
	}
	if again[0].Name == "Changed" {
		t.Fatal("ListCollegesOfEducation() shared mutable slice state")
	}

	wantOwnershipCounts := map[string]int{
		"federal": 28,
		"state":   48,
		"private": 168,
	}
	for ownershipType, want := range wantOwnershipCounts {
		filtered, err := repo.ListCollegesOfEducation(context.Background(), interfaces.CollegeOfEducationFilter{OwnershipType: ownershipType})
		if err != nil {
			t.Fatalf("ListCollegesOfEducation(%s) error = %v", ownershipType, err)
		}
		if filtered == nil {
			t.Fatalf("ListCollegesOfEducation(%s) returned nil slice", ownershipType)
		}
		if got := len(filtered); got != want {
			t.Fatalf("unexpected count for %s: got %d want %d", ownershipType, got, want)
		}
		for _, college := range filtered {
			if college.OwnershipType != ownershipType {
				t.Fatalf("filtered slice contains wrong ownership type: %#v", college)
			}
		}
	}

	fct, err := repo.ListCollegesOfEducation(context.Background(), interfaces.CollegeOfEducationFilter{StateID: "fct"})
	if err != nil {
		t.Fatalf("ListCollegesOfEducation(fct) error = %v", err)
	}
	if got := len(fct); got != 5 {
		t.Fatalf("unexpected FCT count: got %d want 5", got)
	}

	combined, err := repo.ListCollegesOfEducation(context.Background(), interfaces.CollegeOfEducationFilter{OwnershipType: "private", StateID: "lagos"})
	if err != nil {
		t.Fatalf("ListCollegesOfEducation(private+lagos) error = %v", err)
	}
	if combined == nil {
		t.Fatal("ListCollegesOfEducation(private+lagos) returned nil slice")
	}
	for _, college := range combined {
		if college.OwnershipType != "private" || college.StateID != "lagos" {
			t.Fatalf("combined filter returned unexpected record: %#v", college)
		}
	}

	emptyFilter := firstEmptyCollegeFilter(t, colleges)
	empty, err := repo.ListCollegesOfEducation(context.Background(), emptyFilter)
	if err != nil {
		t.Fatalf("ListCollegesOfEducation(empty filter) error = %v", err)
	}
	if empty == nil {
		t.Fatal("ListCollegesOfEducation(empty filter) returned nil slice")
	}
	if len(empty) != 0 {
		t.Fatalf("expected zero matches, got %d", len(empty))
	}

	for _, name := range []string{
		"FCT College of Education, Zuba",
		"Federal College of Education (Special), Oyo",
		"Federal College of Education (T), Umunze",
		"Federal College of Education Ofeme-Ohuhu",
		"Isaac Jasper Boro COE, Sagbama",
		"Yusuf Maitama Sule College of Education & Advanced Studies, Ghari",
	} {
		college, ok := findCollegeByName(colleges, name)
		if !ok {
			t.Fatalf("missing expected college %q", name)
		}
		got, err := repo.GetCollegeOfEducation(context.Background(), college.ID)
		if err != nil {
			t.Fatalf("GetCollegeOfEducation(%s) error = %v", college.ID, err)
		}
		if got.ID != college.ID || got.Name != college.Name || got.StateID != college.StateID || got.OwnershipType != college.OwnershipType {
			t.Fatalf("unexpected college result for %q: %#v", name, got)
		}
	}

	if _, err := repo.GetCollegeOfEducation(context.Background(), "cross-river-state-coll-of-education-akampa"); !errors.Is(err, interfaces.ErrCollegeOfEducationNotFound) {
		t.Fatalf("missing college error = %v, want ErrCollegeOfEducationNotFound", err)
	}
}

func TestEducationRepositoryCollegeIsolationAndContext(t *testing.T) {
	t.Parallel()

	universities := loadApprovedUniversityFixture(t)
	colleges := loadApprovedCollegeFixture(t)

	stub := &educationJSONRepoStub{decodeFn: func(ctx context.Context, relativePath string, destination any) error {
		switch relativePath {
		case "education/universities.json":
			if dest, ok := destination.(*[]models.University); ok {
				*dest = cloneUniversityList(universities)
				return nil
			}
		case "education/colleges_of_education.json":
			if dest, ok := destination.(*[]models.CollegeOfEducation); ok {
				*dest = cloneCollegeOfEducationList(colleges)
				return nil
			}
		}
		return fmt.Errorf("unexpected decode request for %s", relativePath)
	}, pathCalls: map[string]int{}}

	repo, err := NewEducationRepository(stub, "education/universities.json", "education/colleges_of_education.json")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}

	stub.calls = 0
	stub.pathCalls = map[string]int{}
	if _, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{}); err != nil {
		t.Fatalf("ListUniversities() error = %v", err)
	}
	if stub.calls != 1 || stub.pathCalls["education/universities.json"] != 1 || stub.pathCalls["education/colleges_of_education.json"] != 0 {
		t.Fatalf("university call decoded unexpected paths: %#v", stub.pathCalls)
	}

	stub.calls = 0
	stub.pathCalls = map[string]int{}
	if _, err := repo.ListCollegesOfEducation(context.Background(), interfaces.CollegeOfEducationFilter{}); err != nil {
		t.Fatalf("ListCollegesOfEducation() error = %v", err)
	}
	if stub.calls != 1 || stub.pathCalls["education/colleges_of_education.json"] != 1 || stub.pathCalls["education/universities.json"] != 0 {
		t.Fatalf("college call decoded unexpected paths: %#v", stub.pathCalls)
	}

	universityOnlyStub := &educationJSONRepoStub{decodeFn: func(ctx context.Context, relativePath string, destination any) error {
		switch relativePath {
		case "education/universities.json":
			if dest, ok := destination.(*[]models.University); ok {
				*dest = cloneUniversityList(universities)
				return nil
			}
			return fmt.Errorf("unexpected destination %T", destination)
		case "education/colleges_of_education.json":
			return fmt.Errorf("/private/tmp/education/colleges_of_education.json: permission denied")
		default:
			return fmt.Errorf("unexpected path %s", relativePath)
		}
	}}
	repo, err = NewEducationRepository(universityOnlyStub, "education/universities.json", "education/colleges_of_education.json")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}
	if _, err := repo.ListUniversities(context.Background(), interfaces.UniversityFilter{}); err != nil {
		t.Fatalf("ListUniversities() with broken college file should succeed, got %v", err)
	}

	collegeOnlyStub := &educationJSONRepoStub{decodeFn: func(ctx context.Context, relativePath string, destination any) error {
		switch relativePath {
		case "education/colleges_of_education.json":
			if dest, ok := destination.(*[]models.CollegeOfEducation); ok {
				*dest = cloneCollegeOfEducationList(colleges)
				return nil
			}
			return fmt.Errorf("unexpected destination %T", destination)
		case "education/universities.json":
			return fmt.Errorf("/private/tmp/education/universities.json: permission denied")
		default:
			return fmt.Errorf("unexpected path %s", relativePath)
		}
	}}
	repo, err = NewEducationRepository(collegeOnlyStub, "education/universities.json", "education/colleges_of_education.json")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}
	if _, err := repo.ListCollegesOfEducation(context.Background(), interfaces.CollegeOfEducationFilter{}); err != nil {
		t.Fatalf("ListCollegesOfEducation() with broken university file should succeed, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.ListCollegesOfEducation(ctx, interfaces.CollegeOfEducationFilter{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context error = %v, want context.Canceled", err)
	}

	deadlineCtx, cancelDeadline := context.WithTimeout(context.Background(), time.Nanosecond)
	time.Sleep(2 * time.Nanosecond)
	cancelDeadline()
	if _, err := repo.GetCollegeOfEducation(deadlineCtx, colleges[0].ID); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline context error = %v, want context cancellation/deadline", err)
	}
}

func TestEducationRepositoryCollegeValidationRejectsInvalidFixtures(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedCollegeFixture(t)
	tests := []struct {
		name string
		mut  func([]models.CollegeOfEducation) []models.CollegeOfEducation
	}{
		{name: "nil slice", mut: func([]models.CollegeOfEducation) []models.CollegeOfEducation { return nil }},
		{name: "empty slice", mut: func([]models.CollegeOfEducation) []models.CollegeOfEducation {
			return make([]models.CollegeOfEducation, 0)
		}},
		{name: "243 records", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			return append([]models.CollegeOfEducation(nil), records[:243]...)
		}},
		{name: "245 records", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			return append(out, records[0])
		}},
		{name: "incorrect ownership count", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			out[0].OwnershipType = "private"
			return out
		}},
		{name: "invalid state id", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			out[0].StateID = "invalid"
			return out
		}},
		{name: "invalid country code", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			out[0].CountryCode = "GH"
			return out
		}},
		{name: "missing required fields", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			out[0].Name = ""
			return out
		}},
		{name: "duplicate id", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			out[1].ID = out[0].ID
			return out
		}},
		{name: "duplicate state/name pair", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			out[1].Name = out[0].Name
			out[1].ID = slugifyUniversityName(out[1].Name)
			out[1].StateID = out[0].StateID
			return out
		}},
		{name: "invalid id syntax", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			out[0].ID = "bad"
			return out
		}},
		{name: "id name mismatch", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			out[0].ID = "wrong-" + slugifyUniversityName(out[0].Name)
			return out
		}},
		{name: "incorrect ordering", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			out[0], out[1] = out[1], out[0]
			return out
		}},
		{name: "excluded row", mut: func(records []models.CollegeOfEducation) []models.CollegeOfEducation {
			out := append([]models.CollegeOfEducation(nil), records...)
			out[0].Name = "Cross River State Coll. of Education, Akampa"
			return out
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mustNewEducationRepositoryFromEducationFixtures(t, loadApprovedUniversityFixture(t), tc.mut(fixture))
			_, err := repo.ListCollegesOfEducation(context.Background(), interfaces.CollegeOfEducationFilter{})
			if err == nil {
				t.Fatalf("expected %s to fail", tc.name)
			}
			if !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("unexpected error for %s: %v", tc.name, err)
			}
		})
	}
}

func loadApprovedCollegeFixture(t *testing.T) []models.CollegeOfEducation {
	t.Helper()

	var colleges []models.CollegeOfEducation
	dec := json.NewDecoder(bytes.NewReader(readEducationDatasetBytes(t, educationDatasetPath("education", "colleges_of_education.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&colleges); err != nil {
		t.Fatalf("decode approved colleges dataset: %v", err)
	}
	return colleges
}

func mustNewEducationRepositoryFromEducationFixtures(t *testing.T, universities []models.University, colleges []models.CollegeOfEducation) *EducationFileRepository {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "education"), 0o755); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}

	universityData, err := json.MarshalIndent(universities, "", "  ")
	if err != nil {
		t.Fatalf("marshal universities fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "education", "universities.json"), append(universityData, '\n'), 0o600); err != nil {
		t.Fatalf("write universities fixture: %v", err)
	}

	collegeData, err := json.MarshalIndent(colleges, "", "  ")
	if err != nil {
		t.Fatalf("marshal colleges fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "education", "colleges_of_education.json"), append(collegeData, '\n'), 0o600); err != nil {
		t.Fatalf("write colleges fixture: %v", err)
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

func firstEmptyCollegeFilter(t *testing.T, colleges []models.CollegeOfEducation) interfaces.CollegeOfEducationFilter {
	t.Helper()

	counts := make(map[interfaces.CollegeOfEducationFilter]int)
	for _, college := range colleges {
		filter := interfaces.CollegeOfEducationFilter{OwnershipType: college.OwnershipType, StateID: college.StateID}
		counts[filter]++
	}
	for _, stateID := range []string{"abia", "adamawa", "akwa-ibom", "anambra", "bauchi", "bayelsa", "benue", "borno", "cross-river", "delta", "ebonyi", "edo", "ekiti", "enugu", "fct", "gombe", "imo", "jigawa", "kaduna", "kano", "katsina", "kebbi", "kogi", "kwara", "lagos", "nasarawa", "niger", "ogun", "ondo", "osun", "oyo", "plateau", "rivers", "sokoto", "taraba", "yobe", "zamfara"} {
		for _, ownershipType := range []string{"federal", "state", "private"} {
			filter := interfaces.CollegeOfEducationFilter{OwnershipType: ownershipType, StateID: stateID}
			if counts[filter] == 0 {
				return filter
			}
		}
	}
	return interfaces.CollegeOfEducationFilter{OwnershipType: "federal", StateID: "fct"}
}

func findCollegeByName(colleges []models.CollegeOfEducation, want string) (models.CollegeOfEducation, bool) {
	for _, college := range colleges {
		if college.Name == want {
			return college, true
		}
	}
	return models.CollegeOfEducation{}, false
}
