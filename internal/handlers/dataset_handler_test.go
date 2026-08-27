package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type datasetHandlerStub struct {
	listFn     func(context.Context, string, int, int) (services.DatasetListResult, error)
	getFn      func(context.Context, string) (models.DatasetResponse, error)
	sourcesFn  func(context.Context, string) ([]models.DatasetSourceResponse, error)
	versionsFn func(context.Context, string) ([]models.DatasetVersionResponse, error)

	listCalls     int
	getCalls      int
	sourcesCalls  int
	versionsCalls int

	lastSearch string
	lastPage   int
	lastLimit  int
	lastKey    string
}

func (s *datasetHandlerStub) ListDatasets(ctx context.Context, search string, page, limit int) (services.DatasetListResult, error) {
	s.listCalls++
	s.lastSearch = search
	s.lastPage = page
	s.lastLimit = limit
	if s.listFn != nil {
		return s.listFn(ctx, search, page, limit)
	}
	return services.DatasetListResult{}, nil
}

func (s *datasetHandlerStub) GetDataset(ctx context.Context, datasetKey string) (models.DatasetResponse, error) {
	s.getCalls++
	s.lastKey = datasetKey
	if s.getFn != nil {
		return s.getFn(ctx, datasetKey)
	}
	return models.DatasetResponse{}, nil
}

func (s *datasetHandlerStub) ListDatasetSources(ctx context.Context, datasetKey string) ([]models.DatasetSourceResponse, error) {
	s.sourcesCalls++
	s.lastKey = datasetKey
	if s.sourcesFn != nil {
		return s.sourcesFn(ctx, datasetKey)
	}
	return nil, nil
}

func (s *datasetHandlerStub) ListDatasetVersions(ctx context.Context, datasetKey string) ([]models.DatasetVersionResponse, error) {
	s.versionsCalls++
	s.lastKey = datasetKey
	if s.versionsFn != nil {
		return s.versionsFn(ctx, datasetKey)
	}
	return nil, nil
}

func TestNewDatasetHandlerRejectsNilService(t *testing.T) {
	if _, err := NewDatasetHandler(nil); err == nil {
		t.Fatal("NewDatasetHandler(nil) error = nil, want error")
	}
}

