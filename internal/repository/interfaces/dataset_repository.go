package interfaces

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// DatasetRepository defines dataset, source, and version persistence operations.
type DatasetRepository interface {
	Create(ctx context.Context, dataset models.Dataset) (models.Dataset, error)
	GetByID(ctx context.Context, id string) (models.Dataset, error)
	GetByDatasetKey(ctx context.Context, datasetKey string) (models.Dataset, error)
	ListPublic(ctx context.Context, limit, offset int32) ([]models.Dataset, error)
	ListAll(ctx context.Context, limit, offset int32) ([]models.Dataset, error)
	UpdateMetadata(ctx context.Context, dataset models.Dataset) (models.Dataset, error)
	Archive(ctx context.Context, id string) (models.Dataset, error)

	CreateSource(ctx context.Context, source models.DatasetSource) (models.DatasetSource, error)
	GetSourceByID(ctx context.Context, id string) (models.DatasetSource, error)
	ListSources(ctx context.Context, datasetID string) ([]models.DatasetSource, error)
	UpdateSource(ctx context.Context, source models.DatasetSource) (models.DatasetSource, error)
	DeleteSource(ctx context.Context, id string) error

	CreateVersion(ctx context.Context, version models.DatasetVersion) (models.DatasetVersion, error)
	GetVersionByID(ctx context.Context, id string) (models.DatasetVersion, error)
	GetVersionByDatasetAndVersion(ctx context.Context, datasetID, versionName, format string) (models.DatasetVersion, error)
	ListVersions(ctx context.Context, datasetID string) ([]models.DatasetVersion, error)
	UpdateVersion(ctx context.Context, version models.DatasetVersion) (models.DatasetVersion, error)
	PublishVersion(ctx context.Context, id string) (models.DatasetVersion, error)
	DeleteVersion(ctx context.Context, id string) error
}
