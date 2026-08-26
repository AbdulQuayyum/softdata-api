package postgres

import (
	"context"
	"fmt"

	sqlc "github.com/AbdulQuayyum/softdata-api/internal/database/sqlc"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var _ interfaces.DatasetRepository = (*DatasetRepository)(nil)

type DatasetRepository struct {
	queries *sqlc.Queries
}

func NewDatasetRepository(dbt sqlc.DBTX) *DatasetRepository {
	return &DatasetRepository{queries: sqlc.New(dbt)}
}

func datasetCreateParams(dataset models.Dataset) (sqlc.CreateDatasetParams, error) {
	recordCount, err := int64ToInt32(dataset.RecordCount, "record_count")
	if err != nil {
		return sqlc.CreateDatasetParams{}, err
	}
	sourceCount, err := int64ToInt32(dataset.SourceCount, "source_count")
	if err != nil {
		return sqlc.CreateDatasetParams{}, err
	}

	var status any
	if dataset.Status != "" {
		status = string(dataset.Status)
	}

	var primaryFormat any
	if dataset.PrimaryFormat != "" {
		primaryFormat = dataset.PrimaryFormat
	}

	var formats any
	if dataset.Formats != nil {
		formats = cloneStrings(dataset.Formats)
	}

	var maintainers any
	if dataset.Maintainers != nil {
		maintainers = cloneStrings(dataset.Maintainers)
	}

	return sqlc.CreateDatasetParams{
		DatasetKey:      dataset.DatasetKey,
		Slug:            dataset.Slug,
		Name:            dataset.Name,
		Description:     textFromStringPtr(dataset.Description),
		GroupName:       dataset.GroupName,
		CountryCode:     textFromStringPtr(dataset.CountryCode),
		Version:         dataset.Version,
		Column8:         status,
		Column9:         recordCount,
		Column10:        primaryFormat,
		Column11:        formats,
		SchemaPath:      textFromStringPtr(dataset.SchemaPath),
		LicenceID:       textFromStringPtr(dataset.LicenceID),
		Column14:        sourceCount,
		UpdateFrequency: textFromStringPtr(dataset.UpdateFrequency),
		LastUpdatedAt:   dateFromTimePtr(dataset.LastUpdatedAt),
		LastVerifiedAt:  dateFromTimePtr(dataset.LastVerifiedAt),
		Column18:        maintainers,
		Column19:        dataset.IsPublic,
		ArchivedAt:      timestamptzFromTimePtr(dataset.ArchivedAt),
	}, nil
}

func datasetUpdateParams(dataset models.Dataset) (sqlc.UpdateDatasetMetadataParams, error) {
	id, err := uuidFromString(dataset.ID)
	if err != nil {
		return sqlc.UpdateDatasetMetadataParams{}, err
	}
	recordCount, err := int64ToInt32(dataset.RecordCount, "record_count")
	if err != nil {
		return sqlc.UpdateDatasetMetadataParams{}, err
	}
	sourceCount, err := int64ToInt32(dataset.SourceCount, "source_count")
	if err != nil {
		return sqlc.UpdateDatasetMetadataParams{}, err
	}

	var formats []string
	if dataset.Formats != nil {
		formats = cloneStrings(dataset.Formats)
	}

	var maintainers []string
	if dataset.Maintainers != nil {
		maintainers = cloneStrings(dataset.Maintainers)
	}

	return sqlc.UpdateDatasetMetadataParams{
		ID:              id,
		Name:            dataset.Name,
		Description:     textFromStringPtr(dataset.Description),
		GroupName:       dataset.GroupName,
		CountryCode:     textFromStringPtr(dataset.CountryCode),
		Version:         dataset.Version,
		Status:          string(dataset.Status),
		RecordCount:     recordCount,
		PrimaryFormat:   dataset.PrimaryFormat,
		Formats:         formats,
		SchemaPath:      textFromStringPtr(dataset.SchemaPath),
		LicenceID:       textFromStringPtr(dataset.LicenceID),
		SourceCount:     sourceCount,
		UpdateFrequency: textFromStringPtr(dataset.UpdateFrequency),
		LastUpdatedAt:   dateFromTimePtr(dataset.LastUpdatedAt),
		LastVerifiedAt:  dateFromTimePtr(dataset.LastVerifiedAt),
		Maintainers:     maintainers,
		IsPublic:        dataset.IsPublic,
		ArchivedAt:      timestamptzFromTimePtr(dataset.ArchivedAt),
	}, nil
}

func datasetSourceCreateParams(source models.DatasetSource) (sqlc.CreateDatasetSourceParams, error) {
	datasetID, err := uuidFromString(source.DatasetID)
	if err != nil {
		return sqlc.CreateDatasetSourceParams{}, err
	}

	return sqlc.CreateDatasetSourceParams{
		DatasetID:      datasetID,
		SourceKey:      source.SourceKey,
		Name:           source.Name,
		Url:            textFromStringPtr(source.URL),
		Description:    textFromStringPtr(source.Description),
		Publisher:      textFromStringPtr(source.Publisher),
		SourceType:     textFromStringPtr(source.SourceType),
		LicenceID:      textFromStringPtr(source.LicenceID),
		Column9:        source.IsOfficial,
		LastFetchedAt:  timestamptzFromTimePtr(source.LastFetchedAt),
		LastVerifiedAt: timestamptzFromTimePtr(source.LastVerifiedAt),
	}, nil
}

func datasetSourceUpdateParams(source models.DatasetSource) (sqlc.UpdateDatasetSourceParams, error) {
	id, err := uuidFromString(source.ID)
	if err != nil {
		return sqlc.UpdateDatasetSourceParams{}, err
	}

	return sqlc.UpdateDatasetSourceParams{
		ID:             id,
		Name:           source.Name,
		Url:            textFromStringPtr(source.URL),
		Description:    textFromStringPtr(source.Description),
		Publisher:      textFromStringPtr(source.Publisher),
		SourceType:     textFromStringPtr(source.SourceType),
		LicenceID:      textFromStringPtr(source.LicenceID),
		IsOfficial:     source.IsOfficial,
		LastFetchedAt:  timestamptzFromTimePtr(source.LastFetchedAt),
		LastVerifiedAt: timestamptzFromTimePtr(source.LastVerifiedAt),
	}, nil
}

func datasetVersionCreateParams(version models.DatasetVersion) (sqlc.CreateDatasetVersionParams, error) {
	datasetID, err := uuidFromString(version.DatasetID)
	if err != nil {
		return sqlc.CreateDatasetVersionParams{}, err
	}
	recordCount, err := int64ToInt32(version.RecordCount, "record_count")
	if err != nil {
		return sqlc.CreateDatasetVersionParams{}, err
	}

	var status any
	if version.Status != "" {
		status = string(version.Status)
	}

	return sqlc.CreateDatasetVersionParams{
		DatasetID:     datasetID,
		Version:       version.Version,
		SchemaVersion: textFromStringPtr(version.SchemaVersion),
		Format:        version.Format,
		Column5:       status,
		Column6:       recordCount,
		Checksum:      textFromStringPtr(version.Checksum),
		StoragePath:   textFromStringPtr(version.StoragePath),
		Notes:         textFromStringPtr(version.Notes),
		ReleasedAt:    timestamptzFromTimePtr(version.ReleasedAt),
	}, nil
}

func datasetVersionUpdateParams(version models.DatasetVersion) (sqlc.UpdateDatasetVersionParams, error) {
	id, err := uuidFromString(version.ID)
	if err != nil {
		return sqlc.UpdateDatasetVersionParams{}, err
	}
	recordCount, err := int64ToInt32(version.RecordCount, "record_count")
	if err != nil {
		return sqlc.UpdateDatasetVersionParams{}, err
	}

	return sqlc.UpdateDatasetVersionParams{
		ID:            id,
		SchemaVersion: textFromStringPtr(version.SchemaVersion),
		Format:        version.Format,
		Status:        string(version.Status),
		RecordCount:   recordCount,
		Checksum:      textFromStringPtr(version.Checksum),
		StoragePath:   textFromStringPtr(version.StoragePath),
		Notes:         textFromStringPtr(version.Notes),
		ReleasedAt:    timestamptzFromTimePtr(version.ReleasedAt),
	}, nil
}

func (r *DatasetRepository) Create(ctx context.Context, dataset models.Dataset) (models.Dataset, error) {
	params, err := datasetCreateParams(dataset)
	if err != nil {
		return models.Dataset{}, fmt.Errorf("create dataset: %w", err)
	}

	row, err := r.queries.CreateDataset(ctx, params)
	if err != nil {
		return models.Dataset{}, translateError("create dataset", err)
	}
	return datasetFromRow(row), nil
}

func (r *DatasetRepository) GetByID(ctx context.Context, id string) (models.Dataset, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Dataset{}, fmt.Errorf("get dataset by id: %w", err)
	}

	row, err := r.queries.GetDatasetByID(ctx, uid)
	if err != nil {
		return models.Dataset{}, translateError("get dataset by id", err)
	}
	return datasetFromRow(row), nil
}

