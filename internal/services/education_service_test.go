package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type educationRepositoryStub struct {
	listResult                                    map[interfaces.UniversityFilter][]models.University
	getResult                                     map[string]models.University
	listErr                                       error
	getErr                                        error
	listCalls                                     int
	getCalls                                      int
	lastFilter                                    interfaces.UniversityFilter
	lastID                                        string
	polytechnics                                  []models.Polytechnic
	polytechnicErr                                error
	polytechnicGet                                map[string]models.Polytechnic
	polytechnicGetErr                             error
	polytechnicCalls                              int
	polytechnicGetCalls                           int
	monotechnics                                  []models.Monotechnic
	monotechnicErr                                error
	monotechnicGet                                map[string]models.Monotechnic
	monotechnicGetErr                             error
	monotechnicCalls                              int
	monotechnicGetCalls                           int
	collegesOfAgriculture                         []models.CollegeOfAgriculture
	collegesOfAgricultureErr                      error
	collegesOfAgricultureGet                      map[string]models.CollegeOfAgriculture
	collegesOfAgricultureGetErr                   error
	collegesOfAgricultureCalls                    int
	collegesOfAgricultureGetCalls                 int
	collegesOfHealthSciencesAndTechnology         []models.CollegeOfHealthSciencesAndTechnology
	collegesOfHealthSciencesAndTechnologyErr      error
	collegesOfHealthSciencesAndTechnologyGet      map[string]models.CollegeOfHealthSciencesAndTechnology
	collegesOfHealthSciencesAndTechnologyGetErr   error
	collegesOfHealthSciencesAndTechnologyCalls    int
	collegesOfHealthSciencesAndTechnologyGetCalls int
	vocationalEnterpriseInstitutions              []models.VocationalEnterpriseInstitution
	vocationalEnterpriseInstitutionsErr           error
	vocationalEnterpriseInstitutionsGet           map[string]models.VocationalEnterpriseInstitution
	vocationalEnterpriseInstitutionsGetErr        error
	vocationalEnterpriseInstitutionsCalls         int
	vocationalEnterpriseInstitutionsGetCalls      int
	primaryAndSecondarySchoolListResult           interfaces.PrimaryAndSecondarySchoolListResult
	primaryAndSecondarySchoolListErr              error
	primaryAndSecondarySchoolGetResult            map[string]models.PrimaryAndSecondarySchool
	primaryAndSecondarySchoolGetErr               error
	primaryAndSecondarySchoolListCalls            int
	primaryAndSecondarySchoolGetCalls             int
	lastSchoolQuery                               interfaces.PrimaryAndSecondarySchoolQuery
}

func (s *educationRepositoryStub) ListUniversities(_ context.Context, filter interfaces.UniversityFilter) ([]models.University, error) {
	s.listCalls++
	s.lastFilter = filter
	if s.listErr != nil {
		return nil, s.listErr
	}
	if s.listResult != nil {
		if universities, ok := s.listResult[filter]; ok {
			return cloneUniversityList(universities), nil
		}
	}
	return nil, nil
}

func (s *educationRepositoryStub) GetUniversityByID(_ context.Context, universityID string) (models.University, error) {
	s.getCalls++
	s.lastID = universityID
	if s.getErr != nil {
		return models.University{}, s.getErr
	}
	if s.getResult != nil {
		if university, ok := s.getResult[universityID]; ok {
			return university, nil
		}
	}
	return models.University{}, interfaces.ErrUniversityNotFound
}

func (s *educationRepositoryStub) ListCollegesOfEducation(_ context.Context, filter interfaces.CollegeOfEducationFilter) ([]models.CollegeOfEducation, error) {
	s.listCalls++
	return nil, nil
}

func (s *educationRepositoryStub) GetCollegeOfEducation(_ context.Context, collegeID string) (models.CollegeOfEducation, error) {
	s.getCalls++
	s.lastID = collegeID
	return models.CollegeOfEducation{}, interfaces.ErrCollegeOfEducationNotFound
}

func (s *educationRepositoryStub) ListPolytechnics(context.Context) ([]models.Polytechnic, error) {
	s.polytechnicCalls++
	if s.polytechnicErr != nil {
		return nil, s.polytechnicErr
	}
	return cloneServiceSlice(s.polytechnics), nil
}

