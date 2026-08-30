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

func TestEducationHandlerListCollegesOfEducation(t *testing.T) {
	t.Run("list all", func(t *testing.T) {
		stub := &educationHandlerStub{
			collegeListFn: func(context.Context, services.CollegeOfEducationListInput) ([]models.CollegeOfEducation, error) {
				return []models.CollegeOfEducation{{
					ID:            "federal-college-of-education-zaria",
					Name:          "Federal College of Education, Zaria",
					OwnershipType: "federal",
					StateID:       "kaduna",
					CountryCode:   "NG",
				}}, nil
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListCollegesOfEducation(w, r)
		}, rr)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.collegeListCalls != 1 {
			t.Fatalf("unexpected list call count: %d", stub.collegeListCalls)
		}
		if stub.lastCollegeInput != (services.CollegeOfEducationListInput{}) {
			t.Fatalf("unexpected input: %#v", stub.lastCollegeInput)
		}

		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].([]any)
		if len(data) != 1 {
			t.Fatalf("unexpected list payload: %#v", data)
		}
		item := data[0].(map[string]any)
		if len(item) != 5 || item["id"] != "federal-college-of-education-zaria" || item["state_id"] != "kaduna" {
			t.Fatalf("unexpected college payload: %#v", item)
		}
	})

	t.Run("normalize nil slice", func(t *testing.T) {
		stub := &educationHandlerStub{
			collegeListFn: func(context.Context, services.CollegeOfEducationListInput) ([]models.CollegeOfEducation, error) {
				return nil, nil
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListCollegesOfEducation(w, r)
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
			collegeListFn: func(ctx context.Context, input services.CollegeOfEducationListInput) ([]models.CollegeOfEducation, error) {
				if input.OwnershipType != "private" || input.StateID != "lagos" {
					t.Fatalf("unexpected input: %#v", input)
				}
				return []models.CollegeOfEducation{{
					ID:            "micheal-otu-education-college",
					Name:          "Micheal Otu College of Education, Mbo",
					OwnershipType: "private",
					StateID:       "cross-river",
					CountryCode:   "NG",
				}}, nil
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education?ownership_type=private&state_id=lagos", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListCollegesOfEducation(w, r)
		}, rr)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.collegeListCalls != 1 {
			t.Fatalf("unexpected list call count: %d", stub.collegeListCalls)
		}
	})

	t.Run("reject invalid query", func(t *testing.T) {
		stub := &educationHandlerStub{}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education?ownership_type=Federal", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListCollegesOfEducation(w, r)
		}, rr)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.collegeListCalls != 0 {
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
		req := httptest.NewRequest(http.MethodPost, "/v1/education/colleges-of-education", nil)
		h.ListCollegesOfEducation(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestEducationHandlerGetCollegeOfEducation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		stub := &educationHandlerStub{
			collegeGetFn: func(ctx context.Context, collegeID string) (models.CollegeOfEducation, error) {
				if collegeID != "federal-college-of-education-zaria" {
					t.Fatalf("unexpected college id: %q", collegeID)
				}
				return models.CollegeOfEducation{
					ID:            collegeID,
					Name:          "Federal College of Education, Zaria",
					OwnershipType: "federal",
					StateID:       "kaduna",
					CountryCode:   "NG",
				}, nil
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education/federal-college-of-education-zaria", nil)
		req.SetPathValue("college_id", " federal-college-of-education-zaria ")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCollegeOfEducation(w, r)
		}, rr)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.collegeGetCalls != 1 {
			t.Fatalf("unexpected get call count: %d", stub.collegeGetCalls)
		}

		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		item := body["data"].(map[string]any)
		if len(item) != 5 || item["name"] != "Federal College of Education, Zaria" || item["country_code"] != "NG" {
			t.Fatalf("unexpected college response: %#v", item)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		stub := &educationHandlerStub{}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education/College-Of-Education", nil)
		req.SetPathValue("college_id", "College-Of-Education")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCollegeOfEducation(w, r)
		}, rr)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.collegeGetCalls != 0 {
			t.Fatalf("service should not be called for invalid id")
		}
	})

	t.Run("not found", func(t *testing.T) {
		stub := &educationHandlerStub{
			collegeGetFn: func(context.Context, string) (models.CollegeOfEducation, error) {
				return models.CollegeOfEducation{}, services.ErrCollegeOfEducationNotFound
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education/federal-college-of-education-zaria", nil)
		req.SetPathValue("college_id", "federal-college-of-education-zaria")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCollegeOfEducation(w, r)
		}, rr)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.collegeGetCalls != 1 {
			t.Fatalf("unexpected get call count: %d", stub.collegeGetCalls)
		}
		if !strings.Contains(rr.Body.String(), "req_test") {
			t.Fatalf("request id missing from error response: %s", rr.Body.String())
		}
	})

	t.Run("wrapped not found", func(t *testing.T) {
		stub := &educationHandlerStub{
			collegeGetFn: func(context.Context, string) (models.CollegeOfEducation, error) {
				return models.CollegeOfEducation{}, fmt.Errorf("wrap: %w", services.ErrCollegeOfEducationNotFound)
			},
		}
		h, err := NewEducationHandler(stub)
		if err != nil {
			t.Fatalf("NewEducationHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education/federal-college-of-education-zaria", nil)
		req.SetPathValue("college_id", "federal-college-of-education-zaria")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCollegeOfEducation(w, r)
		}, rr)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})
}
