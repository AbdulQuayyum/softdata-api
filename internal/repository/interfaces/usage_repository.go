package interfaces

import (
	"context"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// UsageRepository defines safe request and aggregate-usage persistence operations.
type UsageRepository interface {
	RecordRequest(ctx context.Context, request models.APIRequest) (models.APIRequest, error)
	UpsertAnonymousDaily(ctx context.Context, usageDate time.Time, anonymousID string, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes int64) error
	UpsertAccountDaily(ctx context.Context, usageDate time.Time, accountID string, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes int64) error
	UpsertAPIKeyDaily(ctx context.Context, usageDate time.Time, apiKeyID string, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes int64) error
	GetDailyByDate(ctx context.Context, usageDate time.Time) ([]models.UsageSummary, error)
	GetSummaryByAccountID(ctx context.Context, accountID string, limit, offset int32) ([]models.UsageSummaryResponse, error)
	GetSummaryByAPIKeyID(ctx context.Context, apiKeyID string, limit, offset int32) ([]models.UsageSummaryResponse, error)
	GetSummaryByAnonymousID(ctx context.Context, anonymousID string, limit, offset int32) ([]models.UsageSummaryResponse, error)
	GetDatasetGroupUsageByAccountID(ctx context.Context, accountID string, createdFrom, createdTo time.Time) ([]models.DatasetGroupUsageResponse, error)
	GetDatasetGroupUsageByAPIKeyID(ctx context.Context, apiKeyID string, createdFrom, createdTo time.Time) ([]models.DatasetGroupUsageResponse, error)
	CountRequestsByRoute(ctx context.Context, route string, createdFrom, createdTo time.Time) (int64, error)
	DeleteExpired(ctx context.Context, usageDate time.Time) error
}
