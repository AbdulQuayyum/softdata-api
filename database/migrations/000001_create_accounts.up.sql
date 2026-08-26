-- Migration: create accounts table

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username text NOT NULL,
    email text NULL,
    password_hash text NOT NULL,
    status text NOT NULL DEFAULT 'active',
    email_verified_at timestamptz NULL,
    last_login_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,
    CONSTRAINT accounts_username_not_blank CHECK (char_length(trim(username)) > 0),
    CONSTRAINT accounts_status_check CHECK (status IN ('active', 'suspended', 'deleted'))
);

CREATE UNIQUE INDEX IF NOT EXISTS accounts_username_unique
    ON accounts (lower(username))
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS accounts_email_unique
    ON accounts (lower(email))
    WHERE email IS NOT NULL
      AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS accounts_status_idx
    ON accounts (status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS accounts_created_at_idx
    ON accounts (created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS accounts_created_at_id_idx
    ON accounts (created_at DESC, id DESC)
    WHERE deleted_at IS NULL;
