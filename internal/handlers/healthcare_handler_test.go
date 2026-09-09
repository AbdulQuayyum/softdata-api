package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type healthFacilityHandlerStub struct {
	listResult interfaces.HealthFacilityListResult
	listErr    error
	getResult  models.HealthFacility
	getErr     error
	lastQuery  interfaces.HealthFacilityQuery
	lastID     string
}

func (s *healthFacilityHandlerStub) ListHealthFacilities(_ context.Context, query interfaces.HealthFacilityQuery) (interfaces.HealthFacilityListResult, error) {
	s.lastQuery = query
	return s.listResult, s.listErr
}

func (s *healthFacilityHandlerStub) GetHealthFacility(_ context.Context, id string) (models.HealthFacility, error) {
	s.lastID = id
	return s.getResult, s.getErr
}

func TestHealthFacilityHandlerListContract(t *testing.T) {
	latitude, longitude := 6.6018, 3.3515
	stub := &healthFacilityHandlerStub{listResult: interfaces.HealthFacilityListResult{
		Facilities: []models.HealthFacility{{ID: "example-health-facility", Name: "Example Health Facility", FacilityType: "primary-health-centre", FacilityLevel: "primary", OwnershipType: "private", StateID: "lagos", LGAID: "lagos-ikeja", CountryCode: "NG", SourceFacilityID: "source-1", Latitude: &latitude, Longitude: &longitude}},
		Page:       1, PageSize: 50, Total: 50649, TotalPages: 1013,
	}}
	handler, err := NewHealthFacilityHandler(stub)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/healthcare/health-facilities?page=1&page_size=50&state_id=%20lagos%20&search=%20Example%20", nil)
	response := httptest.NewRecorder()
	handler.ListHealthFacilities(response, req)
	if response.Code != http.StatusOK || !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("status/content type = %d/%q", response.Code, response.Header().Get("Content-Type"))
	}
	var body struct {
		Success bool                    `json:"success"`
		Data    []models.HealthFacility `json:"data"`
		Meta    struct {
			Page       int `json:"page"`
			Limit      int `json:"limit"`
			Total      int `json:"total"`
			TotalPages int `json:"total_pages"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || len(body.Data) != 1 || body.Meta.Page != 1 || body.Meta.Limit != 50 || body.Meta.Total != 50649 || body.Meta.TotalPages != 1013 {
		t.Fatalf("unexpected response: %#v", body)
	}
	if stub.lastQuery.StateID != "lagos" || stub.lastQuery.Search != "Example" {
		t.Fatalf("query was not normalized: %#v", stub.lastQuery)
	}
	if !strings.Contains(response.Body.String(), `"latitude":6.6018`) || !strings.Contains(response.Body.String(), `"longitude":3.3515`) {
		t.Fatalf("coordinates missing: %s", response.Body.String())
	}
}

func TestHealthFacilityHandlerListValidationAndEmptyPage(t *testing.T) {
	stub := &healthFacilityHandlerStub{listResult: interfaces.HealthFacilityListResult{Facilities: []models.HealthFacility{}, Page: 9999, PageSize: 100, Total: 50649, TotalPages: 507}}
	handler, _ := NewHealthFacilityHandler(stub)
	for _, raw := range []string{"page_size=101", "page=0", "page=-1", "page=1.5", "page=+1", "page=999999999999999999999", "facility_type=hospital", "facility_level=quaternary", "ownership_type=public", "search=" + strings.Repeat("x", 101)} {
		req := httptest.NewRequest(http.MethodGet, "/v1/healthcare/health-facilities?"+raw, nil)
		response := httptest.NewRecorder()
		handler.ListHealthFacilities(response, req)
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "INVALID_REQUEST") && !strings.Contains(response.Body.String(), "VALIDATION_FAILED") {
			t.Fatalf("%s: status/body = %d/%s", raw, response.Code, response.Body.String())
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/healthcare/health-facilities?page=9999&page_size=100", nil)
	response := httptest.NewRecorder()
	handler.ListHealthFacilities(response, req)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"data":[]`) {
		t.Fatalf("empty page response = %d/%s", response.Code, response.Body.String())
	}
}

func TestHealthFacilityHandlerAllFilterEnums(t *testing.T) {
	stub := &healthFacilityHandlerStub{listResult: interfaces.HealthFacilityListResult{Facilities: []models.HealthFacility{}, Page: 1, PageSize: 50}}
	handler, _ := NewHealthFacilityHandler(stub)
	for _, value := range []string{"primary-health-centre", "clinic", "health-post", "general-hospital", "other", "teaching-hospital", "specialist-hospital"} {
		recordHealthFilterRequest(t, handler, "facility_type", value)
	}
	for _, value := range []string{"primary", "secondary", "tertiary"} {
		recordHealthFilterRequest(t, handler, "facility_level", value)
	}
	for _, value := range []string{"federal", "state", "local-government", "private", "military", "other-public"} {
		recordHealthFilterRequest(t, handler, "ownership_type", value)
	}
}

