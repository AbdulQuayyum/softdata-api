-- Migration: create datasets table

CREATE TABLE IF NOT EXISTS datasets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_key text NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    description text NULL,
    group_name text NOT NULL,
    country_code char(2) NULL,
    version text NOT NULL,
    status text NOT NULL DEFAULT 'draft',
    record_count integer NOT NULL DEFAULT 0,
    primary_format text NOT NULL DEFAULT 'json',
    formats text[] NOT NULL DEFAULT ARRAY['json']::text[],
    schema_path text NULL,
    licence_id text NULL,
    source_count integer NOT NULL DEFAULT 0,
    update_frequency text NULL,
    last_updated_at date NULL,
    last_verified_at date NULL,
    maintainers text[] NOT NULL DEFAULT '{}'::text[],
    is_public boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    archived_at timestamptz NULL,
    CONSTRAINT datasets_dataset_key_not_blank CHECK (char_length(trim(dataset_key)) > 0),
    CONSTRAINT datasets_slug_not_blank CHECK (char_length(trim(slug)) > 0),
    CONSTRAINT datasets_name_not_blank CHECK (char_length(trim(name)) > 0),
    CONSTRAINT datasets_group_not_blank CHECK (char_length(trim(group_name)) > 0),
    CONSTRAINT datasets_country_code_length CHECK (country_code IS NULL OR char_length(country_code) = 2),
    CONSTRAINT datasets_record_count_check CHECK (record_count >= 0),
    CONSTRAINT datasets_source_count_check CHECK (source_count >= 0),
    CONSTRAINT datasets_status_check CHECK (status IN ('draft', 'review', 'active', 'deprecated', 'archived'))
);

CREATE UNIQUE INDEX IF NOT EXISTS datasets_dataset_key_unique
    ON datasets (lower(dataset_key));

CREATE UNIQUE INDEX IF NOT EXISTS datasets_slug_unique
    ON datasets (lower(slug));

CREATE INDEX IF NOT EXISTS datasets_group_name_idx
    ON datasets (group_name);

CREATE INDEX IF NOT EXISTS datasets_status_idx
    ON datasets (status);

CREATE INDEX IF NOT EXISTS datasets_country_code_idx
    ON datasets (country_code);
