# SoftData Datasets

This document defines how SoftData datasets are organized, sourced, validated and versioned.

## Dataset Groups

### Geography

Location and administrative reference data.

Examples:

- Countries
- States
- Local government areas
- Wards
- Cities
- Geopolitical zones
- Postal codes
- Vehicle plate codes
- GeoJSON boundaries

### Finance

Financial institution and payment reference data.

Examples:

- Financial institutions
- Bank codes
- USSD codes
- Payment channels
- Currencies

### Education

Educational institution data.

Examples:

- Universities
- Polytechnics
- Colleges of education
- Courses
- Institution ownership
- Accreditation information

### Healthcare

Health facility reference data.

Examples:

- Hospitals
- Primary healthcare centres
- Clinics
- Laboratories
- Facility services
- Facility ownership

### Emergency

Emergency and public-safety service data.

Examples:

- Emergency telephone numbers
- Police commands
- Fire stations
- Road-safety commands

### Infrastructure

Physical and public-service infrastructure.

Examples:

- Airports
- Seaports
- Railway stations
- Electricity distribution companies
- Telecommunications prefixes

### Statistics

Time-based statistical observations.

Examples:

- Population
- Inflation
- Food prices
- Fuel prices
- Economic indicators

## Published Datasets

The published geography dataset packages currently include:

### `world-countries-and-areas`

The current English United Nations M49 table of 248 countries or areas.

- `datasets/geography/countries_and_areas.json`
- `datasets/schemas/geography/countries_and_areas.schema.json`
- `datasets/metadata/geography/countries_and_areas.json`

Each record uses the lowercase alpha-2 code as the public `id`, preserves the source English name, keeps `alpha_2_code`, `alpha_3_code` and `numeric_code` as published, and includes `calling_codes`, `flag_emoji`, `flag_svg_url` and region hierarchy fields when the source table provides them. The package follows the UN statistical current-manifest boundary, keeps territories and other areas where they appear in the official table, and states in metadata that the designations are statistical references only and do not imply political recognition or legal status. SoftData's independent compilation, schema and metadata are CC BY 4.0, while the UN source material retains its own rights. Flag SVG assets are vendored separately from MIT-licensed flag-icons v7.5.0.

This global package is separate from the Nigerian geography datasets, which continue to use `country_code: NG`.

### `world-currencies`

The current ISO 4217 monetary-currency snapshot compiled from the official SIX List One XML.

- `datasets/finance/currencies.json`
- `datasets/schemas/finance/currencies.schema.json`
- `datasets/metadata/finance/currencies.json`

The package contains 155 current monetary currencies. Each record uses the lowercase alphabetic code as its public `id`, preserves the source currency name, alphabetic code, numeric code and minor unit, and maps approved world country/area IDs through `country_area_ids`. The package uses the current monetary-currency boundary only, excludes special-purpose, historical and no-currency codes, keeps TWD with an empty `country_area_ids` array, and omits reverse mappings for Antarctica, State of Palestine and South Georgia and the South Sandwich Islands. SoftData's independent compilation, schema and metadata are CC BY 4.0, while the ISO and SIX source publications retain their own rights.

### `ng-states`

Nigeria's 36 states and the Federal Capital Territory.

- `datasets/geography/states.json`
- `datasets/schemas/geography/states.schema.json`
- `datasets/metadata/geography/states.json`

Each state record includes a `geopolitical_zone_id` that links back to the zone catalogue below.

### `ng-geopolitical-zones`

Nigeria's six geopolitical zones.

- `datasets/geography/geopolitical_zones.json`
- `datasets/schemas/geography/geopolitical_zones.schema.json`
- `datasets/metadata/geography/geopolitical_zones.json`

### `ng-lgas`

Nigeria's 768 Local Government Areas and the Federal Capital Territory's six Area Councils.

- `datasets/geography/lgas.json`
- `datasets/schemas/geography/lgas.schema.json`
- `datasets/metadata/geography/lgas.json`

Each unit record links back to `state_id`, and geopolitical zone membership is derived through the state dataset.

### `ng-payment-service-providers`

Nigeria's current Central Bank of Nigeria payment-service-provider register snapshot.

- `datasets/finance/payment_service_providers.json`
- `datasets/schemas/finance/payment_service_providers.schema.json`
- `datasets/metadata/finance/payment_service_providers.json`

The package contains 255 provider-category memberships across seven approved PSP categories. Each record represents one provider-category membership, and a provider may appear in multiple categories.

### `ng-international-money-transfer-operators`

Nigeria's current Central Bank of Nigeria IMTO register snapshot.

- `datasets/finance/international_money_transfer_operators.json`
- `datasets/schemas/finance/international_money_transfer_operators.schema.json`
- `datasets/metadata/finance/international_money_transfer_operators.json`

