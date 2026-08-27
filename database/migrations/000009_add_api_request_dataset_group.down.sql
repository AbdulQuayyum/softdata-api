-- Migration: remove dataset_group from api_requests

DROP INDEX IF EXISTS api_requests_api_key_id_created_at_dataset_group_idx;
DROP INDEX IF EXISTS api_requests_account_id_created_at_dataset_group_idx;

ALTER TABLE api_requests
    DROP CONSTRAINT IF EXISTS api_requests_dataset_group_check;

ALTER TABLE api_requests
    DROP COLUMN IF EXISTS dataset_group;