func (s *educationRepositoryStub) GetPolytechnic(_ context.Context, id string) (models.Polytechnic, error) {
	s.polytechnicGetCalls++
	if s.polytechnicGetErr != nil {
		return models.Polytechnic{}, s.polytechnicGetErr
	}
	if s.polytechnicGet != nil {
		if row, ok := s.polytechnicGet[strings.TrimSpace(id)]; ok {
			return row, nil
		}
	}
	return models.Polytechnic{}, interfaces.ErrPolytechnicNotFound
}

func (s *educationRepositoryStub) ListMonotechnics(context.Context) ([]models.Monotechnic, error) {
	s.monotechnicCalls++
	if s.monotechnicErr != nil {
		return nil, s.monotechnicErr
	}
	return cloneServiceSlice(s.monotechnics), nil
}

func (s *educationRepositoryStub) GetMonotechnic(_ context.Context, id string) (models.Monotechnic, error) {
	s.monotechnicGetCalls++
	if s.monotechnicGetErr != nil {
		return models.Monotechnic{}, s.monotechnicGetErr
	}
	if s.monotechnicGet != nil {
		if row, ok := s.monotechnicGet[strings.TrimSpace(id)]; ok {
			return row, nil
		}
	}
	return models.Monotechnic{}, interfaces.ErrMonotechnicNotFound
}

func (s *educationRepositoryStub) ListCollegesOfAgriculture(context.Context) ([]models.CollegeOfAgriculture, error) {
	s.collegesOfAgricultureCalls++
	if s.collegesOfAgricultureErr != nil {
		return nil, s.collegesOfAgricultureErr
	}
	return cloneServiceSlice(s.collegesOfAgriculture), nil
}

func (s *educationRepositoryStub) GetCollegeOfAgriculture(_ context.Context, id string) (models.CollegeOfAgriculture, error) {
	s.collegesOfAgricultureGetCalls++
	if s.collegesOfAgricultureGetErr != nil {
		return models.CollegeOfAgriculture{}, s.collegesOfAgricultureGetErr
	}
	if s.collegesOfAgricultureGet != nil {
		if row, ok := s.collegesOfAgricultureGet[strings.TrimSpace(id)]; ok {
			return row, nil
		}
	}
	return models.CollegeOfAgriculture{}, interfaces.ErrCollegeOfAgricultureNotFound
}

func (s *educationRepositoryStub) ListCollegesOfHealthSciencesAndTechnology(context.Context) ([]models.CollegeOfHealthSciencesAndTechnology, error) {
	s.collegesOfHealthSciencesAndTechnologyCalls++
	if s.collegesOfHealthSciencesAndTechnologyErr != nil {
		return nil, s.collegesOfHealthSciencesAndTechnologyErr
	}
	return cloneServiceSlice(s.collegesOfHealthSciencesAndTechnology), nil
}

func (s *educationRepositoryStub) GetCollegeOfHealthSciencesAndTechnology(_ context.Context, id string) (models.CollegeOfHealthSciencesAndTechnology, error) {
	s.collegesOfHealthSciencesAndTechnologyGetCalls++
	if s.collegesOfHealthSciencesAndTechnologyGetErr != nil {
		return models.CollegeOfHealthSciencesAndTechnology{}, s.collegesOfHealthSciencesAndTechnologyGetErr
	}
	if s.collegesOfHealthSciencesAndTechnologyGet != nil {
		if row, ok := s.collegesOfHealthSciencesAndTechnologyGet[strings.TrimSpace(id)]; ok {
			return row, nil
		}
	}
	return models.CollegeOfHealthSciencesAndTechnology{}, interfaces.ErrCollegeOfHealthSciencesAndTechnologyNotFound
}

func (s *educationRepositoryStub) ListVocationalEnterpriseInstitutions(context.Context) ([]models.VocationalEnterpriseInstitution, error) {
	s.vocationalEnterpriseInstitutionsCalls++
	if s.vocationalEnterpriseInstitutionsErr != nil {
		return nil, s.vocationalEnterpriseInstitutionsErr
	}
	return cloneServiceSlice(s.vocationalEnterpriseInstitutions), nil
}

