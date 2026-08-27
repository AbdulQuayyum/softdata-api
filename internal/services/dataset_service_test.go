package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type datasetRepoStub struct {
	createFn                        func(context.Context, models.Dataset) (models.Dataset, error)
	getByIDFn                       func(context.Context, string) (models.Dataset, error)
	getByDatasetKeyFn               func(context.Context, string) (models.Dataset, error)
	listPublicFn                    func(context.Context, models.DatasetListFilter) ([]models.Dataset, error)
	countPublicFn                   func(context.Context, models.DatasetListFilter) (int64, error)
	listAllFn                       func(context.Context, int32, int32) ([]models.Dataset, error)
	updateMetadataFn                func(context.Context, models.Dataset) (models.Dataset, error)
	archiveFn                       func(context.Context, string) (models.Dataset, error)
	createSourceFn                  func(context.Context, models.DatasetSource) (models.DatasetSource, error)
	getSourceByIDFn                 func(context.Context, string) (models.DatasetSource, error)
	listSourcesFn                   func(context.Context, string) ([]models.DatasetSource, error)
	updateSourceFn                  func(context.Context, models.DatasetSource) (models.DatasetSource, error)
	deleteSourceFn                  func(context.Context, string) error
	createVersionFn                 func(context.Context, models.DatasetVersion) (models.DatasetVersion, error)
	getVersionByIDFn                func(context.Context, string) (models.DatasetVersion, error)
	getVersionByDatasetAndVersionFn func(context.Context, string, string, string) (models.DatasetVersion, error)
	listVersionsFn                  func(context.Context, string) ([]models.DatasetVersion, error)
	updateVersionFn                 func(context.Context, models.DatasetVersion) (models.DatasetVersion, error)
	publishVersionFn                func(context.Context, string) (models.DatasetVersion, error)
	deleteVersionFn                 func(context.Context, string) error
	lastGetByDatasetKey             string
	lastListPublicFilter            models.DatasetListFilter
	lastListSourcesDatasetID        string
	lastListVersionsDatasetID       string
}

func (s *datasetRepoStub) Create(ctx context.Context, dataset models.Dataset) (models.Dataset, error) {
	if s.createFn != nil {
		return s.createFn(ctx, dataset)
	}
	return models.Dataset{}, nil
}

func (s *datasetRepoStub) GetByID(ctx context.Context, id string) (models.Dataset, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return models.Dataset{}, interfaces.ErrNotFound
}

func (s *datasetRepoStub) GetByDatasetKey(ctx context.Context, datasetKey string) (models.Dataset, error) {
	s.lastGetByDatasetKey = datasetKey
	if s.getByDatasetKeyFn != nil {
		return s.getByDatasetKeyFn(ctx, datasetKey)
	}
	return models.Dataset{}, interfaces.ErrNotFound
}

func (s *datasetRepoStub) ListPublic(ctx context.Context, filter models.DatasetListFilter) ([]models.Dataset, error) {
	s.lastListPublicFilter = filter
	if s.listPublicFn != nil {
		return s.listPublicFn(ctx, filter)
	}
	return nil, nil
}

func (s *datasetRepoStub) CountPublic(ctx context.Context, filter models.DatasetListFilter) (int64, error) {
	if s.countPublicFn != nil {
		return s.countPublicFn(ctx, filter)
	}
	return 0, nil
}

func (s *datasetRepoStub) ListAll(ctx context.Context, limit, offset int32) ([]models.Dataset, error) {
	if s.listAllFn != nil {
		return s.listAllFn(ctx, limit, offset)
	}
	return nil, nil
}

func (s *datasetRepoStub) UpdateMetadata(ctx context.Context, dataset models.Dataset) (models.Dataset, error) {
	if s.updateMetadataFn != nil {
		return s.updateMetadataFn(ctx, dataset)
	}
	return models.Dataset{}, nil
}

func (s *datasetRepoStub) Archive(ctx context.Context, id string) (models.Dataset, error) {
	if s.archiveFn != nil {
		return s.archiveFn(ctx, id)
	}
	return models.Dataset{}, nil
}

func (s *datasetRepoStub) CreateSource(ctx context.Context, source models.DatasetSource) (models.DatasetSource, error) {
	if s.createSourceFn != nil {
		return s.createSourceFn(ctx, source)
	}
	return models.DatasetSource{}, nil
}

func (s *datasetRepoStub) GetSourceByID(ctx context.Context, id string) (models.DatasetSource, error) {
	if s.getSourceByIDFn != nil {
		return s.getSourceByIDFn(ctx, id)
	}
	return models.DatasetSource{}, interfaces.ErrNotFound
}

