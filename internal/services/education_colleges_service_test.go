package services

import (
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

type collegeServiceRepositoryStub struct {
	collegeListResult map[interfaces.CollegeOfEducationFilter][]models.CollegeOfEducation
	collegeGetResult  map[string]models.CollegeOfEducation
	collegeListErr    error
	collegeGetErr     error
	collegeListCalls  int
	collegeGetCalls   int
	lastCollegeInput  interfaces.CollegeOfEducationFilter
	lastCollegeID     string
}

func (s *collegeServiceRepositoryStub) ListUniversities(context.Context, interfaces.UniversityFilter) ([]models.University, error) {
	return nil, nil
}

func (s *collegeServiceRepositoryStub) GetUniversityByID(context.Context, string) (models.University, error) {
	return models.University{}, interfaces.ErrUniversityNotFound
}

func (s *collegeServiceRepositoryStub) ListCollegesOfEducation(_ context.Context, filter interfaces.CollegeOfEducationFilter) ([]models.CollegeOfEducation, error) {
	s.collegeListCalls++
	s.lastCollegeInput = filter
	if s.collegeListErr != nil {
		return nil, s.collegeListErr
	}
	if s.collegeListResult != nil {
		if colleges, ok := s.collegeListResult[filter]; ok {
			return cloneCollegeOfEducationList(colleges), nil
		}
	}
	return nil, nil
}

func (s *collegeServiceRepositoryStub) GetCollegeOfEducation(_ context.Context, collegeID string) (models.CollegeOfEducation, error) {
	s.collegeGetCalls++
	s.lastCollegeID = collegeID
	if s.collegeGetErr != nil {
		return models.CollegeOfEducation{}, s.collegeGetErr
	}
	if s.collegeGetResult != nil {
		if college, ok := s.collegeGetResult[collegeID]; ok {
			return college, nil
		}
	}
	return models.CollegeOfEducation{}, interfaces.ErrCollegeOfEducationNotFound
}

func TestEducationServiceListCollegesOfEducation(t *testing.T) {
	t.Parallel()

	colleges := loadApprovedCollegeFixture(t)
	firstCollege, secondCollege := colleges[0], colleges[1]
	lagosPrivate := firstCollege
	lagosPrivate.StateID = "lagos"
	lagosPrivate.OwnershipType = "private"

	stub := &collegeServiceRepositoryStub{
		collegeListResult: map[interfaces.CollegeOfEducationFilter][]models.CollegeOfEducation{
			{}:                         {firstCollege, secondCollege},
			{OwnershipType: "federal"}: {firstCollege},
			{OwnershipType: "state"}:   {secondCollege},
			{OwnershipType: "private", StateID: "lagos"}: {lagosPrivate},
			{StateID: "fct"}:                                   {firstCollege},
			{OwnershipType: "state", StateID: "fct"}:           {},
			{OwnershipType: "federal", StateID: "abia"}:        {},
			{OwnershipType: "private", StateID: "abia"}:        {},
			{OwnershipType: "private", StateID: "cross-river"}: {},
			{OwnershipType: "federal", StateID: "cross-river"}: {},
			{OwnershipType: "state", StateID: "cross-river"}:   {},
		},
	}
	svc, err := NewEducationService(stub)
	if err != nil {
		t.Fatalf("NewEducationService() error = %v", err)
	}

	cases := []struct {
		name  string
		input CollegeOfEducationListInput
		want  int
	}{
		{name: "no filters", input: CollegeOfEducationListInput{}, want: 2},
		{name: "federal", input: CollegeOfEducationListInput{OwnershipType: " federal "}, want: 1},
		{name: "state", input: CollegeOfEducationListInput{OwnershipType: "state"}, want: 1},
		{name: "private lagos", input: CollegeOfEducationListInput{OwnershipType: " private ", StateID: " lagos "}, want: 1},
		{name: "fct", input: CollegeOfEducationListInput{StateID: " fct "}, want: 1},
		{name: "empty combined", input: CollegeOfEducationListInput{OwnershipType: "state", StateID: "fct"}, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub.collegeListCalls = 0
			listed, err := svc.ListCollegesOfEducation(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("ListCollegesOfEducation() error = %v", err)
			}
			if stub.collegeListCalls != 1 {
				t.Fatalf("unexpected repository call count: %d", stub.collegeListCalls)
			}
			if listed == nil {
				t.Fatal("ListCollegesOfEducation() returned nil slice")
			}
			if got := len(listed); got != tc.want {
				t.Fatalf("unexpected record count: got %d want %d", got, tc.want)
			}
			if len(listed) > 0 {
				listed[0].Name = "Changed"
				again, err := svc.ListCollegesOfEducation(context.Background(), tc.input)
				if err != nil {
					t.Fatalf("ListCollegesOfEducation() second call error = %v", err)
				}
				if again[0].Name == "Changed" {
					t.Fatal("ListCollegesOfEducation() exposed shared mutable slice state")
				}
			}
		})
	}
}