func (s *educationRepositoryStub) GetVocationalEnterpriseInstitution(_ context.Context, id string) (models.VocationalEnterpriseInstitution, error) {
	s.vocationalEnterpriseInstitutionsGetCalls++
	if s.vocationalEnterpriseInstitutionsGetErr != nil {
		return models.VocationalEnterpriseInstitution{}, s.vocationalEnterpriseInstitutionsGetErr
	}
	if s.vocationalEnterpriseInstitutionsGet != nil {
		if row, ok := s.vocationalEnterpriseInstitutionsGet[strings.TrimSpace(id)]; ok {
			return row, nil
		}
	}
	return models.VocationalEnterpriseInstitution{}, interfaces.ErrVocationalEnterpriseInstitutionNotFound
}

func (s *educationRepositoryStub) ListPrimaryAndSecondarySchools(context.Context, interfaces.PrimaryAndSecondarySchoolQuery) (interfaces.PrimaryAndSecondarySchoolListResult, error) {
	s.primaryAndSecondarySchoolListCalls++
	if s.primaryAndSecondarySchoolListErr != nil {
		return interfaces.PrimaryAndSecondarySchoolListResult{}, s.primaryAndSecondarySchoolListErr
	}
	return s.primaryAndSecondarySchoolListResult, nil
}

func (s *educationRepositoryStub) GetPrimaryAndSecondarySchool(_ context.Context, id string) (models.PrimaryAndSecondarySchool, error) {
	s.primaryAndSecondarySchoolGetCalls++
	if s.primaryAndSecondarySchoolGetErr != nil {
		return models.PrimaryAndSecondarySchool{}, s.primaryAndSecondarySchoolGetErr
	}
	if s.primaryAndSecondarySchoolGetResult != nil {
		if row, ok := s.primaryAndSecondarySchoolGetResult[strings.TrimSpace(id)]; ok {
			return row, nil
		}
	}
	return models.PrimaryAndSecondarySchool{}, interfaces.ErrPrimaryAndSecondarySchoolNotFound
}

func cloneServiceSlice[T any](items []T) []T {
	if len(items) == 0 {
		return make([]T, 0)
	}
	cloned := make([]T, len(items))
	copy(cloned, items)
	return cloned
}

func TestNewEducationServiceRejectsNilRepository(t *testing.T) {
	t.Parallel()

	if _, err := NewEducationService(nil); err == nil {
		t.Fatal("expected nil repository to be rejected")
	}
}