func TestDatasetHandlerListDatasetsReturnsPaginatedResponse(t *testing.T) {
	stub := &datasetHandlerStub{
		listFn: func(ctx context.Context, search string, page, limit int) (services.DatasetListResult, error) {
			if search != "states" || page != 2 || limit != 5 {
				t.Fatalf("unexpected query args: %q %d %d", search, page, limit)
			}
			now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
			return services.DatasetListResult{
				Datasets: []models.DatasetResponse{{
					ID:            "ng-states",
					Name:          "Nigerian States",
					Group:         "geography",
					Version:       "1.0.0",
					Status:        models.DatasetStatusActive,
					RecordCount:   37,
					PrimaryFormat: "json",
					Formats:       []string{"json", "csv"},
					IsPublic:      true,
					CreatedAt:     now,
					UpdatedAt:     now,
				}},
				Total:      11,
				Page:       2,
				Limit:      5,
				TotalPages: 3,
			}, nil
		},
	}
	h, err := NewDatasetHandler(stub)
	if err != nil {
		t.Fatalf("NewDatasetHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	values := url.Values{}
	values.Set("search", " states ")
	values.Set("page", "2")
	values.Set("limit", "5")
	req := httptest.NewRequest(http.MethodGet, "/v1/datasets?"+values.Encode(), nil)

	h.ListDatasets(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.listCalls != 1 {
		t.Fatalf("unexpected list call count: %d", stub.listCalls)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %q", got)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("unexpected success flag: %#v", body["success"])
	}
	meta := body["meta"].(map[string]any)
	if meta["page"] != float64(2) || meta["limit"] != float64(5) || meta["total"] != float64(11) || meta["total_pages"] != float64(3) {
		t.Fatalf("unexpected pagination meta: %#v", meta)
	}
	data := body["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("unexpected data length: %#v", data)
	}
	item := data[0].(map[string]any)
	if item["id"] != "ng-states" || item["name"] != "Nigerian States" {
		t.Fatalf("unexpected dataset payload: %#v", item)
	}
}

func TestDatasetHandlerGetDatasetAndListsSourcesAndVersions(t *testing.T) {
	now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
	stub := &datasetHandlerStub{
		getFn: func(ctx context.Context, datasetKey string) (models.DatasetResponse, error) {
			if datasetKey != "ng-states" {
				t.Fatalf("unexpected dataset key: %q", datasetKey)
			}
			return models.DatasetResponse{
				ID:            "ng-states",
				Name:          "Nigerian States",
				Group:         "geography",
				Version:       "1.0.0",
				Status:        models.DatasetStatusActive,
				RecordCount:   37,
				PrimaryFormat: "json",
				IsPublic:      true,
				CreatedAt:     now,
				UpdatedAt:     now,
			}, nil
		},
		sourcesFn: func(ctx context.Context, datasetKey string) ([]models.DatasetSourceResponse, error) {
			if datasetKey != "ng-states" {
				t.Fatalf("unexpected dataset key: %q", datasetKey)
			}
			return []models.DatasetSourceResponse{{
				ID:         "source-example",
				Name:       "Official source",
				IsOfficial: true,
				CreatedAt:  now,
				UpdatedAt:  now,
			}}, nil
		},
		versionsFn: func(ctx context.Context, datasetKey string) ([]models.DatasetVersionResponse, error) {
			if datasetKey != "ng-states" {
				t.Fatalf("unexpected dataset key: %q", datasetKey)
			}
			return []models.DatasetVersionResponse{{
				Version:   "1.0.0",
				Format:    "json",
				Status:    models.DatasetVersionStatusPublished,
				CreatedAt: now,
				UpdatedAt: now,
			}}, nil
		},
	}
	h, err := NewDatasetHandler(stub)
	if err != nil {
		t.Fatalf("NewDatasetHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/datasets/ng-states", nil)
	req.SetPathValue("dataset_id", " ng-states ")
	h.GetDataset(rr, req)

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
	data := body["data"].(map[string]any)
	if data["id"] != "ng-states" || data["name"] != "Nigerian States" {
		t.Fatalf("unexpected dataset response: %#v", data)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/datasets/ng-states/sources", nil)
	req.SetPathValue("dataset_id", "ng-states")
	h.ListDatasetSources(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.sourcesCalls != 1 {
		t.Fatalf("unexpected source call count: %d", stub.sourcesCalls)
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	dataSlice := body["data"].([]any)
	if len(dataSlice) != 1 {
		t.Fatalf("unexpected sources payload: %#v", dataSlice)
	}
	source := dataSlice[0].(map[string]any)
	if source["id"] != "source-example" {
		t.Fatalf("unexpected source payload: %#v", source)
	}
	if _, exists := source["dataset_id"]; exists {
		t.Fatal("dataset_id leaked in source response")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/datasets/ng-states/versions", nil)
	req.SetPathValue("dataset_id", "ng-states")
	h.ListDatasetVersions(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.versionsCalls != 1 {
		t.Fatalf("unexpected version call count: %d", stub.versionsCalls)
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	dataSlice = body["data"].([]any)
	if len(dataSlice) != 1 {
		t.Fatalf("unexpected versions payload: %#v", dataSlice)
	}
	version := dataSlice[0].(map[string]any)
	if version["version"] != "1.0.0" {
		t.Fatalf("unexpected version payload: %#v", version)
	}
	if _, exists := version["dataset_id"]; exists {
		t.Fatal("dataset_id leaked in version response")
	}
}

func TestDatasetHandlerRejectsInvalidKeyValidationAndNonGet(t *testing.T) {
	stub := &datasetHandlerStub{}
	h, err := NewDatasetHandler(stub)
	if err != nil {
		t.Fatalf("NewDatasetHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/datasets/blank", nil)
	req.SetPathValue("dataset_id", "   ")
	h.GetDataset(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.getCalls != 0 {
		t.Fatalf("service should not be called for invalid key")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/datasets", nil)
	h.ListDatasets(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected allow header: %q", got)
	}
}

func TestDatasetHandlerRejectsDuplicateQueryValues(t *testing.T) {
	stub := &datasetHandlerStub{}
	h, err := NewDatasetHandler(stub)
	if err != nil {
		t.Fatalf("NewDatasetHandler() error = %v", err)
	}

	values := url.Values{}
	values.Add("search", "states")
	values.Add("search", "cities")
	values.Add("page", "1")
	values.Add("limit", "20")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/datasets?"+values.Encode(), nil)
	h.ListDatasets(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.listCalls != 0 {
		t.Fatalf("service should not be called for duplicate query values")
	}
}
