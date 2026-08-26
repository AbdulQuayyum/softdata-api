-- Migration: create usage_daily table

CREATE TABLE IF NOT EXISTS usage_daily (
    id bigserial PRIMARY KEY,
    usage_date date NOT NULL,
    scope_type text NOT NULL,
    account_id uuid NULL REFERENCES accounts(id) ON DELETE SET NULL,
    api_key_id uuid NULL REFERENCES api_keys(id) ON DELETE SET NULL,
    anonymous_id uuid NULL,
    request_count integer NOT NULL DEFAULT 0,
    successful_count integer NOT NULL DEFAULT 0,
    error_count integer NOT NULL DEFAULT 0,
    dataset_download_count integer NOT NULL DEFAULT 0,
    response_bytes bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT usage_daily_scope_check CHECK (scope_type IN ('anonymous', 'account', 'api_key')),
    CONSTRAINT usage_daily_scope_identity_check CHECK (
        (scope_type = 'anonymous' AND anonymous_id IS NOT NULL AND account_id IS NULL AND api_key_id IS NULL) OR
        (scope_type = 'account' AND account_id IS NOT NULL AND api_key_id IS NULL AND anonymous_id IS NULL) OR
        (scope_type = 'api_key' AND api_key_id IS NOT NULL AND account_id IS NULL AND anonymous_id IS NULL)
    ),
    CONSTRAINT usage_daily_request_count_check CHECK (request_count >= 0),
    CONSTRAINT usage_daily_successful_count_check CHECK (successful_count >= 0),
    CONSTRAINT usage_daily_error_count_check CHECK (error_count >= 0),
    CONSTRAINT usage_daily_dataset_download_count_check CHECK (dataset_download_count >= 0),
    CONSTRAINT usage_daily_response_bytes_check CHECK (response_bytes >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS usage_daily_anonymous_unique_idx
    ON usage_daily (usage_date, anonymous_id)
    WHERE scope_type = 'anonymous';

CREATE INDEX IF NOT EXISTS usage_daily_anonymous_id_usage_date_idx
    ON usage_daily (anonymous_id, usage_date DESC)
    WHERE scope_type = 'anonymous';

CREATE UNIQUE INDEX IF NOT EXISTS usage_daily_account_unique_idx
    ON usage_daily (usage_date, account_id)
    WHERE scope_type = 'account';

CREATE INDEX IF NOT EXISTS usage_daily_account_id_usage_date_idx
    ON usage_daily (account_id, usage_date DESC)
    WHERE scope_type = 'account';

CREATE UNIQUE INDEX IF NOT EXISTS usage_daily_api_key_unique_idx
    ON usage_daily (usage_date, api_key_id)
    WHERE scope_type = 'api_key';

CREATE INDEX IF NOT EXISTS usage_daily_api_key_id_usage_date_idx
    ON usage_daily (api_key_id, usage_date DESC)
    WHERE scope_type = 'api_key';

CREATE INDEX IF NOT EXISTS usage_daily_usage_date_idx
    ON usage_daily (usage_date DESC);

CREATE INDEX IF NOT EXISTS usage_daily_account_id_idx
    ON usage_daily (account_id);

CREATE INDEX IF NOT EXISTS usage_daily_api_key_id_idx
    ON usage_daily (api_key_id);
