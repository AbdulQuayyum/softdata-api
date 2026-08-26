package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/netip"
	"time"

	sqlc "github.com/AbdulQuayyum/softdata-api/internal/database/sqlc"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var conflictConstraints = map[string]struct{}{
	"accounts_username_unique":                       {},
	"accounts_email_unique":                          {},
	"sessions_refresh_token_hash_key":                {},
	"sessions_access_token_jti_key":                  {},
	"api_keys_key_hash_key":                          {},
	"api_keys_prefix_unique":                         {},
	"api_requests_request_id_key":                    {},
	"datasets_dataset_key_unique":                    {},
	"datasets_slug_unique":                           {},
	"dataset_sources_dataset_id_source_key_unique":   {},
	"dataset_versions_dataset_version_format_unique": {},
	"usage_daily_anonymous_unique_idx":               {},
	"usage_daily_account_unique_idx":                 {},
	"usage_daily_api_key_unique_idx":                 {},
}

func translateError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return fmt.Errorf("%w: %s", interfaces.ErrNotFound, op)
	default:
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if _, ok := conflictConstraints[pgErr.ConstraintName]; ok {
				return fmt.Errorf("%w: %s", interfaces.ErrConflict, op)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
}

func uuidFromString(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		return pgtype.UUID{}, fmt.Errorf("parse uuid: %w", err)
	}
	return id, nil
}

func uuidFromStringPtr(value *string) (pgtype.UUID, error) {
	if value == nil {
		return pgtype.UUID{}, nil
	}
	return uuidFromString(*value)
}

func uuidToString(value pgtype.UUID) *string {
	if !value.Valid {
		return nil
	}
	s := value.String()
	return &s
}

func textFromString(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}

func textFromStringPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return textFromString(*value)
}

func stringPtrFromText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	s := value.String
	return &s
}

func timestamptzFromTimePtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}

func timePtrFromTimestamptz(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func dateFromTimePtr(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *value, Valid: true}
}

func dateFromTime(value time.Time) pgtype.Date {
	return pgtype.Date{Time: value, Valid: true}
}

func timePtrFromDate(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func stringPtrFromNetAddr(value *netip.Addr) *string {
	if value == nil {
		return nil
	}
	s := value.String()
	return &s
}

func netAddrPtrFromString(value *string) (*netip.Addr, error) {
	if value == nil {
		return nil, nil
	}
	addr, err := netip.ParseAddr(*value)
	if err != nil {
		return nil, fmt.Errorf("parse ip address: %w", err)
	}
	return &addr, nil
}

func int64ToInt32(value int64, field string) (int32, error) {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, fmt.Errorf("%s out of range", field)
	}
	return int32(value), nil
}

func pgInt4FromInt64Ptr(value *int64, field string) (pgtype.Int4, error) {
	if value == nil {
		return pgtype.Int4{}, nil
	}
	converted, err := int64ToInt32(*value, field)
	if err != nil {
		return pgtype.Int4{}, err
	}
	return pgtype.Int4{Int32: converted, Valid: true}, nil
}

func pgInt8FromInt64Ptr(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func queryParamsFromBytes(data []byte) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var params map[string]any
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, fmt.Errorf("decode query params: %w", err)
	}
	if params == nil {
		params = map[string]any{}
	}
	return params, nil
}

func queryParamsToBytes(params map[string]any) ([]byte, error) {
	if params == nil {
		return nil, nil
	}
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode query params: %w", err)
	}
	return data, nil
}

func accountFromRow(row sqlc.Account) models.Account {
	return models.Account{
		ID:              row.ID.String(),
		Username:        row.Username,
		Email:           stringPtrFromText(row.Email),
		PasswordHash:    row.PasswordHash,
		Status:          models.AccountStatus(row.Status),
		EmailVerifiedAt: timePtrFromTimestamptz(row.EmailVerifiedAt),
		LastLoginAt:     timePtrFromTimestamptz(row.LastLoginAt),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DeletedAt:       timePtrFromTimestamptz(row.DeletedAt),
	}
}

