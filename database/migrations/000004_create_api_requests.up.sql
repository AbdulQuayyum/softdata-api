-- Migration: create api_requests table

CREATE TABLE IF NOT EXISTS api_requests (
    id bigserial PRIMARY KEY,
    request_id text NOT NULL UNIQUE,
    account_id uuid NULL REFERENCES accounts(id) ON DELETE SET NULL,
    api_key_id uuid NULL REFERENCES api_keys(id) ON DELETE SET NULL,
    anonymous_id uuid NULL,
    method text NOT NULL,
    path text NOT NULL,
    route text NULL,
    query_params jsonb NOT NULL DEFAULT '{}'::jsonb,
    status_code integer NOT NULL,
    ip_address inet NULL,
    user_agent text NULL,
    response_time_ms integer NULL,
    request_bytes bigint NULL,
    response_bytes bigint NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT api_requests_request_id_not_blank CHECK (char_length(trim(request_id)) > 0),
    CONSTRAINT api_requests_method_not_blank CHECK (char_length(trim(method)) > 0),
    CONSTRAINT api_requests_path_not_blank CHECK (char_length(trim(path)) > 0),
    CONSTRAINT api_requests_status_code_check CHECK (status_code >= 100 AND status_code <= 599),
    CONSTRAINT api_requests_response_time_check CHECK (response_time_ms IS NULL OR response_time_ms >= 0),
    CONSTRAINT api_requests_request_bytes_check CHECK (request_bytes IS NULL OR request_bytes >= 0),
    CONSTRAINT api_requests_response_bytes_check CHECK (response_bytes IS NULL OR response_bytes >= 0)
);

CREATE INDEX IF NOT EXISTS api_requests_created_at_idx
    ON api_requests (created_at DESC);

CREATE INDEX IF NOT EXISTS api_requests_account_id_idx
    ON api_requests (account_id);

CREATE INDEX IF NOT EXISTS api_requests_account_id_created_at_idx
    ON api_requests (account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS api_requests_api_key_id_idx
    ON api_requests (api_key_id);

CREATE INDEX IF NOT EXISTS api_requests_api_key_id_created_at_idx
    ON api_requests (api_key_id, created_at DESC);

CREATE INDEX IF NOT EXISTS api_requests_anonymous_id_idx
    ON api_requests (anonymous_id);

CREATE INDEX IF NOT EXISTS api_requests_anonymous_id_created_at_idx
    ON api_requests (anonymous_id, created_at DESC);

CREATE INDEX IF NOT EXISTS api_requests_route_idx
    ON api_requests (route);

CREATE INDEX IF NOT EXISTS api_requests_route_created_at_idx
    ON api_requests (route, created_at DESC);