func recordHealthFilterRequest(t *testing.T, handler *HealthFacilityHandler, field, value string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/v1/healthcare/health-facilities?"+field+"="+value, nil)
	response := httptest.NewRecorder()
	handler.ListHealthFacilities(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("%s=%s returned %d: %s", field, value, response.Code, response.Body.String())
	}
}

func TestHealthFacilityHandlerDetailAndErrors(t *testing.T) {
	stub := &healthFacilityHandlerStub{getResult: models.HealthFacility{ID: "example-health-facility", Name: "Example", CountryCode: "NG"}}
	handler, _ := NewHealthFacilityHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/v1/healthcare/health-facilities/example-health-facility", nil)
	req.SetPathValue("facility_id", " example-health-facility ")
	response := httptest.NewRecorder()
	handler.GetHealthFacility(response, req)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"id":"example-health-facility"`) || stub.lastID != "example-health-facility" {
		t.Fatalf("detail response = %d/%s", response.Code, response.Body.String())
	}
	for _, id := range []string{"", "Upper-case", "bad/id", "bad--id", "-leading", "trailing-"} {
		req := httptest.NewRequest(http.MethodGet, "/v1/healthcare/health-facilities/"+id, nil)
		req.SetPathValue("facility_id", id)
		response := httptest.NewRecorder()
		handler.GetHealthFacility(response, req)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("invalid id %q returned %d", id, response.Code)
		}
	}
	stub.getErr = services.ErrHealthFacilityNotFound
	req = httptest.NewRequest(http.MethodGet, "/v1/healthcare/health-facilities/missing-facility", nil)
	req.SetPathValue("facility_id", "missing-facility")
	response = httptest.NewRecorder()
	handler.GetHealthFacility(response, req)
	if response.Code != http.StatusNotFound {
		t.Fatalf("not found returned %d: %s", response.Code, response.Body.String())
	}
	stub.getErr = errors.New("/private/path health internals")
	response = httptest.NewRecorder()
	handler.GetHealthFacility(response, req)
	if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), "/private/path") {
		t.Fatalf("unexpected error leaked: %d/%s", response.Code, response.Body.String())
	}
}

func TestHealthFacilityHandlerContextErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
	}{{"canceled", context.Canceled}, {"deadline", context.DeadlineExceeded}} {
		t.Run(test.name, func(t *testing.T) {
			stub := &healthFacilityHandlerStub{listErr: test.err}
			handler, _ := NewHealthFacilityHandler(stub)
			req := httptest.NewRequest(http.MethodGet, "/v1/healthcare/health-facilities", nil)
			response := httptest.NewRecorder()
			handler.ListHealthFacilities(response, req)
			if response.Code != http.StatusServiceUnavailable {
				t.Fatalf("context error status = %d", response.Code)
			}
		})
	}
}

func TestHealthFacilityHandlerMapsStateLGARelationshipError(t *testing.T) {
	stub := &healthFacilityHandlerStub{
		listErr: services.ErrInvalidHealthFacilityStateLGA,
	}
	handler, _ := NewHealthFacilityHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/v1/healthcare/health-facilities?state_id=lagos&lga_id=lagos-ikeja", nil)
	response := httptest.NewRecorder()
	handler.ListHealthFacilities(response, req)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("state/LGA relationship returned %d: %s", response.Code, response.Body.String())
	}
}

func TestHealthFacilityHandlerPublishedAndBoundedIDs(t *testing.T) {
	jsonRepository, err := fileRepo.NewJSONRepository("../../datasets", 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := fileRepo.NewHealthFacilityRepository(jsonRepository, "healthcare/health_facilities.json", "geography/states.json", "geography/lgas.json")
	if err != nil {
		t.Fatal(err)
	}
	service, err := services.NewHealthFacilityService(repository)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHealthFacilityHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	var facilities []models.HealthFacility
	if err := jsonRepository.Decode(context.Background(), "healthcare/health_facilities.json", &facilities); err != nil {
		t.Fatal(err)
	}
	longest := ""
	for _, facility := range facilities {
		if len(facility.ID) > len(longest) {
			longest = facility.ID
		}
	}
	if len(longest) <= 128 || len(longest) > models.HealthFacilityIDMaxLength {
		t.Fatalf("unexpected longest ID: %d", len(longest))
	}
	cases := []struct {
		id     string
		status int
	}{
		{longest, http.StatusOK},
		{strings.Repeat("a", models.HealthFacilityIDMaxLength), http.StatusNotFound},
		{strings.Repeat("a", models.HealthFacilityIDMaxLength+1), http.StatusBadRequest},
		{strings.Repeat("a", 129), http.StatusNotFound},
		{"invalid_id", http.StatusBadRequest},
	}
	for _, tc := range cases {
		request := httptest.NewRequest(http.MethodGet, "/v1/healthcare/health-facilities/"+tc.id, nil)
		request.SetPathValue("facility_id", tc.id)
		response := httptest.NewRecorder()
		handler.GetHealthFacility(response, request)
		if response.Code != tc.status || !strings.Contains(response.Header().Get("Content-Type"), "application/json") || !json.Valid(response.Body.Bytes()) {
			t.Fatalf("ID length %d: %d %s", len(tc.id), response.Code, response.Body.String())
		}
	}
}