The package contains 108 CBN-listed IMTO entries. Each record represents one current register listing, with the source-side concatenation defect at SN 63 normalized transparently in metadata. The package is names-only, does not record addresses or inferred country data, and IMTOs remain separate from the payment-service-provider register. No API routes are introduced by this dataset package.

### `ng-universities`

Nigeria's current National Universities Commission register of federal, state and private universities.

- `datasets/education/universities.json`
- `datasets/schemas/education/universities.schema.json`
- `datasets/metadata/education/universities.json`

The package contains 328 university records across the current NUC federal, state and private registers. Each record represents one university listing, with `state_id` linking the record to `ng-states` and `ownership_type` preserving the category published by the NUC.

### `ng-colleges-of-education`

Nigeria's current National Commission for Colleges of Education register of active colleges of education.

- `datasets/education/colleges_of_education.json`
- `datasets/schemas/education/colleges_of_education.schema.json`
- `datasets/metadata/education/colleges_of_education.json`

The package version is `1.0.0` and contains 244 college records across the current NCCE federal, state and private categories (`28` federal, `48` state, `168` private). Each record represents one active college listing, with `state_id` linking the record to `ng-states` and `ownership_type` preserving the category published by the NCCE. The stale `Cross River State Coll. of Education, Akampa` row is excluded because current Cross River State Government evidence describes the successor as a university that is already represented in `ng-universities`. SoftData's independent compilation, schema and metadata are CC BY 4.0; the NCCE and other official publications retain their own rights.

## Directory Structure

```text
datasets/
├── geography/
├── finance/
├── education/
├── healthcare/
├── emergency/
├── infrastructure/
├── statistics/
├── schemas/
└── metadata/
```

## Dataset Metadata

Every dataset must have a record in:

```text
datasets/metadata/datasets.json
```

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

A dataset can have one of these statuses:

```text
draft
review
active
deprecated
archived
```

- `draft`: Still being assembled.
- `review`: Awaiting validation or source review.
- `active`: Available through the public API.
- `deprecated`: Still available but scheduled for replacement.
- `archived`: No longer served by the current API.

## Stable Identifiers

Every record should have a stable identifier where possible.

Identifiers should not depend on a record’s array position.

Good:

```json
{
  "id": "NG-KW",
  "name": "Kwara"
}
```

Bad:

```json
{
  "id": 14,
  "name": "Kwara"
}
```

Use recognized standards when they exist.

## Dates

Use ISO 8601:

```text
2026-08-26
2026-08-26T14:30:00Z
```

Do not use ambiguous formats such as:

```text
08/09/26
```

## Missing Values

Use `null` when a value is unknown or unavailable:

```json
{
  "website": null
}
```

Do not use:

```json
{
  "website": "N/A"
}
```

An empty array should be used when a collection has no known entries:

```json
{
  "services": []
}
```

## Source Requirements

Every active dataset must have at least one documented source.

Source records belong in:

```text
datasets/metadata/sources.json
```

Example:

```json
{
  "id": "source-example",
  "name": "Official source name",
  "publisher": "Publishing organization",
  "url": "https://example.gov.ng",
  "accessed_at": "2026-08-26",
  "notes": null
}
```

## Licence Requirements

Licence records belong in:

```text
datasets/metadata/licences.json
```

Example:

```json
{
  "id": "licence-example",
  "name": "Open Data Licence",
  "url": "https://example.gov.ng/licence",
  "attribution_required": true,
  "commercial_use_allowed": true,
  "redistribution_allowed": true,
  "notes": null
}
```

Do not publish a dataset when redistribution permission is unknown.

## Versioning

Datasets use semantic versions:

```text
MAJOR.MINOR.PATCH
```

Examples:

```text
1.0.0
1.1.0
1.1.1
2.0.0
```

- `MAJOR`: Schema changes or removed fields
- `MINOR`: New fields or significant new records
- `PATCH`: Corrections that do not change the schema

## Validation

Before publication, a dataset must pass:

- JSON or CSV parsing
- Schema validation
- Required-field validation
- Unique-identifier validation
- Duplicate-record validation
- Referential-integrity validation
- Date-format validation
- Coordinate-range validation where applicable
- Source and licence validation

## Dataset Downloads

Downloadable formats may include:

```text
JSON
CSV
GeoJSON
```

Example:

```http
GET /v1/datasets/ng-states/download?format=json
```

## Corrections

Corrections must include:

- A description of the error
- The affected record
- The proposed correction
- A reliable supporting source
- The correction date
- The contributor

Corrections should not silently overwrite dataset history.

## Deprecation

When a dataset or schema is deprecated:

1. Mark it as deprecated in metadata.
2. Document the replacement.
3. Add response deprecation headers where applicable.
4. Keep it available during a reasonable migration period.
5. Record its removal in the changelog.
