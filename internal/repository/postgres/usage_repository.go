package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	sqlc "github.com/AbdulQuayyum/softdata-api/internal/database/sqlc"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/jackc/pgx/v5/pgtype"
)

var _ interfaces.UsageRepository = (*UsageRepository)(nil)

type UsageRepository struct {
	queries *sqlc.Queries
}

func NewUsageRepository(dbt sqlc.DBTX) *UsageRepository {
	return &UsageRepository{queries: sqlc.New(dbt)}
}

func (r *UsageRepository) RecordRequest(ctx context.Context, request models.APIRequest) (models.APIRequest, error) {
	accountID, err := uuidFromStringPtr(request.AccountID)
	if err != nil {
		return models.APIRequest{}, fmt.Errorf("record api request: %w", err)
	}
	apiKeyID, err := uuidFromStringPtr(request.APIKeyID)
	if err != nil {
		return models.APIRequest{}, fmt.Errorf("record api request: %w", err)
	}
	anonymousID, err := uuidFromStringPtr(request.AnonymousID)
	if err != nil {
		return models.APIRequest{}, fmt.Errorf("record api request: %w", err)
	}
	route := textFromStringPtr(request.Route)
	queryParams, err := queryParamsToBytes(request.QueryParams)
	if err != nil {
		return models.APIRequest{}, fmt.Errorf("record api request: %w", err)
	}
	statusCode, err := int64ToInt32(int64(request.StatusCode), "status_code")
	if err != nil {
		return models.APIRequest{}, fmt.Errorf("record api request: %w", err)
	}
	ipAddress, err := netAddrPtrFromString(request.IPAddress)
	if err != nil {
		return models.APIRequest{}, fmt.Errorf("record api request: %w", err)
	}
	responseTime, err := pgInt4FromInt64Ptr(request.ResponseTimeMS, "response_time_ms")
	if err != nil {
		return models.APIRequest{}, fmt.Errorf("record api request: %w", err)
	}
	requestBytes := pgInt8FromInt64Ptr(request.RequestBytes)
	responseBytes := pgInt8FromInt64Ptr(request.ResponseBytes)

	row, err := r.queries.InsertAPIRequestLog(ctx, sqlc.InsertAPIRequestLogParams{
		RequestID:      request.RequestID,
		AccountID:      accountID,
		ApiKeyID:       apiKeyID,
		AnonymousID:    anonymousID,
		Method:         request.Method,
		Path:           request.Path,
		Route:          route,
		DatasetGroup:   textFromStringPtr(request.DatasetGroup),
		Column9:        queryParams,
		StatusCode:     statusCode,
		IpAddress:      ipAddress,
		UserAgent:      textFromStringPtr(request.UserAgent),
		ResponseTimeMs: responseTime,
		RequestBytes:   requestBytes,
		ResponseBytes:  responseBytes,
	})
	if err != nil {
		return models.APIRequest{}, translateError("record api request", err)
	}
	return apiRequestFromRow(row)
}

func (r *UsageRepository) UpsertAnonymousDaily(ctx context.Context, usageDate time.Time, anonymousID string, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes int64) error {
	anonymousUUID, err := uuidFromString(anonymousID)
	if err != nil {
		return fmt.Errorf("upsert anonymous usage daily: %w", err)
	}
	reqCount, err := int64ToInt32(requestCount, "request_count")
	if err != nil {
		return fmt.Errorf("upsert anonymous usage daily: %w", err)
	}
	successCount, err := int64ToInt32(successfulCount, "successful_count")
	if err != nil {
		return fmt.Errorf("upsert anonymous usage daily: %w", err)
	}
	errCount, err := int64ToInt32(errorCount, "error_count")
	if err != nil {
		return fmt.Errorf("upsert anonymous usage daily: %w", err)
	}
	downloadCount, err := int64ToInt32(datasetDownloadCount, "dataset_download_count")
	if err != nil {
		return fmt.Errorf("upsert anonymous usage daily: %w", err)
	}

	return translateError("upsert anonymous usage daily", r.queries.UpsertAnonymousUsageDaily(ctx, sqlc.UpsertAnonymousUsageDailyParams{
		UsageDate:            dateFromTime(usageDate),
		AnonymousID:          anonymousUUID,
		RequestCount:         reqCount,
		SuccessfulCount:      successCount,
		ErrorCount:           errCount,
		DatasetDownloadCount: downloadCount,
		ResponseBytes:        responseBytes,
	}))
}