func (s *datasetRepoStub) ListSources(ctx context.Context, datasetID string) ([]models.DatasetSource, error) {
	s.lastListSourcesDatasetID = datasetID
	if s.listSourcesFn != nil {
		return s.listSourcesFn(ctx, datasetID)
	}
	return nil, nil
}

func (s *datasetRepoStub) UpdateSource(ctx context.Context, source models.DatasetSource) (models.DatasetSource, error) {
	if s.updateSourceFn != nil {
		return s.updateSourceFn(ctx, source)
	}
	return models.DatasetSource{}, nil
}

func (s *datasetRepoStub) DeleteSource(ctx context.Context, id string) error {
	if s.deleteSourceFn != nil {
		return s.deleteSourceFn(ctx, id)
	}
	return nil
}

func (s *datasetRepoStub) CreateVersion(ctx context.Context, version models.DatasetVersion) (models.DatasetVersion, error) {
	if s.createVersionFn != nil {
		return s.createVersionFn(ctx, version)
	}
	return models.DatasetVersion{}, nil
}

func (s *datasetRepoStub) GetVersionByID(ctx context.Context, id string) (models.DatasetVersion, error) {
	if s.getVersionByIDFn != nil {
		return s.getVersionByIDFn(ctx, id)
	}
	return models.DatasetVersion{}, interfaces.ErrNotFound
}

func (s *datasetRepoStub) GetVersionByDatasetAndVersion(ctx context.Context, datasetID, versionName, format string) (models.DatasetVersion, error) {
	if s.getVersionByDatasetAndVersionFn != nil {
		return s.getVersionByDatasetAndVersionFn(ctx, datasetID, versionName, format)
	}
	return models.DatasetVersion{}, interfaces.ErrNotFound
}

func (s *datasetRepoStub) ListVersions(ctx context.Context, datasetID string) ([]models.DatasetVersion, error) {
	s.lastListVersionsDatasetID = datasetID
	if s.listVersionsFn != nil {
		return s.listVersionsFn(ctx, datasetID)
	}
	return nil, nil
}

func (s *datasetRepoStub) UpdateVersion(ctx context.Context, version models.DatasetVersion) (models.DatasetVersion, error) {
	if s.updateVersionFn != nil {
		return s.updateVersionFn(ctx, version)
	}
	return models.DatasetVersion{}, nil
}

func (s *datasetRepoStub) PublishVersion(ctx context.Context, id string) (models.DatasetVersion, error) {
	if s.publishVersionFn != nil {
		return s.publishVersionFn(ctx, id)
	}
	return models.DatasetVersion{}, nil
}

func (s *datasetRepoStub) DeleteVersion(ctx context.Context, id string) error {
	if s.deleteVersionFn != nil {
		return s.deleteVersionFn(ctx, id)
	}
	return nil
}

func TestNewDatasetServiceRejectsNilRepository(t *testing.T) {
	if _, err := NewDatasetService(nil); err == nil {
		t.Fatal("NewDatasetService() error = nil, want error")
	}
}

func TestDatasetServiceListDatasets(t *testing.T) {
	repo := &datasetRepoStub{
		listPublicFn: func(_ context.Context, filter models.DatasetListFilter) ([]models.Dataset, error) {
			if filter.Search != "states" {
				t.Fatalf("unexpected search: %q", filter.Search)
			}
			if filter.Limit != 7 {
				t.Fatalf("unexpected limit: %d", filter.Limit)
			}
			if filter.Offset != 0 {
				t.Fatalf("unexpected offset: %d", filter.Offset)
			}
			now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
			return []models.Dataset{
				{
					ID:            "uuid-active",
					DatasetKey:    "ng-states",
					Name:          "Nigerian States",
					GroupName:     "geography",
					Version:       "1.0.0",
					Status:        models.DatasetStatusActive,
					RecordCount:   37,
					PrimaryFormat: "json",
					Formats:       []string{"json", "csv"},
					IsPublic:      true,
					CreatedAt:     now,
					UpdatedAt:     now,
				},
				{
					ID:            "uuid-deprecated",
					DatasetKey:    "ng-wards",
					Name:          "Nigerian Wards",
					GroupName:     "geography",
					Version:       "1.0.0",
					Status:        models.DatasetStatusDeprecated,
					RecordCount:   7749,
					PrimaryFormat: "json",
					IsPublic:      true,
					CreatedAt:     now,
					UpdatedAt:     now,
				},
			}, nil
		},
		countPublicFn: func(_ context.Context, filter models.DatasetListFilter) (int64, error) {
			if filter.Search != "states" {
				t.Fatalf("count search mismatch: %q", filter.Search)
			}
			return 2, nil
		},
	}
	svc, err := NewDatasetService(repo)
	if err != nil {
		t.Fatalf("NewDatasetService() error = %v", err)
	}

	result, err := svc.ListDatasets(context.Background(), " states ", 1, 7)
	if err != nil {
		t.Fatalf("ListDatasets() error = %v", err)
	}
	if len(result.Datasets) != 2 {
		t.Fatalf("unexpected dataset count: %d", len(result.Datasets))
	}
	if result.Datasets[0].ID != "ng-states" || result.Datasets[1].ID != "ng-wards" {
		t.Fatalf("dataset ids were not mapped from dataset keys: %#v", result.Datasets)
	}
	if result.Total != 2 || result.TotalPages != 1 || result.Page != 1 || result.Limit != 7 {
		t.Fatalf("unexpected pagination result: %#v", result)
	}
}