func TestEducationServiceCollegeListValidationAndEmptyNormalization(t *testing.T) {
	t.Parallel()

	stub := &collegeServiceRepositoryStub{}
	svc, err := NewEducationService(stub)
	if err != nil {
		t.Fatalf("NewEducationService() error = %v", err)
	}

	listed, err := svc.ListCollegesOfEducation(context.Background(), CollegeOfEducationListInput{})
	if err != nil {
		t.Fatalf("ListCollegesOfEducation() error = %v", err)
	}
	if listed == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(listed) != 0 {
		t.Fatalf("unexpected length: %d", len(listed))
	}

	for _, input := range []CollegeOfEducationListInput{
		{OwnershipType: "public"},
		{OwnershipType: "PRIVATE"},
		{StateID: "Abia"},
		{StateID: "invalid"},
		{OwnershipType: "state", StateID: "Abia"},
	} {
		if _, err := svc.ListCollegesOfEducation(context.Background(), input); !errors.Is(err, ErrInvalidCollegeOfEducationOwnershipType) && !errors.Is(err, ErrInvalidCollegeOfEducationStateID) {
			t.Fatalf("unexpected validation error for %#v: %v", input, err)
		}
	}
	if stub.collegeListCalls != 1 {
		t.Fatalf("repository was called for invalid inputs: %d", stub.collegeListCalls)
	}
}

func TestEducationServiceGetCollegeOfEducation(t *testing.T) {
	t.Parallel()

	college := loadCollegeFixtureByName(t, "FCT College of Education, Zuba")
	stub := &collegeServiceRepositoryStub{
		collegeGetResult: map[string]models.CollegeOfEducation{
			college.ID: college,
		},
	}
	svc, err := NewEducationService(stub)
	if err != nil {
		t.Fatalf("NewEducationService() error = %v", err)
	}

	got, err := svc.GetCollegeOfEducation(context.Background(), "  "+college.ID+"  ")
	if err != nil {
		t.Fatalf("GetCollegeOfEducation() error = %v", err)
	}
	if stub.collegeGetCalls != 1 {
		t.Fatalf("unexpected repository call count: %d", stub.collegeGetCalls)
	}
	if stub.lastCollegeID != college.ID {
		t.Fatalf("unexpected repository lookup id: %q", stub.lastCollegeID)
	}
	if got.ID != college.ID || got.Name != college.Name || got.StateID != college.StateID || got.OwnershipType != college.OwnershipType {
		t.Fatalf("unexpected college response: %#v", got)
	}

	data, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(fields) != 5 {
		t.Fatalf("unexpected field count: %d", len(fields))
	}

	for _, input := range []string{"", "   ", "A", "bad id", "bad_id", "-bad", "bad-", "bad--id", "Bad-ID", "bad/id", "Bad College", "bad\x00id"} {
		if _, err := svc.GetCollegeOfEducation(context.Background(), input); !errors.Is(err, ErrInvalidCollegeOfEducationID) {
			t.Fatalf("GetCollegeOfEducation(%q) error = %v, want ErrInvalidCollegeOfEducationID", input, err)
		}
	}
	if stub.collegeGetCalls != 1 {
		t.Fatalf("repository was called for invalid ids: %d", stub.collegeGetCalls)
	}
}