func (r *UsageRepository) UpsertAccountDaily(ctx context.Context, usageDate time.Time, accountID string, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes int64) error {
	accountUUID, err := uuidFromString(accountID)
	if err != nil {
		return fmt.Errorf("upsert account usage daily: %w", err)
	}
	reqCount, err := int64ToInt32(requestCount, "request_count")
	if err != nil {
		return fmt.Errorf("upsert account usage daily: %w", err)
	}
	successCount, err := int64ToInt32(successfulCount, "successful_count")
	if err != nil {
		return fmt.Errorf("upsert account usage daily: %w", err)
	}
	errCount, err := int64ToInt32(errorCount, "error_count")
	if err != nil {
		return fmt.Errorf("upsert account usage daily: %w", err)
	}
	downloadCount, err := int64ToInt32(datasetDownloadCount, "dataset_download_count")
	if err != nil {
		return fmt.Errorf("upsert account usage daily: %w", err)
	}

	return translateError("upsert account usage daily", r.queries.UpsertAccountUsageDaily(ctx, sqlc.UpsertAccountUsageDailyParams{
		UsageDate:            dateFromTime(usageDate),
		AccountID:            accountUUID,
		RequestCount:         reqCount,
		SuccessfulCount:      successCount,
		ErrorCount:           errCount,
		DatasetDownloadCount: downloadCount,
		ResponseBytes:        responseBytes,
	}))
}

func (r *UsageRepository) UpsertAPIKeyDaily(ctx context.Context, usageDate time.Time, apiKeyID string, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes int64) error {
	apiKeyUUID, err := uuidFromString(apiKeyID)
	if err != nil {
		return fmt.Errorf("upsert api key usage daily: %w", err)
	}
	reqCount, err := int64ToInt32(requestCount, "request_count")
	if err != nil {
		return fmt.Errorf("upsert api key usage daily: %w", err)
	}
	successCount, err := int64ToInt32(successfulCount, "successful_count")
	if err != nil {
		return fmt.Errorf("upsert api key usage daily: %w", err)
	}
	errCount, err := int64ToInt32(errorCount, "error_count")
	if err != nil {
		return fmt.Errorf("upsert api key usage daily: %w", err)
	}
	downloadCount, err := int64ToInt32(datasetDownloadCount, "dataset_download_count")
	if err != nil {
		return fmt.Errorf("upsert api key usage daily: %w", err)
	}

	return translateError("upsert api key usage daily", r.queries.UpsertAPIKeyUsageDaily(ctx, sqlc.UpsertAPIKeyUsageDailyParams{
		UsageDate:            dateFromTime(usageDate),
		ApiKeyID:             apiKeyUUID,
		RequestCount:         reqCount,
		SuccessfulCount:      successCount,
		ErrorCount:           errCount,
		DatasetDownloadCount: downloadCount,
		ResponseBytes:        responseBytes,
	}))
}

func (r *UsageRepository) GetDailyByDate(ctx context.Context, usageDate time.Time) ([]models.UsageSummary, error) {
	rows, err := r.queries.GetUsageDailyByDate(ctx, dateFromTime(usageDate))
	if err != nil {
		return nil, translateError("get usage daily by date", err)
	}

	items := make([]models.UsageSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, usageSummaryFromRow(row))
	}
	return items, nil
}

func (r *UsageRepository) GetSummaryByAccountID(ctx context.Context, accountID string, limit, offset int32) ([]models.UsageSummaryResponse, error) {
	accountUUID, err := uuidFromString(accountID)
	if err != nil {
		return nil, fmt.Errorf("get usage summary by account id: %w", err)
	}

	rows, err := r.queries.GetUsageSummaryByAccountID(ctx, sqlc.GetUsageSummaryByAccountIDParams{
		AccountID: accountUUID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, translateError("get usage summary by account id", err)
	}

	items := make([]models.UsageSummaryResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, usageSummaryResponse(
			models.UsageScopeAccount,
			row.UsageDate,
			row.RequestCount,
			row.SuccessfulCount,
			row.ErrorCount,
			row.DatasetDownloadCount,
			row.ResponseBytes,
		))
	}
	return items, nil
}

func (r *UsageRepository) GetSummaryByAPIKeyID(ctx context.Context, apiKeyID string, limit, offset int32) ([]models.UsageSummaryResponse, error) {
	apiKeyUUID, err := uuidFromString(apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("get usage summary by api key id: %w", err)
	}

	rows, err := r.queries.GetUsageSummaryByAPIKeyID(ctx, sqlc.GetUsageSummaryByAPIKeyIDParams{
		ApiKeyID: apiKeyUUID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, translateError("get usage summary by api key id", err)
	}

	items := make([]models.UsageSummaryResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, usageSummaryResponse(
			models.UsageScopeAPIKey,
			row.UsageDate,
			row.RequestCount,
			row.SuccessfulCount,
			row.ErrorCount,
			row.DatasetDownloadCount,
			row.ResponseBytes,
		))
	}
	return items, nil
}

func (r *UsageRepository) GetSummaryByAnonymousID(ctx context.Context, anonymousID string, limit, offset int32) ([]models.UsageSummaryResponse, error) {
	anonymousUUID, err := uuidFromString(anonymousID)
	if err != nil {
		return nil, fmt.Errorf("get usage summary by anonymous id: %w", err)
	}

	rows, err := r.queries.GetUsageSummaryByAnonymousID(ctx, sqlc.GetUsageSummaryByAnonymousIDParams{
		AnonymousID: anonymousUUID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, translateError("get usage summary by anonymous id", err)
	}

	items := make([]models.UsageSummaryResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, usageSummaryResponse(
			models.UsageScopeAnonymous,
			row.UsageDate,
			row.RequestCount,
			row.SuccessfulCount,
			row.ErrorCount,
			row.DatasetDownloadCount,
			row.ResponseBytes,
		))
	}
	return items, nil
}