func (r *DatasetRepository) GetByDatasetKey(ctx context.Context, datasetKey string) (models.Dataset, error) {
	row, err := r.queries.GetDatasetByKey(ctx, datasetKey)
	if err != nil {
		return models.Dataset{}, translateError("get dataset by dataset key", err)
	}
	return datasetFromRow(row), nil
}

func (r *DatasetRepository) ListPublic(ctx context.Context, limit, offset int32) ([]models.Dataset, error) {
	rows, err := r.queries.ListPublicDatasets(ctx, sqlc.ListPublicDatasetsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, translateError("list public datasets", err)
	}

	items := make([]models.Dataset, 0, len(rows))
	for _, row := range rows {
		items = append(items, datasetFromRow(row))
	}
	return items, nil
}

func (r *DatasetRepository) ListAll(ctx context.Context, limit, offset int32) ([]models.Dataset, error) {
	rows, err := r.queries.ListAllDatasets(ctx, sqlc.ListAllDatasetsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, translateError("list all datasets", err)
	}

	items := make([]models.Dataset, 0, len(rows))
	for _, row := range rows {
		items = append(items, datasetFromRow(row))
	}
	return items, nil
}

func (r *DatasetRepository) UpdateMetadata(ctx context.Context, dataset models.Dataset) (models.Dataset, error) {
	params, err := datasetUpdateParams(dataset)
	if err != nil {
		return models.Dataset{}, fmt.Errorf("update dataset metadata: %w", err)
	}

	row, err := r.queries.UpdateDatasetMetadata(ctx, params)
	if err != nil {
		return models.Dataset{}, translateError("update dataset metadata", err)
	}
	return datasetFromRow(row), nil
}