func TestDatasetServiceListDatasetsPaginationAndValidation(t *testing.T) {
	repo := &datasetRepoStub{
		listPublicFn: func(context.Context, models.DatasetListFilter) ([]models.Dataset, error) {
			return nil, nil
		},
		countPublicFn: func(context.Context, models.DatasetListFilter) (int64, error) {
			return 5, nil
		},
	}
	svc, err := NewDatasetService(repo)
	if err != nil {
		t.Fatalf("NewDatasetService() error = %v", err)
	}

	result, err := svc.ListDatasets(context.Background(), "", 3, 2)
	if err != nil {
		t.Fatalf("ListDatasets() error = %v", err)
	}
	if repo.lastListPublicFilter.Offset != 4 {
		t.Fatalf("unexpected offset: %d", repo.lastListPublicFilter.Offset)
	}
	if result.TotalPages != 3 {
		t.Fatalf("unexpected total pages: %d", result.TotalPages)
	}
	if result.Datasets == nil || len(result.Datasets) != 0 {
		t.Fatalf("expected non-nil empty slice, got %#v", result.Datasets)
	}

	if _, err := svc.ListDatasets(context.Background(), "", 0, 2); !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("page validation error = %v, want ErrInvalidPagination", err)
	}
	if _, err := svc.ListDatasets(context.Background(), "", 1, 0); !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("limit validation error = %v, want ErrInvalidPagination", err)
	}
	if _, err := svc.ListDatasets(context.Background(), "", 1, maxDatasetListLimit+1); !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("max limit validation error = %v, want ErrInvalidPagination", err)
	}
}

func TestDatasetServiceListDatasetsContextCancellation(t *testing.T) {
	repo := &datasetRepoStub{
		listPublicFn: func(context.Context, models.DatasetListFilter) ([]models.Dataset, error) {
			t.Fatal("ListPublic should not be called after context cancellation")
			return nil, nil
		},
		countPublicFn: func(context.Context, models.DatasetListFilter) (int64, error) {
			t.Fatal("CountPublic should not be called after context cancellation")
			return 0, nil
		},
	}
	svc, err := NewDatasetService(repo)
	if err != nil {
		t.Fatalf("NewDatasetService() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = svc.ListDatasets(ctx, "", 1, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ListDatasets() error = %v, want context.Canceled", err)
	}
}

func TestDatasetServiceListDatasetsRepositoryFailure(t *testing.T) {
	wantErr := errors.New("repository unavailable")
	repo := &datasetRepoStub{
		listPublicFn: func(context.Context, models.DatasetListFilter) ([]models.Dataset, error) {
			return nil, wantErr
		},
		countPublicFn: func(context.Context, models.DatasetListFilter) (int64, error) {
			t.Fatal("CountPublic should not be called after ListPublic failure")
			return 0, nil
		},
	}
	svc, err := NewDatasetService(repo)
	if err != nil {
		t.Fatalf("NewDatasetService() error = %v", err)
	}

	_, err = svc.ListDatasets(context.Background(), "", 1, 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ListDatasets() error = %v, want repository error", err)
	}
}