func (r *UsageRepository) GetDatasetGroupUsageByAccountID(ctx context.Context, accountID string, createdFrom, createdTo time.Time) ([]models.DatasetGroupUsageResponse, error) {
	accountUUID, err := uuidFromString(accountID)
	if err != nil {
		return nil, fmt.Errorf("get dataset group usage by account id: %w", err)
	}

	rows, err := r.queries.GetDatasetGroupUsageByAccountID(ctx, sqlc.GetDatasetGroupUsageByAccountIDParams{
		AccountID:   accountUUID,
		CreatedAt:   timestamptzFromTimePtr(&createdFrom),
		CreatedAt_2: timestamptzFromTimePtr(&createdTo),
	})
	if err != nil {
		return nil, translateError("get dataset group usage by account id", err)
	}

	items := make([]models.DatasetGroupUsageResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, datasetGroupUsageResponse(row.DatasetGroup.String, row.RequestCount))
	}
	return items, nil
}

func (r *UsageRepository) GetDatasetGroupUsageByAPIKeyID(ctx context.Context, apiKeyID string, createdFrom, createdTo time.Time) ([]models.DatasetGroupUsageResponse, error) {
	apiKeyUUID, err := uuidFromString(apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("get dataset group usage by api key id: %w", err)
	}

	rows, err := r.queries.GetDatasetGroupUsageByAPIKeyID(ctx, sqlc.GetDatasetGroupUsageByAPIKeyIDParams{
		ApiKeyID:    apiKeyUUID,
		CreatedAt:   timestamptzFromTimePtr(&createdFrom),
		CreatedAt_2: timestamptzFromTimePtr(&createdTo),
	})
	if err != nil {
		return nil, translateError("get dataset group usage by api key id", err)
	}

	items := make([]models.DatasetGroupUsageResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, datasetGroupUsageResponse(row.DatasetGroup.String, row.RequestCount))
	}
	return items, nil
}

func (r *UsageRepository) GetEndpointUsageByAccountID(ctx context.Context, accountID string, createdFrom, createdTo time.Time) ([]models.EndpointUsageResponse, error) {
	accountUUID, err := uuidFromString(accountID)
	if err != nil {
		return nil, fmt.Errorf("get endpoint usage by account id: %w", err)
	}

	rows, err := r.queries.GetEndpointUsageByAccountID(ctx, sqlc.GetEndpointUsageByAccountIDParams{
		AccountID:   accountUUID,
		CreatedAt:   timestamptzFromTimePtr(&createdFrom),
		CreatedAt_2: timestamptzFromTimePtr(&createdTo),
	})
	if err != nil {
		return nil, translateError("get endpoint usage by account id", err)
	}

	items := make([]models.EndpointUsageResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, endpointUsageResponse(row.Route, row.RequestCount))
	}
	return items, nil
}

func (r *UsageRepository) GetEndpointUsageByAPIKeyID(ctx context.Context, accountID, apiKeyID string, createdFrom, createdTo time.Time) ([]models.EndpointUsageResponse, error) {
	accountUUID, err := uuidFromString(accountID)
	if err != nil {
		return nil, fmt.Errorf("get endpoint usage by api key id: %w", err)
	}
	apiKeyUUID, err := uuidFromString(apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("get endpoint usage by api key id: %w", err)
	}

	rows, err := r.queries.GetEndpointUsageByAPIKeyID(ctx, sqlc.GetEndpointUsageByAPIKeyIDParams{
		AccountID:   accountUUID,
		ApiKeyID:    apiKeyUUID,
		CreatedAt:   timestamptzFromTimePtr(&createdFrom),
		CreatedAt_2: timestamptzFromTimePtr(&createdTo),
	})
	if err != nil {
		return nil, translateError("get endpoint usage by api key id", err)
	}

	items := make([]models.EndpointUsageResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, endpointUsageResponse(row.Route, row.RequestCount))
	}
	return items, nil
}

func (r *UsageRepository) CountRequestsByRoute(ctx context.Context, route string, createdFrom, createdTo time.Time) (int64, error) {
	count, err := r.queries.CountRequestsByRoute(ctx, sqlc.CountRequestsByRouteParams{
		Route:       textFromString(route),
		CreatedAt:   timestamptzFromTimePtr(&createdFrom),
		CreatedAt_2: timestamptzFromTimePtr(&createdTo),
	})
	if err != nil {
		return 0, translateError("count requests by route", err)
	}
	return count, nil
}

func (r *UsageRepository) DeleteExpired(ctx context.Context, usageDate time.Time) error {
	return translateError("delete expired usage daily", r.queries.DeleteExpiredUsageDaily(ctx, dateFromTime(usageDate)))
}

func endpointUsageResponse(route pgtype.Text, requestCount int64) models.EndpointUsageResponse {
	return models.EndpointUsageResponse{
		Endpoint:     strings.TrimSpace(route.String),
		RequestCount: requestCount,
	}
}