func (r *DatasetRepository) Archive(ctx context.Context, id string) (models.Dataset, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Dataset{}, fmt.Errorf("archive dataset: %w", err)
	}

	row, err := r.queries.ArchiveDataset(ctx, uid)
	if err != nil {
		return models.Dataset{}, translateError("archive dataset", err)
	}
	return datasetFromRow(row), nil
}

func (r *DatasetRepository) CreateSource(ctx context.Context, source models.DatasetSource) (models.DatasetSource, error) {
	params, err := datasetSourceCreateParams(source)
	if err != nil {
		return models.DatasetSource{}, fmt.Errorf("create dataset source: %w", err)
	}

	row, err := r.queries.CreateDatasetSource(ctx, params)
	if err != nil {
		return models.DatasetSource{}, translateError("create dataset source", err)
	}
	return datasetSourceFromRow(row), nil
}

func (r *DatasetRepository) GetSourceByID(ctx context.Context, id string) (models.DatasetSource, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.DatasetSource{}, fmt.Errorf("get dataset source by id: %w", err)
	}

	row, err := r.queries.GetDatasetSourceByID(ctx, uid)
	if err != nil {
		return models.DatasetSource{}, translateError("get dataset source by id", err)
	}
	return datasetSourceFromRow(row), nil
}

func (r *DatasetRepository) ListSources(ctx context.Context, datasetID string) ([]models.DatasetSource, error) {
	uid, err := uuidFromString(datasetID)
	if err != nil {
		return nil, fmt.Errorf("list dataset sources: %w", err)
	}

	rows, err := r.queries.ListDatasetSources(ctx, uid)
	if err != nil {
		return nil, translateError("list dataset sources", err)
	}

	items := make([]models.DatasetSource, 0, len(rows))
	for _, row := range rows {
		items = append(items, datasetSourceFromRow(row))
	}
	return items, nil
}

