package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type educationHandlerStub struct {
	listFn        func(context.Context, services.UniversityListInput) ([]models.University, error)
	getFn         func(context.Context, string) (models.University, error)
	collegeListFn func(context.Context, services.CollegeOfEducationListInput) ([]models.CollegeOfEducation, error)
	collegeGetFn  func(context.Context, string) (models.CollegeOfEducation, error)

	listCalls        int
	getCalls         int
	collegeListCalls int
	collegeGetCalls  int
	lastInput        services.UniversityListInput
	lastID           string
	lastCollegeInput services.CollegeOfEducationListInput
	lastCollegeID    string
}

func (s *educationHandlerStub) ListUniversities(ctx context.Context, input services.UniversityListInput) ([]models.University, error) {
	s.listCalls++
	s.lastInput = input
	if s.listFn != nil {
		return s.listFn(ctx, input)
	}
	return nil, nil
}

func (s *educationHandlerStub) GetUniversity(ctx context.Context, universityID string) (models.University, error) {
	s.getCalls++
	s.lastID = universityID
	if s.getFn != nil {
		return s.getFn(ctx, universityID)
	}
	return models.University{}, nil
}

func (s *educationHandlerStub) ListCollegesOfEducation(ctx context.Context, input services.CollegeOfEducationListInput) ([]models.CollegeOfEducation, error) {
	s.collegeListCalls++
	s.lastCollegeInput = input
	if s.collegeListFn != nil {
		return s.collegeListFn(ctx, input)
	}
	return nil, nil
}

func (s *educationHandlerStub) GetCollegeOfEducation(ctx context.Context, collegeID string) (models.CollegeOfEducation, error) {
	s.collegeGetCalls++
	s.lastCollegeID = collegeID
	if s.collegeGetFn != nil {
		return s.collegeGetFn(ctx, collegeID)
	}
	return models.CollegeOfEducation{}, nil
}

func TestNewEducationHandlerRejectsNilService(t *testing.T) {
	if _, err := NewEducationHandler(nil); err == nil {
		t.Fatal("NewEducationHandler(nil) error = nil, want error")
	}
}

