-- Migration: create sessions table

CREATE TABLE IF NOT EXISTS sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    refresh_token_hash text NOT NULL UNIQUE,
    access_token_jti uuid NULL UNIQUE,
    user_agent text NULL,
    ip_address inet NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz NULL,
    last_used_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS sessions_account_id_idx
    ON sessions (account_id);

CREATE INDEX IF NOT EXISTS sessions_account_id_created_at_idx
    ON sessions (account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS sessions_account_id_created_at_id_idx
    ON sessions (account_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS sessions_expires_at_idx
    ON sessions (expires_at);

CREATE INDEX IF NOT EXISTS sessions_revoked_at_idx
    ON sessions (revoked_at);

CREATE INDEX IF NOT EXISTS sessions_expires_at_revoked_at_idx
    ON sessions (expires_at, revoked_at);
