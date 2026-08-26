-- Migration: create dataset_sources table

CREATE TABLE IF NOT EXISTS dataset_sources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id uuid NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    source_key text NOT NULL,
    name text NOT NULL,
    url text NULL,
    description text NULL,
    publisher text NULL,
    source_type text NULL,
    licence_id text NULL,
    is_official boolean NOT NULL DEFAULT false,
    last_fetched_at timestamptz NULL,
    last_verified_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT dataset_sources_source_key_not_blank CHECK (char_length(trim(source_key)) > 0),
    CONSTRAINT dataset_sources_name_not_blank CHECK (char_length(trim(name)) > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS dataset_sources_dataset_id_source_key_unique
    ON dataset_sources (dataset_id, lower(source_key));

CREATE INDEX IF NOT EXISTS dataset_sources_dataset_id_idx
    ON dataset_sources (dataset_id);

CREATE INDEX IF NOT EXISTS dataset_sources_is_official_idx
    ON dataset_sources (is_official);
