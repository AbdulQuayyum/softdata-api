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

## Published Datasets

The published geography packages include:

### `world-countries-and-areas`

The current English United Nations M49 table of 248 countries or areas.

- Data: `datasets/geography/countries_and_areas.json`
- Schema: `datasets/schemas/geography/countries_and_areas.schema.json`
- Metadata: `datasets/metadata/geography/countries_and_areas.json`

Each record uses the lowercase alpha-2 code as its public `id`, preserves the source English name and published codes, and includes `calling_codes`, `flag_emoji`, `flag_svg_url` and region hierarchy fields when the source table provides them. The package follows the current UN statistical manifest boundary, keeps territories and other areas where they appear in the official table, and states in metadata that the designations are statistical references only and do not imply political recognition or legal status. SoftData's independent compilation, schema and metadata are CC BY 4.0, while the UN source material retains its own rights. Flag SVG assets are vendored separately from MIT-licensed flag-icons v7.5.0.

This global package is separate from the Nigerian geography datasets, which continue to use `country_code: NG`.

### `ng-states`

Nigeria's 36 states and the Federal Capital Territory.

- Data: `datasets/geography/states.json`
- Schema: `datasets/schemas/geography/states.schema.json`
- Metadata: `datasets/metadata/geography/states.json`

Each state record includes `geopolitical_zone_id`, which links to the zone catalogue below.

### `ng-geopolitical-zones`

Nigeria's six geopolitical zones.

- Data: `datasets/geography/geopolitical_zones.json`
- Schema: `datasets/schemas/geography/geopolitical_zones.schema.json`
- Metadata: `datasets/metadata/geography/geopolitical_zones.json`

### `ng-lgas`

Nigeria's 768 Local Government Areas and the Federal Capital Territory's six Area Councils.

- Data: `datasets/geography/lgas.json`
- Schema: `datasets/schemas/geography/lgas.schema.json`
- Metadata: `datasets/metadata/geography/lgas.json`

Each unit links to `state_id`, and geopolitical-zone membership is derived from the state dataset.

### `ng-payment-service-providers`

Nigeria's current Central Bank of Nigeria payment-service-provider register snapshot.

- Data: `datasets/finance/payment_service_providers.json`
- Schema: `datasets/schemas/finance/payment_service_providers.schema.json`
- Metadata: `datasets/metadata/finance/payment_service_providers.json`

The package contains 255 provider-category memberships across seven approved PSP categories. Each record is one provider-category membership, and a provider may appear in multiple categories.

### `ng-international-money-transfer-operators`

Nigeria's current Central Bank of Nigeria IMTO register snapshot.

- Data: `datasets/finance/international_money_transfer_operators.json`
- Schema: `datasets/schemas/finance/international_money_transfer_operators.schema.json`
- Metadata: `datasets/metadata/finance/international_money_transfer_operators.json`

The package contains 108 CBN-listed IMTO entries. Each record represents one current register listing, with the source-side concatenation defect at SN 63 normalized transparently in metadata. The package is names-only, does not record addresses or inferred country data, and IMTOs remain separate from the payment-service-provider register. No API routes are introduced by this dataset package.

### `ng-universities`

Nigeria's current National Universities Commission register of federal, state and private universities.

- Data: `datasets/education/universities.json`
- Schema: `datasets/schemas/education/universities.schema.json`
- Metadata: `datasets/metadata/education/universities.json`

The package contains 328 university records across the current NUC federal, state and private registers. Each record represents one university listing, with `state_id` linking the record to `ng-states` and `ownership_type` preserving the category published by the NUC.

### `ng-colleges-of-education`

Nigeria's current National Commission for Colleges of Education register of active colleges of education.

- Data: `datasets/education/colleges_of_education.json`
- Schema: `datasets/schemas/education/colleges_of_education.schema.json`
- Metadata: `datasets/metadata/education/colleges_of_education.json`

The package version is `1.0.0` and contains 244 college records across the current NCCE federal, state and private categories (`28` federal, `48` state, `168` private). Each record represents one active college listing, with `state_id` linking the record to `ng-states` and `ownership_type` preserving the category published by the NCCE. The stale `Cross River State Coll. of Education, Akampa` row is excluded because current Cross River State Government evidence describes the successor as a university that is already represented in `ng-universities`. SoftData's independent compilation, schema and metadata are CC BY 4.0; the NCCE and other official publications retain their own rights.

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
