-- Usage queries

-- name: UpsertAnonymousUsageDaily :exec
INSERT INTO usage_daily (
    usage_date,
    scope_type,
    anonymous_id,
    request_count,
    successful_count,
    error_count,
    dataset_download_count,
    response_bytes
)
SELECT
    $1,
    'anonymous',
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
ON CONFLICT (usage_date, anonymous_id) WHERE scope_type = 'anonymous'
DO UPDATE SET
    request_count = usage_daily.request_count + EXCLUDED.request_count,
    successful_count = usage_daily.successful_count + EXCLUDED.successful_count,
    error_count = usage_daily.error_count + EXCLUDED.error_count,
    dataset_download_count = usage_daily.dataset_download_count + EXCLUDED.dataset_download_count,
    response_bytes = usage_daily.response_bytes + EXCLUDED.response_bytes,
    updated_at = now();

-- name: UpsertAccountUsageDaily :exec
INSERT INTO usage_daily (
    usage_date,
    scope_type,
    account_id,
    request_count,
    successful_count,
    error_count,
    dataset_download_count,
    response_bytes
)
SELECT
    $1,
    'account',
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
ON CONFLICT (usage_date, account_id) WHERE scope_type = 'account'
DO UPDATE SET
    request_count = usage_daily.request_count + EXCLUDED.request_count,
    successful_count = usage_daily.successful_count + EXCLUDED.successful_count,
    error_count = usage_daily.error_count + EXCLUDED.error_count,
    dataset_download_count = usage_daily.dataset_download_count + EXCLUDED.dataset_download_count,
    response_bytes = usage_daily.response_bytes + EXCLUDED.response_bytes,
    updated_at = now();

-- name: UpsertAPIKeyUsageDaily :exec
INSERT INTO usage_daily (
    usage_date,
    scope_type,
    api_key_id,
    request_count,
    successful_count,
    error_count,
    dataset_download_count,
    response_bytes
)
SELECT
    $1,
    'api_key',
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
ON CONFLICT (usage_date, api_key_id) WHERE scope_type = 'api_key'
DO UPDATE SET
    request_count = usage_daily.request_count + EXCLUDED.request_count,
    successful_count = usage_daily.successful_count + EXCLUDED.successful_count,
    error_count = usage_daily.error_count + EXCLUDED.error_count,
    dataset_download_count = usage_daily.dataset_download_count + EXCLUDED.dataset_download_count,
    response_bytes = usage_daily.response_bytes + EXCLUDED.response_bytes,
    updated_at = now();

-- name: InsertAPIRequestLog :one
INSERT INTO api_requests (
    request_id,
    account_id,
    api_key_id,
    anonymous_id,
    method,
    path,
    route,
    query_params,
    status_code,
    ip_address,
    user_agent,
    response_time_ms,
    request_bytes,
    response_bytes
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    COALESCE($8, '{}'::jsonb),
    $9,
    $10,
    $11,
    $12,
    $13,
    $14
)
RETURNING *;

-- name: GetUsageDailyByDate :many
SELECT *
FROM usage_daily
WHERE usage_date = $1
ORDER BY scope_type, account_id, api_key_id, anonymous_id;

-- name: GetUsageSummaryByAccountID :many
SELECT
    usage_date,
    request_count,
    successful_count,
    error_count,
    dataset_download_count,
    response_bytes
FROM usage_daily
WHERE account_id = $1
ORDER BY usage_date DESC
LIMIT $2 OFFSET $3;

-- name: GetUsageSummaryByAPIKeyID :many
SELECT
    usage_date,
    request_count,
    successful_count,
    error_count,
    dataset_download_count,
    response_bytes
FROM usage_daily
WHERE api_key_id = $1
ORDER BY usage_date DESC
LIMIT $2 OFFSET $3;

-- name: GetUsageSummaryByAnonymousID :many
SELECT
    usage_date,
    request_count,
    successful_count,
    error_count,
    dataset_download_count,
    response_bytes
FROM usage_daily
WHERE anonymous_id = $1
ORDER BY usage_date DESC
LIMIT $2 OFFSET $3;

-- name: CountRequestsByRoute :one
SELECT COUNT(*)::bigint AS request_count
FROM api_requests
WHERE route = $1
  AND created_at >= $2
  AND created_at < $3;

-- name: DeleteExpiredUsageDaily :exec
DELETE FROM usage_daily
WHERE usage_date < $1;