func TestEducationHandlerListUniversities(t *testing.T) {
	t.Run("list all", func(t *testing.T) {
		stub := &educationHandlerStub{
			listFn: func(context.Context, services.UniversityListInput) ([]models.University, error) {
				return []models.University{
					{ID: "ahmadu-bello-university-zaria", Name: "Ahmadu Bello University, Zaria", OwnershipType: "federal", StateID: "kaduna", CountryCode: "NG"},
					{ID: "university-of-lagos", Name: "University of Lagos", OwnershipType: "federal", StateID: "lagos", CountryCode: "NG"},
				}, nil
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/universities", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListUniversities(w, r)
		}, rr)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.listCalls != 1 {
			t.Fatalf("unexpected list call count: %d", stub.listCalls)
		}
		if stub.lastInput != (services.UniversityListInput{}) {
			t.Fatalf("unexpected input: %#v", stub.lastInput)
		}

		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].([]any)
		if len(data) != 2 {
			t.Fatalf("unexpected list payload: %#v", data)
		}
		first := data[0].(map[string]any)
		if len(first) != 5 || first["id"] != "ahmadu-bello-university-zaria" || first["state_id"] != "kaduna" {
			t.Fatalf("unexpected first university payload: %#v", first)
		}
	})

	t.Run("normalize nil slice", func(t *testing.T) {
		stub := &educationHandlerStub{
			listFn: func(context.Context, services.UniversityListInput) ([]models.University, error) {
				return nil, nil
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/universities", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListUniversities(w, r)
		}, rr)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}

		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].([]any)
		if len(data) != 0 {
			t.Fatalf("expected empty array, got %#v", data)
		}
	})

	t.Run("filter by ownership and state", func(t *testing.T) {
		stub := &educationHandlerStub{
			listFn: func(ctx context.Context, input services.UniversityListInput) ([]models.University, error) {
				if input.OwnershipType != "private" || input.StateID != "fct" {
					t.Fatalf("unexpected input: %#v", input)
				}
				return []models.University{{ID: "kwararafa-university", Name: "Kwararafa University, Wukari", OwnershipType: "private", StateID: "taraba", CountryCode: "NG"}}, nil
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/universities?ownership_type=private&state_id=fct", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListUniversities(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.listCalls != 1 {
			t.Fatalf("unexpected list call count: %d", stub.listCalls)
		}
	})

	t.Run("filter by state only", func(t *testing.T) {
		stub := &educationHandlerStub{
			listFn: func(ctx context.Context, input services.UniversityListInput) ([]models.University, error) {
				if input.OwnershipType != "" || input.StateID != "taraba" {
					t.Fatalf("unexpected input: %#v", input)
				}
				return []models.University{{
					ID:            "kwararafa-university",
					Name:          "Kwararafa University, Wukari",
					OwnershipType: "private",
					StateID:       "taraba",
					CountryCode:   "NG",
				}}, nil
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/universities?state_id=taraba", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListUniversities(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.listCalls != 1 {
			t.Fatalf("unexpected list call count: %d", stub.listCalls)
		}
	})

	t.Run("reject invalid query", func(t *testing.T) {
		stub := &educationHandlerStub{}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/universities?ownership_type=Federal", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListUniversities(w, r)
		}, rr)
		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.listCalls != 0 {
			t.Fatalf("service should not be called for invalid query")
		}
		if !strings.Contains(rr.Body.String(), "req_test") {
			t.Fatalf("request id missing from validation response: %s", rr.Body.String())
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewEducationHandler(&educationHandlerStub{})
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/education/universities", nil)
		h.ListUniversities(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestEducationHandlerGetUniversity(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		stub := &educationHandlerStub{
			getFn: func(ctx context.Context, universityID string) (models.University, error) {
				if universityID != "rev-fr-moses-orshio-adasu-makurdi" {
					t.Fatalf("unexpected university id: %q", universityID)
				}
				return models.University{
					ID:            universityID,
					Name:          "Rev. Fr. Moses Orshio Adasu University, Makurdi",
					OwnershipType: "state",
					StateID:       "benue",
					CountryCode:   "NG",
				}, nil
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/universities/rev-fr-moses-orshio-adasu-makurdi", nil)
		req.SetPathValue("university_id", " rev-fr-moses-orshio-adasu-makurdi ")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetUniversity(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.getCalls != 1 {
			t.Fatalf("unexpected get call count: %d", stub.getCalls)
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		item := body["data"].(map[string]any)
		if len(item) != 5 || item["name"] != "Rev. Fr. Moses Orshio Adasu University, Makurdi" || item["state_id"] != "benue" {
			t.Fatalf("unexpected university payload: %#v", item)
		}
	})

	t.Run("reject invalid id", func(t *testing.T) {
		stub := &educationHandlerStub{}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/universities/Abia-University", nil)
		req.SetPathValue("university_id", "Abia-University")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetUniversity(w, r)
		}, rr)
		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.getCalls != 0 {
			t.Fatalf("service should not be called for invalid id")
		}
	})

	t.Run("not found", func(t *testing.T) {
		stub := &educationHandlerStub{
			getFn: func(context.Context, string) (models.University, error) {
				return models.University{}, services.ErrUniversityNotFound
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/universities/non-existent-university", nil)
		req.SetPathValue("university_id", "non-existent-university")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetUniversity(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.getCalls != 1 {
			t.Fatalf("unexpected get call count: %d", stub.getCalls)
		}
		if !strings.Contains(rr.Body.String(), "req_test") {
			t.Fatalf("request id missing from error response: %s", rr.Body.String())
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewEducationHandler(&educationHandlerStub{})
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/education/universities/rev-fr-moses-orshio-adasu-makurdi", nil)
		h.GetUniversity(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestEducationHandlerGetUniversityWrappedNotFound(t *testing.T) {
	stub := &educationHandlerStub{
		getFn: func(context.Context, string) (models.University, error) {
			return models.University{}, fmt.Errorf("wrap: %w", services.ErrUniversityNotFound)
		},
	}
	h, err := NewEducationHandler(stub)
	if err != nil {
		t.Fatalf("NewEducationHandler() error = %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/education/universities/rev-fr-moses-orshio-adasu-makurdi", nil)
	req.SetPathValue("university_id", "rev-fr-moses-orshio-adasu-makurdi")
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
		h.GetUniversity(w, r)
	}, rr)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.getCalls != 1 {
		t.Fatalf("unexpected get call count: %d", stub.getCalls)
	}
}