func TestDatasetServiceGetDatasetVisibilityAndMapping(t *testing.T) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name      string
		status    models.DatasetStatus
		public    bool
		wantError error
	}{
		{name: "active", status: models.DatasetStatusActive, public: true},
		{name: "deprecated", status: models.DatasetStatusDeprecated, public: true},
		{name: "draft hidden", status: models.DatasetStatusDraft, public: true, wantError: ErrDatasetNotFound},
		{name: "review hidden", status: models.DatasetStatusReview, public: true, wantError: ErrDatasetNotFound},
		{name: "archived hidden", status: models.DatasetStatusArchived, public: true, wantError: ErrDatasetNotFound},
		{name: "missing", status: "", public: false, wantError: ErrDatasetNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &datasetRepoStub{
				getByDatasetKeyFn: func(context.Context, string) (models.Dataset, error) {
					if tc.name == "missing" {
						return models.Dataset{}, interfaces.ErrNotFound
					}
					return models.Dataset{
						ID:              "uuid-dataset",
						DatasetKey:      "ng-states",
						Name:            "Nigerian States",
						GroupName:       "geography",
						Description:     strPtr("States and the FCT."),
						CountryCode:     strPtr("NG"),
						Version:         "1.0.0",
						Status:          tc.status,
						RecordCount:     37,
						PrimaryFormat:   "json",
						Formats:         []string{"json", "csv"},
						SchemaPath:      strPtr("geography.schema.json"),
						LicenceID:       strPtr("licence-example"),
						SourceCount:     1,
						UpdateFrequency: strPtr("yearly"),
						LastUpdatedAt:   timePtr(now),
						LastVerifiedAt:  timePtr(now),
						Maintainers:     []string{"A"},
						IsPublic:        tc.public,
						CreatedAt:       now,
						UpdatedAt:       now,
					}, nil
				},
			}
			svc, err := NewDatasetService(repo)
			if err != nil {
				t.Fatalf("NewDatasetService() error = %v", err)
			}

			result, err := svc.GetDataset(context.Background(), " ng-states ")
			if tc.wantError != nil {
				if !errors.Is(err, tc.wantError) {
					t.Fatalf("GetDataset() error = %v, want %v", err, tc.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetDataset() error = %v", err)
			}
			if result.ID != "ng-states" {
				t.Fatalf("unexpected public id: %q", result.ID)
			}
			marshaled, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if strings.Contains(string(marshaled), "uuid-dataset") {
				t.Fatalf("serialized dataset exposed internal uuid: %s", marshaled)
			}
		})
	}
}

func TestDatasetServiceListDatasetSources(t *testing.T) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	repo := &datasetRepoStub{
		getByDatasetKeyFn: func(context.Context, string) (models.Dataset, error) {
			return models.Dataset{
				ID:         "uuid-dataset",
				DatasetKey: "ng-states",
				Status:     models.DatasetStatusActive,
				IsPublic:   true,
			}, nil
		},
		listSourcesFn: func(_ context.Context, datasetID string) ([]models.DatasetSource, error) {
			if datasetID != "uuid-dataset" {
				t.Fatalf("unexpected dataset id passed to ListSources: %q", datasetID)
			}
			return []models.DatasetSource{
				{
					ID:             "uuid-source",
					DatasetID:      datasetID,
					SourceKey:      "source-key-1",
					Name:           "Ministry of Examples",
					URL:            strPtr("https://example.test"),
					IsOfficial:     true,
					LastFetchedAt:  timePtr(now),
					LastVerifiedAt: timePtr(now),
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			}, nil
		},
	}
	svc, err := NewDatasetService(repo)
	if err != nil {
		t.Fatalf("NewDatasetService() error = %v", err)
	}

	result, err := svc.ListDatasetSources(context.Background(), "ng-states")
	if err != nil {
		t.Fatalf("ListDatasetSources() error = %v", err)
	}
	if len(result) != 1 || result[0].ID != "source-key-1" {
		t.Fatalf("unexpected sources: %#v", result)
	}
	marshaled, err := json.Marshal(result[0])
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(marshaled), "uuid-source") || strings.Contains(string(marshaled), "dataset-id") {
		t.Fatalf("serialized source exposed internal identifiers: %s", marshaled)
	}
}