func TestEducationServiceListUniversities(t *testing.T) {
	t.Parallel()

	all := []models.University{
		{ID: "abubakar-tafawa-balewa-university-bauchi", Name: "Abubakar Tafawa Balewa University, Bauchi", OwnershipType: "federal", StateID: "bauchi", CountryCode: "NG"},
		{ID: "baze-university", Name: "Baze University", OwnershipType: "private", StateID: "fct", CountryCode: "NG"},
	}
	stub := &educationRepositoryStub{
		listResult: map[interfaces.UniversityFilter][]models.University{
			{}:                         all,
			{OwnershipType: "federal"}: []models.University{all[0]},
			{OwnershipType: "state"}: []models.University{
				{ID: "gombe-state-university-gombe", Name: "Gombe State University, Gombe", OwnershipType: "state", StateID: "gombe", CountryCode: "NG"},
			},
			{OwnershipType: "private", StateID: "lagos"}: []models.University{
				{ID: "caleb-university-lagos", Name: "Caleb University, Lagos", OwnershipType: "private", StateID: "lagos", CountryCode: "NG"},
			},
			{StateID: "taraba"}: []models.University{
				{ID: "taraba-state-university-of-tropical-agriculture-science-technology-and-climate-action-gembu", Name: "Taraba State University of Tropical Agriculture, Science, Technology and Climate Action, Gembu", OwnershipType: "state", StateID: "taraba", CountryCode: "NG"},
			},
			{StateID: "fct"}: []models.University{
				{ID: "baze-university", Name: "Baze University", OwnershipType: "private", StateID: "fct", CountryCode: "NG"},
			},
			{OwnershipType: "state", StateID: "fct"}: []models.University{},
		},
	}
	svc, err := NewEducationService(stub)
	if err != nil {
		t.Fatalf("NewEducationService() error = %v", err)
	}

	cases := []struct {
		name  string
		input UniversityListInput
		want  int
	}{
		{name: "no filters", input: UniversityListInput{}, want: 2},
		{name: "federal", input: UniversityListInput{OwnershipType: " federal "}, want: 1},
		{name: "state", input: UniversityListInput{OwnershipType: "state"}, want: 1},
		{name: "private lagos", input: UniversityListInput{OwnershipType: " private ", StateID: " lagos "}, want: 1},
		{name: "taraba", input: UniversityListInput{StateID: " taraba "}, want: 1},
		{name: "fct", input: UniversityListInput{StateID: "fct"}, want: 1},
		{name: "empty combined", input: UniversityListInput{OwnershipType: "state", StateID: "fct"}, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub.listCalls = 0
			listed, err := svc.ListUniversities(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("ListUniversities() error = %v", err)
			}
			if stub.listCalls != 1 {
				t.Fatalf("unexpected repository call count: %d", stub.listCalls)
			}
			if tc.input.OwnershipType == " federal " && stub.lastFilter.OwnershipType != "federal" {
				t.Fatalf("unexpected canonical ownership filter: %#v", stub.lastFilter)
			}
			if tc.input.StateID == " taraba " && stub.lastFilter.StateID != "taraba" {
				t.Fatalf("unexpected canonical state filter: %#v", stub.lastFilter)
			}
			if listed == nil {
				t.Fatal("ListUniversities() returned nil slice")
			}
			if got := len(listed); got != tc.want {
				t.Fatalf("unexpected record count: got %d want %d", got, tc.want)
			}
			if len(listed) > 0 {
				listed[0].Name = "Changed"
				again, err := svc.ListUniversities(context.Background(), tc.input)
				if err != nil {
					t.Fatalf("ListUniversities() second call error = %v", err)
				}
				if again[0].Name == "Changed" {
					t.Fatal("ListUniversities() exposed shared mutable slice state")
				}
			}
		})
	}
}

func TestEducationServiceListUniversitiesNilResultBecomesEmpty(t *testing.T) {
	t.Parallel()

	stub := &educationRepositoryStub{}
	svc, err := NewEducationService(stub)
	if err != nil {
		t.Fatalf("NewEducationService() error = %v", err)
	}

	listed, err := svc.ListUniversities(context.Background(), UniversityListInput{})
	if err != nil {
		t.Fatalf("ListUniversities() error = %v", err)
	}
	if listed == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(listed) != 0 {
		t.Fatalf("unexpected length: %d", len(listed))
	}
}

func TestEducationServiceRejectsInvalidFiltersBeforeRepositoryAccess(t *testing.T) {
	t.Parallel()

	stub := &educationRepositoryStub{}
	svc, err := NewEducationService(stub)
	if err != nil {
		t.Fatalf("NewEducationService() error = %v", err)
	}

	for _, input := range []UniversityListInput{
		{OwnershipType: "public"},
		{OwnershipType: "PRIVATE"},
		{StateID: "Abia"},
		{StateID: "invalid"},
		{OwnershipType: "state", StateID: "Abia"},
	} {
		if _, err := svc.ListUniversities(context.Background(), input); !errors.Is(err, ErrInvalidUniversityOwnershipType) && !errors.Is(err, ErrInvalidUniversityStateID) {
			t.Fatalf("unexpected validation error for %#v: %v", input, err)
		}
	}
	if stub.listCalls != 0 {
		t.Fatalf("repository was called for invalid inputs: %d", stub.listCalls)
	}
}