func apiKeyFromRow(row sqlc.ApiKey) models.APIKey {
	return models.APIKey{
		ID:         row.ID.String(),
		AccountID:  row.AccountID.String(),
		Name:       row.Name,
		KeyPrefix:  row.KeyPrefix,
		KeyHash:    row.KeyHash,
		KeyLast4:   row.KeyLast4,
		Status:     models.APIKeyStatus(row.Status),
		LastUsedAt: timePtrFromTimestamptz(row.LastUsedAt),
		ExpiresAt:  timePtrFromTimestamptz(row.ExpiresAt),
		RevokedAt:  timePtrFromTimestamptz(row.RevokedAt),
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}

func apiRequestFromRow(row sqlc.ApiRequest) (models.APIRequest, error) {
	queryParams, err := queryParamsFromBytes(row.QueryParams)
	if err != nil {
		return models.APIRequest{}, err
	}

	return models.APIRequest{
		ID:             row.ID,
		RequestID:      row.RequestID,
		AccountID:      uuidToString(row.AccountID),
		APIKeyID:       uuidToString(row.ApiKeyID),
		AnonymousID:    uuidToString(row.AnonymousID),
		Method:         row.Method,
		Path:           row.Path,
		Route:          stringPtrFromText(row.Route),
		QueryParams:    queryParams,
		StatusCode:     int(row.StatusCode),
		IPAddress:      stringPtrFromNetAddr(row.IpAddress),
		UserAgent:      stringPtrFromText(row.UserAgent),
		ResponseTimeMS: int64PtrFromInt4(row.ResponseTimeMs),
		RequestBytes:   int64PtrFromInt8(row.RequestBytes),
		ResponseBytes:  int64PtrFromInt8(row.ResponseBytes),
		CreatedAt:      row.CreatedAt.Time,
	}, nil
}

func datasetFromRow(row sqlc.Dataset) models.Dataset {
	return models.Dataset{
		ID:              row.ID.String(),
		DatasetKey:      row.DatasetKey,
		Slug:            row.Slug,
		Name:            row.Name,
		Description:     stringPtrFromText(row.Description),
		GroupName:       row.GroupName,
		CountryCode:     stringPtrFromText(row.CountryCode),
		Version:         row.Version,
		Status:          models.DatasetStatus(row.Status),
		RecordCount:     int64(row.RecordCount),
		PrimaryFormat:   row.PrimaryFormat,
		Formats:         cloneStrings(row.Formats),
		SchemaPath:      stringPtrFromText(row.SchemaPath),
		LicenceID:       stringPtrFromText(row.LicenceID),
		SourceCount:     int64(row.SourceCount),
		UpdateFrequency: stringPtrFromText(row.UpdateFrequency),
		LastUpdatedAt:   timePtrFromDate(row.LastUpdatedAt),
		LastVerifiedAt:  timePtrFromDate(row.LastVerifiedAt),
		Maintainers:     cloneStrings(row.Maintainers),
		IsPublic:        row.IsPublic,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		ArchivedAt:      timePtrFromTimestamptz(row.ArchivedAt),
	}
}

func datasetSourceFromRow(row sqlc.DatasetSource) models.DatasetSource {
	return models.DatasetSource{
		ID:             row.ID.String(),
		DatasetID:      row.DatasetID.String(),
		SourceKey:      row.SourceKey,
		Name:           row.Name,
		URL:            stringPtrFromText(row.Url),
		Description:    stringPtrFromText(row.Description),
		Publisher:      stringPtrFromText(row.Publisher),
		SourceType:     stringPtrFromText(row.SourceType),
		LicenceID:      stringPtrFromText(row.LicenceID),
		IsOfficial:     row.IsOfficial,
		LastFetchedAt:  timePtrFromTimestamptz(row.LastFetchedAt),
		LastVerifiedAt: timePtrFromTimestamptz(row.LastVerifiedAt),
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
}

func datasetVersionFromRow(row sqlc.DatasetVersion) models.DatasetVersion {
	return models.DatasetVersion{
		ID:            row.ID.String(),
		DatasetID:     row.DatasetID.String(),
		Version:       row.Version,
		SchemaVersion: stringPtrFromText(row.SchemaVersion),
		Format:        row.Format,
		Status:        models.DatasetVersionStatus(row.Status),
		RecordCount:   int64(row.RecordCount),
		Checksum:      stringPtrFromText(row.Checksum),
		StoragePath:   stringPtrFromText(row.StoragePath),
		Notes:         stringPtrFromText(row.Notes),
		ReleasedAt:    timePtrFromTimestamptz(row.ReleasedAt),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func sessionFromRow(row sqlc.Session) models.Session {
	return models.Session{
		ID:               row.ID.String(),
		AccountID:        row.AccountID.String(),
		RefreshTokenHash: row.RefreshTokenHash,
		AccessTokenJTI:   uuidToString(row.AccessTokenJti),
		UserAgent:        stringPtrFromText(row.UserAgent),
		IPAddress:        stringPtrFromNetAddr(row.IpAddress),
		ExpiresAt:        row.ExpiresAt.Time,
		RevokedAt:        timePtrFromTimestamptz(row.RevokedAt),
		LastUsedAt:       timePtrFromTimestamptz(row.LastUsedAt),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func usageSummaryFromRow(row sqlc.UsageDaily) models.UsageSummary {
	return models.UsageSummary{
		ID:                   row.ID,
		UsageDate:            row.UsageDate.Time,
		ScopeType:            models.UsageScopeType(row.ScopeType),
		AccountID:            uuidToString(row.AccountID),
		APIKeyID:             uuidToString(row.ApiKeyID),
		AnonymousID:          uuidToString(row.AnonymousID),
		RequestCount:         int64(row.RequestCount),
		SuccessfulCount:      int64(row.SuccessfulCount),
		ErrorCount:           int64(row.ErrorCount),
		DatasetDownloadCount: int64(row.DatasetDownloadCount),
		ResponseBytes:        row.ResponseBytes,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
	}
}

func usageSummaryResponse(scope models.UsageScopeType, usageDate pgtype.Date, requestCount, successfulCount, errorCount, datasetDownloadCount int32, responseBytes int64) models.UsageSummaryResponse {
	date := ""
	if usageDate.Valid {
		date = usageDate.Time.Format("2006-01-02")
	}

	return models.UsageSummaryResponse{
		UsageDate:            date,
		ScopeType:            scope,
		RequestCount:         int64(requestCount),
		SuccessfulCount:      int64(successfulCount),
		ErrorCount:           int64(errorCount),
		DatasetDownloadCount: int64(datasetDownloadCount),
		ResponseBytes:        responseBytes,
	}
}

func int64PtrFromInt4(value pgtype.Int4) *int64 {
	if !value.Valid {
		return nil
	}
	v := int64(value.Int32)
	return &v
}

func int64PtrFromInt8(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}
