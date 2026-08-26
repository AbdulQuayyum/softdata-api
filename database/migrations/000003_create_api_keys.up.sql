-- Migration: create api_keys table

CREATE TABLE IF NOT EXISTS api_keys (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name text NOT NULL,
    key_prefix text NOT NULL,
    key_hash text NOT NULL UNIQUE,
    key_last4 char(4) NOT NULL,
    status text NOT NULL DEFAULT 'active',
    last_used_at timestamptz NULL,
    expires_at timestamptz NULL,
    revoked_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT api_keys_name_not_blank CHECK (char_length(trim(name)) > 0),
    CONSTRAINT api_keys_prefix_not_blank CHECK (char_length(trim(key_prefix)) > 0),
    CONSTRAINT api_keys_hash_not_blank CHECK (char_length(trim(key_hash)) > 0),
    CONSTRAINT api_keys_last4_length CHECK (char_length(key_last4) = 4),
    CONSTRAINT api_keys_status_check CHECK (status IN ('active', 'revoked', 'expired'))
);

CREATE UNIQUE INDEX IF NOT EXISTS api_keys_prefix_unique
    ON api_keys (key_prefix);

CREATE INDEX IF NOT EXISTS api_keys_account_id_idx
    ON api_keys (account_id);

CREATE INDEX IF NOT EXISTS api_keys_status_idx
    ON api_keys (status);

CREATE INDEX IF NOT EXISTS api_keys_expires_at_idx
    ON api_keys (expires_at);