func TestEducationServiceCollegeErrorTranslationAndContextPreservation(t *testing.T) {
	t.Parallel()

	t.Run("not found", func(t *testing.T) {
		stub := &collegeServiceRepositoryStub{collegeGetErr: interfaces.ErrCollegeOfEducationNotFound}
		svc, err := NewEducationService(stub)
		if err != nil {
			t.Fatalf("NewEducationService() error = %v", err)
		}
		if _, err := svc.GetCollegeOfEducation(context.Background(), "fct-college-of-education-zuba"); !errors.Is(err, ErrCollegeOfEducationNotFound) {
			t.Fatalf("unexpected not-found translation: %v", err)
		}
	})

	t.Run("wrapped not found", func(t *testing.T) {
		stub := &collegeServiceRepositoryStub{collegeGetErr: fmt.Errorf("wrapped: %w", interfaces.ErrCollegeOfEducationNotFound)}
		svc, err := NewEducationService(stub)
		if err != nil {
			t.Fatalf("NewEducationService() error = %v", err)
		}
		if _, err := svc.GetCollegeOfEducation(context.Background(), "fct-college-of-education-zuba"); !errors.Is(err, ErrCollegeOfEducationNotFound) {
			t.Fatalf("unexpected wrapped not-found translation: %v", err)
		}
	})

	t.Run("unexpected repository failure", func(t *testing.T) {
		stub := &collegeServiceRepositoryStub{collegeGetErr: errors.New("/private/tmp/education/colleges_of_education.json: permission denied")}
		svc, err := NewEducationService(stub)
		if err != nil {
			t.Fatalf("NewEducationService() error = %v", err)
		}
		_, err = svc.GetCollegeOfEducation(context.Background(), "fct-college-of-education-zuba")
		if err == nil {
			t.Fatal("expected sanitized repository failure")
		}
		if strings.Contains(err.Error(), "/private/tmp/education/colleges_of_education.json") {
			t.Fatalf("error leaked filesystem path: %v", err)
		}
		if !strings.Contains(err.Error(), "repository unavailable") {
			t.Fatalf("unexpected sanitized error: %v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		stub := &collegeServiceRepositoryStub{collegeListResult: map[interfaces.CollegeOfEducationFilter][]models.CollegeOfEducation{}}
		svc, err := NewEducationService(stub)
		if err != nil {
			t.Fatalf("NewEducationService() error = %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := svc.ListCollegesOfEducation(ctx, CollegeOfEducationListInput{}); !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled list context error = %v, want context.Canceled", err)
		}
		if _, err := svc.GetCollegeOfEducation(ctx, "fct-college-of-education-zuba"); !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled get context error = %v, want context.Canceled", err)
		}
	})

	t.Run("context deadline", func(t *testing.T) {
		stub := &collegeServiceRepositoryStub{collegeListResult: map[interfaces.CollegeOfEducationFilter][]models.CollegeOfEducation{}, collegeGetResult: map[string]models.CollegeOfEducation{}}
		svc, err := NewEducationService(stub)
		if err != nil {
			t.Fatalf("NewEducationService() error = %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		time.Sleep(2 * time.Nanosecond)
		cancel()
		if _, err := svc.ListCollegesOfEducation(ctx, CollegeOfEducationListInput{}); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline list context error = %v", err)
		}
		if _, err := svc.GetCollegeOfEducation(ctx, "fct-college-of-education-zuba"); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline get context error = %v", err)
		}
	})
}

func loadApprovedCollegeFixture(t *testing.T) []models.CollegeOfEducation {
	t.Helper()

	var colleges []models.CollegeOfEducation
	data, err := os.ReadFile(filepath.Clean("../../datasets/education/colleges_of_education.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := json.Unmarshal(data, &colleges); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return colleges
}

func loadCollegeFixtureByName(t *testing.T, want string) models.CollegeOfEducation {
	t.Helper()

	for _, college := range loadApprovedCollegeFixture(t) {
		if college.Name == want {
			return college
		}
	}
	t.Fatalf("missing expected college %q", want)
	return models.CollegeOfEducation{}
}