func TestEducationServiceGetUniversity(t *testing.T) {
	t.Parallel()

	stub := &educationRepositoryStub{
		getResult: map[string]models.University{
			"baze-university": {
				ID:            "baze-university",
				Name:          "Baze University",
				OwnershipType: "private",
				StateID:       "fct",
				CountryCode:   "NG",
			},
		},
	}
	svc, err := NewEducationService(stub)
	if err != nil {
		t.Fatalf("NewEducationService() error = %v", err)
	}

	university, err := svc.GetUniversity(context.Background(), "  baze-university  ")
	if err != nil {
		t.Fatalf("GetUniversity() error = %v", err)
	}
	if stub.getCalls != 1 {
		t.Fatalf("unexpected repository call count: %d", stub.getCalls)
	}
	if stub.lastID != "baze-university" {
		t.Fatalf("unexpected repository lookup id: %q", stub.lastID)
	}
	if university.ID != "baze-university" || university.Name != "Baze University" {
		t.Fatalf("unexpected university response: %#v", university)
	}

	data, err := json.Marshal(university)
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

	for _, input := range []string{"", "   ", "A", "bad id", "bad_id", "-bad", "bad-", "bad--id", "Bad-ID", "baze/university", "Baze University", "baze\x00university"} {
		if _, err := svc.GetUniversity(context.Background(), input); !errors.Is(err, ErrInvalidUniversityID) {
			t.Fatalf("GetUniversity(%q) error = %v, want ErrInvalidUniversityID", input, err)
		}
	}
	if stub.getCalls != 1 {
		t.Fatalf("repository was called for invalid ids: %d", stub.getCalls)
	}
}

func TestEducationServiceErrorTranslationAndContextPreservation(t *testing.T) {
	t.Parallel()

	t.Run("not found", func(t *testing.T) {
		stub := &educationRepositoryStub{getErr: interfaces.ErrUniversityNotFound}
		svc, err := NewEducationService(stub)
		if err != nil {
			t.Fatalf("NewEducationService() error = %v", err)
		}
		if _, err := svc.GetUniversity(context.Background(), "baze-university"); !errors.Is(err, ErrUniversityNotFound) {
			t.Fatalf("unexpected not-found translation: %v", err)
		}
	})

	t.Run("wrapped not found", func(t *testing.T) {
		stub := &educationRepositoryStub{getErr: fmt.Errorf("wrapped: %w", interfaces.ErrUniversityNotFound)}
		svc, err := NewEducationService(stub)
		if err != nil {
			t.Fatalf("NewEducationService() error = %v", err)
		}
		if _, err := svc.GetUniversity(context.Background(), "baze-university"); !errors.Is(err, ErrUniversityNotFound) {
			t.Fatalf("unexpected wrapped not-found translation: %v", err)
		}
	})

	t.Run("unexpected repository failure", func(t *testing.T) {
		stub := &educationRepositoryStub{getErr: fmt.Errorf("/private/tmp/education/universities.json: permission denied")}
		svc, err := NewEducationService(stub)
		if err != nil {
			t.Fatalf("NewEducationService() error = %v", err)
		}
		_, err = svc.GetUniversity(context.Background(), "baze-university")
		if err == nil {
			t.Fatal("expected sanitized repository failure")
		}
		if strings.Contains(err.Error(), "/private/tmp/education/universities.json") {
			t.Fatalf("error leaked filesystem path: %v", err)
		}
		if !strings.Contains(err.Error(), "repository unavailable") {
			t.Fatalf("unexpected sanitized error: %v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		stub := &educationRepositoryStub{getResult: map[string]models.University{}}
		svc, err := NewEducationService(stub)
		if err != nil {
			t.Fatalf("NewEducationService() error = %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := svc.ListUniversities(ctx, UniversityListInput{}); !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled list context error = %v, want context.Canceled", err)
		}
		if _, err := svc.GetUniversity(ctx, "baze-university"); !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled get context error = %v, want context.Canceled", err)
		}
	})

	t.Run("context deadline", func(t *testing.T) {
		stub := &educationRepositoryStub{listResult: map[interfaces.UniversityFilter][]models.University{}, getResult: map[string]models.University{}}
		svc, err := NewEducationService(stub)
		if err != nil {
			t.Fatalf("NewEducationService() error = %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		time.Sleep(2 * time.Nanosecond)
		cancel()
		if _, err := svc.ListUniversities(ctx, UniversityListInput{}); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline list context error = %v", err)
		}
		if _, err := svc.GetUniversity(ctx, "baze-university"); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline get context error = %v", err)
		}
	})
}
