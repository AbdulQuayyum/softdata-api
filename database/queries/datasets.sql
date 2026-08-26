-- Dataset queries

-- name: CreateDataset :one
INSERT INTO datasets (
    dataset_key,
    slug,
    name,
    description,
    group_name,
    country_code,
    version,
    status,
    record_count,
    primary_format,
    formats,
    schema_path,
    licence_id,
    source_count,
    update_frequency,
    last_updated_at,
    last_verified_at,
    maintainers,
    is_public,
    archived_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    COALESCE($8, 'draft'),
    COALESCE($9, 0),
    COALESCE($10, 'json'),
    COALESCE($11, ARRAY['json']::text[]),
    $12,
    $13,
    COALESCE($14, 0),
    $15,
    $16,
    $17,
    COALESCE($18, '{}'::text[]),
    COALESCE($19, true),
    $20
)
RETURNING *;

-- name: GetDatasetByID :one
SELECT *
FROM datasets
WHERE id = $1;

-- name: GetDatasetByKey :one
SELECT *
FROM datasets
WHERE lower(dataset_key) = lower($1);

-- name: GetDatasetBySlug :one
SELECT *
FROM datasets
WHERE lower(slug) = lower($1);

-- name: ListPublicDatasets :many
SELECT *
FROM datasets
WHERE is_public = true
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListPublicDatasetsAfter :many
SELECT *
FROM datasets
WHERE is_public = true
  AND (
    created_at < $2
    OR (created_at = $2 AND id < $3)
  )
ORDER BY created_at DESC, id DESC
LIMIT $1;

-- name: ListPublicDatasetsByGroup :many
SELECT *
FROM datasets
WHERE is_public = true
  AND group_name = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPublicDatasetsByGroupAfter :many
SELECT *
FROM datasets
WHERE is_public = true
  AND group_name = $1
  AND (
    created_at < $3
    OR (created_at = $3 AND id < $4)
  )
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: ListPublicDatasetsByStatus :many
SELECT *
FROM datasets
WHERE is_public = true
  AND status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPublicDatasetsByStatusAfter :many
SELECT *
FROM datasets
WHERE is_public = true
  AND status = $1
  AND (
    created_at < $3
    OR (created_at = $3 AND id < $4)
  )
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: ListPublicDatasetsByGroupAndStatus :many
SELECT *
FROM datasets
WHERE is_public = true
  AND group_name = $1
  AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListPublicDatasetsByGroupAndStatusAfter :many
SELECT *
FROM datasets
WHERE is_public = true
  AND group_name = $1
  AND status = $2
  AND (
    created_at < $4
    OR (created_at = $4 AND id < $5)
  )
ORDER BY created_at DESC, id DESC
LIMIT $3;

-- name: ListAllDatasets :many
SELECT *
FROM datasets
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAllDatasetsAfter :many
SELECT *
FROM datasets
WHERE (
    created_at < $2
    OR (created_at = $2 AND id < $3)
)
ORDER BY created_at DESC, id DESC
LIMIT $1;

-- name: UpdateDatasetMetadata :one
UPDATE datasets
SET
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    group_name = COALESCE($4, group_name),
    country_code = COALESCE($5, country_code),
    version = COALESCE($6, version),
    status = COALESCE($7, status),
    record_count = COALESCE($8, record_count),
    primary_format = COALESCE($9, primary_format),
    formats = COALESCE($10, formats),
    schema_path = COALESCE($11, schema_path),
    licence_id = COALESCE($12, licence_id),
    source_count = COALESCE($13, source_count),
    update_frequency = COALESCE($14, update_frequency),
    last_updated_at = COALESCE($15, last_updated_at),
    last_verified_at = COALESCE($16, last_verified_at),
    maintainers = COALESCE($17, maintainers),
    is_public = COALESCE($18, is_public),
    archived_at = $19,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ArchiveDataset :one
UPDATE datasets
SET
    status = 'archived',
    archived_at = now(),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateDatasetSource :one
INSERT INTO dataset_sources (
    dataset_id,
    source_key,
    name,
    url,
    description,
    publisher,
    source_type,
    licence_id,
    is_official,
    last_fetched_at,
    last_verified_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    COALESCE($9, false),
    $10,
    $11
)
RETURNING *;

-- name: ListDatasetSources :many
SELECT *
FROM dataset_sources
WHERE dataset_id = $1
ORDER BY created_at ASC;

-- name: GetDatasetSourceByID :one
SELECT *
FROM dataset_sources
WHERE id = $1;

-- name: UpdateDatasetSource :one
UPDATE dataset_sources
SET
    name = COALESCE($2, name),
    url = COALESCE($3, url),
    description = COALESCE($4, description),
    publisher = COALESCE($5, publisher),
    source_type = COALESCE($6, source_type),
    licence_id = COALESCE($7, licence_id),
    is_official = COALESCE($8, is_official),
    last_fetched_at = COALESCE($9, last_fetched_at),
    last_verified_at = COALESCE($10, last_verified_at),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDatasetSource :exec
DELETE FROM dataset_sources
WHERE id = $1;

-- name: CreateDatasetVersion :one
INSERT INTO dataset_versions (
    dataset_id,
    version,
    schema_version,
    format,
    status,
    record_count,
    checksum,
    storage_path,
    notes,
    released_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    COALESCE($5, 'draft'),
    COALESCE($6, 0),
    $7,
    $8,
    $9,
    $10
)
RETURNING *;

-- name: ListDatasetVersions :many
SELECT *
FROM dataset_versions
WHERE dataset_id = $1
ORDER BY released_at DESC NULLS LAST, created_at DESC;

-- name: GetDatasetVersionByID :one
SELECT *
FROM dataset_versions
WHERE id = $1;

-- name: GetDatasetVersionByDatasetAndVersion :one
SELECT *
FROM dataset_versions
WHERE dataset_id = $1
  AND version = $2
  AND lower(format) = lower($3);

-- name: UpdateDatasetVersion :one
UPDATE dataset_versions
SET
    schema_version = COALESCE($2, schema_version),
    format = COALESCE($3, format),
    status = COALESCE($4, status),
    record_count = COALESCE($5, record_count),
    checksum = COALESCE($6, checksum),
    storage_path = COALESCE($7, storage_path),
    notes = COALESCE($8, notes),
    released_at = COALESCE($9, released_at),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: PublishDatasetVersion :one
UPDATE dataset_versions
SET
    status = 'published',
    released_at = COALESCE(released_at, now()),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDatasetVersion :exec
DELETE FROM dataset_versions
WHERE id = $1;