func TestDatasetServiceListDatasetSourcesEmptyAndHidden(t *testing.T) {
	repo := &datasetRepoStub{
		getByDatasetKeyFn: func(context.Context, string) (models.Dataset, error) {
			return models.Dataset{
				ID:         "uuid-dataset",
				DatasetKey: "ng-states",
				Status:     models.DatasetStatusActive,
				IsPublic:   true,
			}, nil
		},
		listSourcesFn: func(context.Context, string) ([]models.DatasetSource, error) {
			return nil, nil
		},
	}
	svc, err := NewDatasetService(repo)
	if err != nil {
		t.Fatalf("NewDatasetService() error = %v", err)
	}

	result, err := svc.ListDatasetSources(context.Background(), "ng-states")
	if err != nil {
		t.Fatalf("ListDatasetSources() error = %v", err)
	}
	if result == nil || len(result) != 0 {
		t.Fatalf("expected non-nil empty slice, got %#v", result)
	}

	repo.getByDatasetKeyFn = func(context.Context, string) (models.Dataset, error) {
		return models.Dataset{ID: "uuid-dataset", DatasetKey: "ng-states", Status: models.DatasetStatusDraft, IsPublic: true}, nil
	}
	if _, err := svc.ListDatasetSources(context.Background(), "ng-states"); !errors.Is(err, ErrDatasetNotFound) {
		t.Fatalf("hidden dataset error = %v, want ErrDatasetNotFound", err)
	}
}

func TestDatasetServiceListDatasetVersions(t *testing.T) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	repo := &datasetRepoStub{
		getByDatasetKeyFn: func(context.Context, string) (models.Dataset, error) {
			return models.Dataset{
				ID:         "uuid-dataset",
				DatasetKey: "ng-states",
				Status:     models.DatasetStatusDeprecated,
				IsPublic:   true,
			}, nil
		},
		listVersionsFn: func(_ context.Context, datasetID string) ([]models.DatasetVersion, error) {
			if datasetID != "uuid-dataset" {
				t.Fatalf("unexpected dataset id passed to ListVersions: %q", datasetID)
			}
			return []models.DatasetVersion{
				{
					ID:            "uuid-version",
					DatasetID:     datasetID,
					Version:       "1.0.0",
					SchemaVersion: strPtr("1"),
					Format:        "json",
					Status:        models.DatasetVersionStatusPublished,
					RecordCount:   37,
					ReleasedAt:    timePtr(now),
					CreatedAt:     now,
					UpdatedAt:     now,
				},
			}, nil
		},
	}
	svc, err := NewDatasetService(repo)
	if err != nil {
		t.Fatalf("NewDatasetService() error = %v", err)
	}

	result, err := svc.ListDatasetVersions(context.Background(), "ng-states")
	if err != nil {
		t.Fatalf("ListDatasetVersions() error = %v", err)
	}
	if len(result) != 1 || result[0].Version != "1.0.0" {
		t.Fatalf("unexpected versions: %#v", result)
	}
	marshaled, err := json.Marshal(result[0])
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(marshaled), "uuid-version") || strings.Contains(string(marshaled), "dataset-id") {
		t.Fatalf("serialized version exposed internal identifiers: %s", marshaled)
	}
}

func TestDatasetServiceListDatasetVersionsEmpty(t *testing.T) {
	repo := &datasetRepoStub{
		getByDatasetKeyFn: func(context.Context, string) (models.Dataset, error) {
			return models.Dataset{
				ID:         "uuid-dataset",
				DatasetKey: "ng-states",
				Status:     models.DatasetStatusActive,
				IsPublic:   true,
			}, nil
		},
		listVersionsFn: func(context.Context, string) ([]models.DatasetVersion, error) {
			return nil, nil
		},
	}
	svc, err := NewDatasetService(repo)
	if err != nil {
		t.Fatalf("NewDatasetService() error = %v", err)
	}

	result, err := svc.ListDatasetVersions(context.Background(), "ng-states")
	if err != nil {
		t.Fatalf("ListDatasetVersions() error = %v", err)
	}
	if result == nil || len(result) != 0 {
		t.Fatalf("expected non-nil empty slice, got %#v", result)
	}
}

func TestDatasetServiceListDatasetVersionsHiddenDataset(t *testing.T) {
	repo := &datasetRepoStub{
		getByDatasetKeyFn: func(context.Context, string) (models.Dataset, error) {
			return models.Dataset{ID: "uuid-dataset", DatasetKey: "ng-states", Status: models.DatasetStatusReview, IsPublic: true}, nil
		},
		listVersionsFn: func(context.Context, string) ([]models.DatasetVersion, error) {
			t.Fatal("ListVersions should not be called for hidden datasets")
			return nil, nil
		},
	}
	svc, err := NewDatasetService(repo)
	if err != nil {
		t.Fatalf("NewDatasetService() error = %v", err)
	}

	_, err = svc.ListDatasetVersions(context.Background(), "ng-states")
	if !errors.Is(err, ErrDatasetNotFound) {
		t.Fatalf("ListDatasetVersions() error = %v, want ErrDatasetNotFound", err)
	}
}

func strPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	cloned := value
	return &cloned
}
