package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const maxDatasetListLimit = 100

// DatasetListResult is the service-level list result for datasets.
type DatasetListResult struct {
	Datasets   []models.DatasetResponse
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

type DatasetService struct {
	repo interfaces.DatasetRepository
}

func NewDatasetService(repo interfaces.DatasetRepository) (*DatasetService, error) {
	if repo == nil {
		return nil, fmt.Errorf("dataset repository is required")
	}
	return &DatasetService{repo: repo}, nil
}

func (s *DatasetService) ListDatasets(ctx context.Context, search string, page, limit int) (DatasetListResult, error) {
	if err := ctx.Err(); err != nil {
		return DatasetListResult{}, err
	}

	search = strings.TrimSpace(search)
	if page < 1 || limit < 1 || limit > maxDatasetListLimit {
		return DatasetListResult{}, ErrInvalidPagination
	}

	offset, err := pageLimitToOffset(page, limit)
	if err != nil {
		return DatasetListResult{}, err
	}

	filter := models.DatasetListFilter{
		Search: search,
		Limit:  int32(limit),
		Offset: offset,
	}

	rows, err := s.repo.ListPublic(ctx, filter)
	if err != nil {
		return DatasetListResult{}, translateDatasetServiceError("list datasets", err)
	}

	total, err := s.repo.CountPublic(ctx, filter)
	if err != nil {
		return DatasetListResult{}, translateDatasetServiceError("count datasets", err)
	}

	items := make([]models.DatasetResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDatasetResponse(row))
	}

	return DatasetListResult{
		Datasets:   items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages(total, limit),
	}, nil
}

func (s *DatasetService) GetDataset(ctx context.Context, datasetKey string) (models.DatasetResponse, error) {
	if err := ctx.Err(); err != nil {
		return models.DatasetResponse{}, err
	}

	dataset, err := s.resolveVisibleDataset(ctx, datasetKey)
	if err != nil {
		return models.DatasetResponse{}, err
	}

	return toDatasetResponse(dataset), nil
}

func (s *DatasetService) ListDatasetSources(ctx context.Context, datasetKey string) ([]models.DatasetSourceResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	dataset, err := s.resolveVisibleDataset(ctx, datasetKey)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.ListSources(ctx, dataset.ID)
	if err != nil {
		return nil, translateDatasetServiceError("list dataset sources", err)
	}

	items := make([]models.DatasetSourceResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDatasetSourceResponse(row))
	}
	return items, nil
}

func (s *DatasetService) ListDatasetVersions(ctx context.Context, datasetKey string) ([]models.DatasetVersionResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	dataset, err := s.resolveVisibleDataset(ctx, datasetKey)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.ListVersions(ctx, dataset.ID)
	if err != nil {
		return nil, translateDatasetServiceError("list dataset versions", err)
	}

	items := make([]models.DatasetVersionResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDatasetVersionResponse(row))
	}
	return items, nil
}

func (s *DatasetService) resolveVisibleDataset(ctx context.Context, datasetKey string) (models.Dataset, error) {
	datasetKey = strings.TrimSpace(datasetKey)
	if datasetKey == "" {
		return models.Dataset{}, ErrInvalidDatasetKey
	}

	dataset, err := s.repo.GetByDatasetKey(ctx, datasetKey)
	if err != nil {
		return models.Dataset{}, translateDatasetServiceError("get dataset", err)
	}
	if !datasetVisible(dataset) {
		return models.Dataset{}, ErrDatasetNotFound
	}
	return dataset, nil
}

func pageLimitToOffset(page, limit int) (int32, error) {
	offset := int64(page-1) * int64(limit)
	if offset < 0 || offset > math.MaxInt32 {
		return 0, ErrInvalidPagination
	}
	return int32(offset), nil
}

func totalPages(total int64, limit int) int {
	if total == 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}

func datasetVisible(dataset models.Dataset) bool {
	switch dataset.Status {
	case models.DatasetStatusActive, models.DatasetStatusDeprecated:
		return dataset.IsPublic
	default:
		return false
	}
}

func translateDatasetServiceError(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, interfaces.ErrNotFound) {
		switch op {
		case "get dataset", "list dataset sources", "list dataset versions":
			return ErrDatasetNotFound
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}

func toDatasetResponse(dataset models.Dataset) models.DatasetResponse {
	return models.DatasetResponse{
		ID:              dataset.DatasetKey,
		Name:            dataset.Name,
		Description:     cloneStringPointer(dataset.Description),
		Group:           dataset.GroupName,
		CountryCode:     cloneStringPointer(dataset.CountryCode),
		Version:         dataset.Version,
		Status:          dataset.Status,
		RecordCount:     dataset.RecordCount,
		PrimaryFormat:   dataset.PrimaryFormat,
		Formats:         cloneStrings(dataset.Formats),
		Schema:          cloneStringPointer(dataset.SchemaPath),
		SourceIDs:       nil,
		LicenceID:       cloneStringPointer(dataset.LicenceID),
		SourceCount:     dataset.SourceCount,
		UpdateFrequency: cloneStringPointer(dataset.UpdateFrequency),
		LastUpdatedAt:   timePtrToDateString(dataset.LastUpdatedAt),
		LastVerifiedAt:  timePtrToDateString(dataset.LastVerifiedAt),
		Maintainers:     cloneStrings(dataset.Maintainers),
		IsPublic:        dataset.IsPublic,
		CreatedAt:       dataset.CreatedAt,
		UpdatedAt:       dataset.UpdatedAt,
		ArchivedAt:      dataset.ArchivedAt,
	}
}

func toDatasetSourceResponse(source models.DatasetSource) models.DatasetSourceResponse {
	return models.DatasetSourceResponse{
		ID:             source.SourceKey,
		Name:           source.Name,
		URL:            cloneStringPointer(source.URL),
		Description:    cloneStringPointer(source.Description),
		Publisher:      cloneStringPointer(source.Publisher),
		SourceType:     cloneStringPointer(source.SourceType),
		LicenceID:      cloneStringPointer(source.LicenceID),
		IsOfficial:     source.IsOfficial,
		LastFetchedAt:  cloneTimePtr(source.LastFetchedAt),
		LastVerifiedAt: cloneTimePtr(source.LastVerifiedAt),
		CreatedAt:      source.CreatedAt,
		UpdatedAt:      source.UpdatedAt,
	}
}

func toDatasetVersionResponse(version models.DatasetVersion) models.DatasetVersionResponse {
	return models.DatasetVersionResponse{
		Version:       version.Version,
		SchemaVersion: cloneStringPointer(version.SchemaVersion),
		Format:        version.Format,
		Status:        version.Status,
		RecordCount:   version.RecordCount,
		Checksum:      cloneStringPointer(version.Checksum),
		Notes:         cloneStringPointer(version.Notes),
		ReleasedAt:    cloneTimePtr(version.ReleasedAt),
		CreatedAt:     version.CreatedAt,
		UpdatedAt:     version.UpdatedAt,
	}
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func timePtrToDateString(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format("2006-01-02")
	return &formatted
}
