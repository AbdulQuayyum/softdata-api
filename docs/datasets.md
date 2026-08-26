# Datasets

SoftData organizes public data into a small set of dataset groups with consistent metadata, versioning, and source tracking.

## Dataset Groups

- Geography
- Finance
- Education
- Healthcare
- Emergency
- Infrastructure
- Statistics

## Dataset Principles

- Every dataset should have a stable identifier.
- Identifiers should not depend on array position.
- Records should use recognized standards where possible.
- Missing values should be represented with `null`.
- Dates should use ISO 8601 formats.

## Dataset Metadata

Each dataset should include metadata such as:

- `id`
- `name`
- `description`
- `group`
- `country_code`
- `formats`
- `version`
- `record_count`
- `schema`
- `source_ids`
- `licence_id`
- `update_frequency`
- `last_updated_at`
- `last_verified_at`
- `status`
- `maintainers`

Example:

```json
{
  "id": "ng-states",
  "name": "Nigerian States",
  "description": "States and the Federal Capital Territory of Nigeria.",
  "group": "geography",
  "country_code": "NG",
  "formats": ["json", "csv"],
  "version": "1.0.0",
  "record_count": 37,
  "schema": "geography.schema.json",
  "source_ids": ["source-example"],
  "licence_id": "licence-example",
  "update_frequency": "yearly",
  "last_updated_at": "2026-08-26",
  "last_verified_at": "2026-08-26",
  "status": "active",
  "maintainers": ["Abdul-Quayyum Alao"]
}
```

## Dataset Status

- `draft`: Data is still being assembled.
- `review`: Data is awaiting validation or source review.
- `active`: Data is available through the public API.
- `deprecated`: Data is still available but scheduled for replacement.
- `archived`: Data is no longer served by the current API.

## Formats

Datasets may be published in:

- JSON
- CSV
- GeoJSON

Some datasets may support more than one format depending on the source data and intended use.

## Sources

Each dataset should reference one or more source records. Source metadata should make it clear:

- where the data came from
- when it was last verified
- what licence applies
- whether the source is official, derived, or manually curated

## Versioning

Dataset versions should change when:

- the source data changes
- fields are added or removed
- records are corrected
- the schema changes

Version numbers should be predictable and human-readable, typically following semantic versioning.

## Validation

Before a dataset is promoted to `active`, it should be checked for:

- schema validity
- duplicate or malformed identifiers
- missing required fields
- date and type consistency
- source traceability

## Contribution Notes

When adding or updating datasets:

- keep records normalized
- prefer official sources
- document transformations clearly
- preserve historical versions when possible
- update the dataset metadata alongside the data files
