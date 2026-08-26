-- Migration: create dataset_versions table

CREATE TABLE IF NOT EXISTS dataset_versions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id uuid NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    version text NOT NULL,
    schema_version text NULL,
    format text NOT NULL,
    status text NOT NULL DEFAULT 'draft',
    record_count integer NOT NULL DEFAULT 0,
    checksum text NULL,
    storage_path text NULL,
    notes text NULL,
    released_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT dataset_versions_version_not_blank CHECK (char_length(trim(version)) > 0),
    CONSTRAINT dataset_versions_format_not_blank CHECK (char_length(trim(format)) > 0),
    CONSTRAINT dataset_versions_status_check CHECK (status IN ('draft', 'published', 'deprecated', 'archived')),
    CONSTRAINT dataset_versions_record_count_check CHECK (record_count >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS dataset_versions_dataset_version_format_unique
    ON dataset_versions (dataset_id, version, lower(format));

CREATE INDEX IF NOT EXISTS dataset_versions_dataset_id_idx
    ON dataset_versions (dataset_id);

CREATE INDEX IF NOT EXISTS dataset_versions_status_idx
    ON dataset_versions (status);

CREATE INDEX IF NOT EXISTS dataset_versions_released_at_idx
    ON dataset_versions (released_at DESC);
