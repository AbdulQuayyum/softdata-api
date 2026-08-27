-- Migration: add dataset_group to api_requests

ALTER TABLE api_requests
    ADD COLUMN IF NOT EXISTS dataset_group text NULL;

ALTER TABLE api_requests
    ADD CONSTRAINT api_requests_dataset_group_check
    CHECK (
        dataset_group IS NULL OR dataset_group IN (
            'geography',
            'finance',
            'education',
            'healthcare',
            'emergency',
            'infrastructure',
            'statistics'
        )
    );

CREATE INDEX IF NOT EXISTS api_requests_account_id_created_at_dataset_group_idx
    ON api_requests (account_id, created_at DESC, dataset_group)
    WHERE dataset_group IS NOT NULL;

CREATE INDEX IF NOT EXISTS api_requests_api_key_id_created_at_dataset_group_idx
    ON api_requests (api_key_id, created_at DESC, dataset_group)
    WHERE dataset_group IS NOT NULL;