func (r *DatasetRepository) UpdateSource(ctx context.Context, source models.DatasetSource) (models.DatasetSource, error) {
	params, err := datasetSourceUpdateParams(source)
	if err != nil {
		return models.DatasetSource{}, fmt.Errorf("update dataset source: %w", err)
	}

	row, err := r.queries.UpdateDatasetSource(ctx, params)
	if err != nil {
		return models.DatasetSource{}, translateError("update dataset source", err)
	}
	return datasetSourceFromRow(row), nil
}

func (r *DatasetRepository) DeleteSource(ctx context.Context, id string) error {
	uid, err := uuidFromString(id)
	if err != nil {
		return fmt.Errorf("delete dataset source: %w", err)
	}
	return translateError("delete dataset source", r.queries.DeleteDatasetSource(ctx, uid))
}

func (r *DatasetRepository) CreateVersion(ctx context.Context, version models.DatasetVersion) (models.DatasetVersion, error) {
	params, err := datasetVersionCreateParams(version)
	if err != nil {
		return models.DatasetVersion{}, fmt.Errorf("create dataset version: %w", err)
	}

	row, err := r.queries.CreateDatasetVersion(ctx, params)
	if err != nil {
		return models.DatasetVersion{}, translateError("create dataset version", err)
	}
	return datasetVersionFromRow(row), nil
}

func (r *DatasetRepository) GetVersionByID(ctx context.Context, id string) (models.DatasetVersion, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.DatasetVersion{}, fmt.Errorf("get dataset version by id: %w", err)
	}

	row, err := r.queries.GetDatasetVersionByID(ctx, uid)
	if err != nil {
		return models.DatasetVersion{}, translateError("get dataset version by id", err)
	}
	return datasetVersionFromRow(row), nil
}

func (r *DatasetRepository) GetVersionByDatasetAndVersion(ctx context.Context, datasetID, versionName, format string) (models.DatasetVersion, error) {
	uid, err := uuidFromString(datasetID)
	if err != nil {
		return models.DatasetVersion{}, fmt.Errorf("get dataset version by dataset and version: %w", err)
	}

	row, err := r.queries.GetDatasetVersionByDatasetAndVersion(ctx, sqlc.GetDatasetVersionByDatasetAndVersionParams{
		DatasetID: uid,
		Version:   versionName,
		Lower:     format,
	})
	if err != nil {
		return models.DatasetVersion{}, translateError("get dataset version by dataset and version", err)
	}
	return datasetVersionFromRow(row), nil
}

func (r *DatasetRepository) ListVersions(ctx context.Context, datasetID string) ([]models.DatasetVersion, error) {
	uid, err := uuidFromString(datasetID)
	if err != nil {
		return nil, fmt.Errorf("list dataset versions: %w", err)
	}

	rows, err := r.queries.ListDatasetVersions(ctx, uid)
	if err != nil {
		return nil, translateError("list dataset versions", err)
	}

	items := make([]models.DatasetVersion, 0, len(rows))
	for _, row := range rows {
		items = append(items, datasetVersionFromRow(row))
	}
	return items, nil
}

func (r *DatasetRepository) UpdateVersion(ctx context.Context, version models.DatasetVersion) (models.DatasetVersion, error) {
	params, err := datasetVersionUpdateParams(version)
	if err != nil {
		return models.DatasetVersion{}, fmt.Errorf("update dataset version: %w", err)
	}

	row, err := r.queries.UpdateDatasetVersion(ctx, params)
	if err != nil {
		return models.DatasetVersion{}, translateError("update dataset version", err)
	}
	return datasetVersionFromRow(row), nil
}

func (r *DatasetRepository) PublishVersion(ctx context.Context, id string) (models.DatasetVersion, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.DatasetVersion{}, fmt.Errorf("publish dataset version: %w", err)
	}

	row, err := r.queries.PublishDatasetVersion(ctx, uid)
	if err != nil {
		return models.DatasetVersion{}, translateError("publish dataset version", err)
	}
	return datasetVersionFromRow(row), nil
}

func (r *DatasetRepository) DeleteVersion(ctx context.Context, id string) error {
	uid, err := uuidFromString(id)
	if err != nil {
		return fmt.Errorf("delete dataset version: %w", err)
	}
	return translateError("delete dataset version", r.queries.DeleteDatasetVersion(ctx, uid))
}
