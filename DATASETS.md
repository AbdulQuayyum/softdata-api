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

### `ng-health-facilities`

Unified Nigerian health-facility snapshot derived from the public GRID3 Nigeria health-facilities layer, which documents 2024 NHFR inputs and NPHCDA standards.

- `datasets/healthcare/health_facilities.json`
- `datasets/schemas/healthcare/health_facilities.schema.json`
- `datasets/metadata/healthcare/health_facilities.json`
- `datasets/metadata/healthcare/health_facilities_reconciliation/index.json` and its 37 state/FCT partitions

The package contains 50,649 retained records from 51,022 source rows, covering all 36 states and the FCT and 768 canonical LGAs. It is a dated observed registry snapshot, not a continuously current active-facility register. Ten observations in five unresolved identity pairs and 363 unresolved geography rows are excluded and recorded in reconciliation metadata. There are no duplicate merges. See the [identity review](docs/healthcare-identity-resolution.md). No operational-status claim, website, logo or inferred coordinate is published. Source ownership remains with GRID3, CIESIN and the credited contributors; SoftData claims only its independent normalization, schema, identifiers, reconciliation and metadata work.

### `ng-medical-laboratory-accreditations`

Nigeria Medical Laboratory Accreditation Register Snapshot.

- `datasets/healthcare/medical_laboratory_accreditations.json`
- `datasets/schemas/healthcare/medical_laboratory_accreditations.schema.json`
- `datasets/metadata/healthcare/medical_laboratory_accreditations.json`
- `datasets/metadata/healthcare/medical_laboratory_accreditations_reconciliation/index.json` and its 11 state/FCT partitions

A dated snapshot of medical laboratory facility accreditation records published by the MLSCN Accreditation Service. It is not a complete register of all licensed medical laboratory premises in Nigeria. The package contains 30 facility accreditation records from 30 source observations across 10 states and the FCT. Four entries had certificate expiry dates before the 2026-09-09 retrieval date and remain distinguishable as `expired`; the other 26 are marked `accredited`. The source does not publish reliable LGA, ownership, premises-registration number, renewal-year or laboratory-type fields, so those fields are absent from the active contract. No individual medical laboratory scientists, personal registration numbers, personal phone numbers or personal email addresses are published. Runtime repository, service, route, OpenAPI and startup integration are intentionally deferred.

### `ng-licensed-pharmacies` — deferred pending sanitized PCN premises data

This future pharmacy-premises dataset remains deferred because the available PCN premises source requires a privacy-safe, authoritative row-level publication before it can be added. It is separate from MLSCN medical laboratory accreditation regulation.

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

Each record uses the lowercase alpha-2 code as the public `id`, preserves the source English name, keeps `alpha_2_code`, `alpha_3_code` and `numeric_code` as published, and includes `calling_codes`, `flag_emoji`, `flag_svg_url`, `flag_url` and region hierarchy fields when the source table provides them. `flag_url` is qualified with `PUBLIC_API_URL` in API responses when configured. The package follows the UN statistical current-manifest boundary, keeps territories and other areas where they appear in the official table, and states in metadata that the designations are statistical references only and do not imply political recognition or legal status. SoftData's independent compilation, schema and metadata are CC BY 4.0, while the UN source material retains its own rights. Flag SVG assets are vendored separately from MIT-licensed flag-icons v7.5.0.

This global package is separate from the Nigerian geography datasets, which continue to use `country_code: NG`.

### `world-time-zones`

The canonical IANA tzdb `zone1970.tab` manifest from release `2026c`, compiled into 312 time-zone records.

- `datasets/geography/time_zones.json`
- `datasets/schemas/geography/time_zones.schema.json`
- `datasets/metadata/geography/time_zones.json`

Each record preserves the exact IANA identifier as its public `id` and maps approved `world-countries-and-areas` IDs through `country_area_ids`. The package excludes aliases, `zone.tab`-only records, `backzone`, fixed-offset identifiers and other non-canonical entries, keeps `Asia/Taipei` with an empty `country_area_ids` array, records `bv` and `hm` as the two M49 country/area IDs with zero canonical zones, and omits static offsets, DST flags and coordinates. SoftData's independent compilation, schema and metadata are CC BY 4.0, while the IANA source files retain their own rights and are cited for reference only.

### `world-languages` and `world-country-languages`

These paired geography datasets are compiled from Unicode CLDR JSON release 48.2.0, accessed 2026-09-04. `world-languages` contains 633 current English CLDR base-language identifiers and names. `world-country-languages` contains 1,289 CLDR-derived country/area relationships covering all 248 approved country-area IDs, with 523 referenced language IDs and 110 catalog languages intentionally having no approved relationship. Relationship statuses are `official`, `de_facto_official`, `official_regional` and SoftData's normalized `used` value for rows without an `officialStatus`; this is not a legal or exhaustive census claim.

- Data: `datasets/geography/languages.json`, `datasets/geography/country_languages.json`
- Schema: `datasets/schemas/geography/languages.schema.json`, `datasets/schemas/geography/country_languages.schema.json`
- Metadata: `datasets/metadata/geography/languages.json`, `datasets/metadata/geography/country_languages.json`

Alias normalization applies `fat -> ak`, `sh -> sr-Latn -> sr`, `tl -> fil` and `tw -> ak`. Script-qualified rows collapse to base IDs, and an exact base-language row wins status conflicts over a script-qualified row. The package excludes regional aggregates, historic territories, Taiwan and Kosovo. SoftData's independent compilation, schema and metadata are CC BY 4.0; Unicode CLDR source material retains its own licence and attribution.

### Derived Country Profiles

`GET /v1/geography/countries/{country_id}/profile` is a derived view rather than a static `world-country-profiles` dataset. It combines the approved country record with relationships from `world-currencies`, `world-time-zones`, and `world-country-languages`; its `language_ids` values reference `world-languages.id`. Language IDs are unique, deterministically sorted, and include all CLDR-derived relationship statuses. Presence does not mean that every language is nationally official, and an empty array is valid. Use `GET /v1/geography/country-languages?country_area_id={country_id}` for status details.

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
- `datasets/metadata/finance/payment_service_providers_reconciliation.json`
- `datasets/metadata/finance/payment_service_provider_mobile_money_logo_reconciliation.json`
- `datasets/metadata/finance/cross_dataset_finance_logo_phase2b_reconciliation.json`
- `datasets/metadata/finance/cross_dataset_finance_logo_phase2d_reconciliation.json`
- `datasets/metadata/finance/cross_dataset_finance_logo_reconciliation.json`
- `datasets/assets/financial-institutions/ng/payment-service-providers/`

The package contains 255 provider-category memberships across seven approved PSP categories. Each record represents one provider-category membership, and a provider may appear in multiple categories. The optional enrichment contains 11 `cbn_code` values, 16 `nip_code` values, 109 verified `website_url` values and 23 curated PNG assets. `cbn_code` represents the three-digit legacy CBN/sort-code field (`scCode`); `nip_code` represents the six-digit NIP institution-code field (`bankCode`). Most values are from the pinned community catalogue, Paystack is supporting evidence only, and unresolved values are omitted pending possible CBN/NIBSS confirmation. Phase 2A reviewed all 55 provisional unmatched mobile-money assets and records the accepted, preserved and rejected decisions in the dedicated reconciliation manifest.

### `ng-international-money-transfer-operators`

Nigeria's current Central Bank of Nigeria IMTO register snapshot.

- `datasets/finance/international_money_transfer_operators.json`
- `datasets/schemas/finance/international_money_transfer_operators.schema.json`
- `datasets/metadata/finance/international_money_transfer_operators.json`

The package contains 108 CBN-listed IMTO entries. Each record represents one current register listing, with the source-side concatenation defect at SN 63 normalized transparently in metadata. Optional HTTPS website URLs and verified logo URLs are included where the global or Nigerian legal-entity relationship and asset provenance were established. IMTOs remain separate from the payment-service-provider register.

### `ng-commercial-banks`

The 28-record commercial-bank snapshot from the supplied CBN `Export (1).xlsx` workbook.

- Data: `datasets/finance/commercial_banks.json`
- Schema: `datasets/schemas/finance/commercial_banks.schema.json`
- Metadata: `datasets/metadata/finance/commercial_banks.json`
- Logos: `datasets/assets/banks/ng/`

Each record uses a deterministic lowercase bank ID, `country_code` `NG`, an official website URL and a required local logo URL. `cbn_code` is the three-digit legacy CBN/sort-code field represented by source `scCode`; `nip_code` is the six-digit NIP institution-code field represented by source `bankCode`. Most enrichment values come from the pinned community catalogue, with Paystack used only as supporting evidence; community mappings may require future confirmation against a current CBN/NIBSS register. These fields are optional in this snapshot; omitted values were not sufficiently verified and are not inferred. Logos are immutable vendored PNG assets with source URLs, hashes and provenance recorded in the attribution files. Bank names and logos remain the property and trademarks of their respective owners; SoftData provides them for identification and directory purposes only. Rights holders may request correction, replacement or removal of an inaccurate logo without removal of the factual bank record.

### `ng-microfinance-banks`

The roster-only 790-record active Nigerian microfinance-bank dataset from the reconciled CBN/NDIC status snapshot.

- Data: `datasets/finance/microfinance_banks.json`
- Schema: `datasets/schemas/finance/microfinance_banks.schema.json`
- Metadata: `datasets/metadata/finance/microfinance_banks.json`
- Reconciliation manifest: `datasets/metadata/finance/microfinance_banks_reconciliation.json`
- Identifier reconciliation: `datasets/metadata/finance/financial_institution_codes_reconciliation.json`

Records require `id`, `name` and `country_code`; `cbn_code`, `nip_code`, `website_url` and `logo_url` are optional fields. CBN/NIP values are sourced from the pinned catalogue and accepted only when they match the required three- or six-digit semantics; Paystack provider codes are never published. The 829-row CBN snapshot is reduced by 39 evidence-backed exclusions: four duplicate/stale rows, 33 matched revoked institutions, the AKPO predecessor and Verdant-Capital. Website/logo evidence is tracked in `datasets/metadata/finance/microfinance_banks_enrichment.json`, `datasets/metadata/finance/microfinance_banks_phase2c_logo_reconciliation.json` and the category asset attribution files. Phase 2C reviewed all 315 provisional MFB archive assets without changing the 330 accepted-logo coverage. Logo acceptance requires exact identity and byte validation; explicit redistribution permission is not a prerequisite, and logo/trademark rights remain with the respective institutions. Current enrichment coverage is 330 logos for 790 records.

### Nigerian regulated-finance categories

Six additional CBN category snapshots are available as dataset-only packages:

- `ng-non-interest-financial-institutions` (6)
- `ng-merchant-banks` (6)
- `ng-payment-service-banks` (5)
- `ng-financial-holding-companies` (7)
- `ng-development-finance-institutions` (8)
- `ng-primary-mortgage-institutions` (31)

They are sourced from the supplied CBN exports. Category packages use required `id`, `name` and `country_code` fields; CBN/NIP fields are optional only for categories whose public contract includes them, while holding-company and DFI contracts intentionally exclude bank identifiers. Websites and logos are optional and are not inferred. The non-interest source is published under the broader financial-institutions key because its supplied CBN snapshot includes Mint Microfinance Bank. Parent holding companies remain separate from subsidiaries.

Verified PNG assets from official institutional sources and the pinned baseline are embedded under `datasets/assets/financial-institutions/ng/` with per-asset hashes in `ATTRIBUTION.md`. Official-source marks are third-party assets and are not claimed under SoftData CC BY 4.0; the FHA Homes asset is explicitly a Federal Housing Authority parent-brand representative mark. Coverage is complete: non-interest 6/6, merchant 6/6, payment-service 5/5, holding companies 7/7, DFIs 8/8 and primary mortgage 31/31.
The combined decoded coverage is 63/63 websites and 63/63 logos. This comprises 62 independently sourced institution marks and one disclosed Federal Housing Authority parent-brand representative mark for the wholly owned FHA Homes subsidiary. Akwa Savings was reconciled to current Ibom Mortgage Bank while retaining the stable dataset ID; TrustBond Mortgage Bank was removed after the FirstTrust merger was confirmed.

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

### `ng-polytechnics`

- `datasets/education/polytechnics.json`
- `datasets/schemas/education/polytechnics.schema.json`
- `datasets/metadata/education/polytechnics.json`
- `datasets/metadata/education/polytechnics_reconciliation.json`

The static foundation contains 168 current NBTE-listed polytechnics: 35 federal, 43 state and 90 private. It exposes only `id`, `name`, `ownership_type`, `state_id` and `country_code`, with `state_id` resolving to `ng-states`. Colleges, specialist institutes, transition records and other wrong-category entries are excluded and documented. Website/logo enrichment and API integration are intentionally deferred.

### `ng-monotechnics`

- `datasets/education/monotechnics.json`
- `datasets/schemas/education/monotechnics.schema.json`
- `datasets/metadata/education/monotechnics.json`
- `datasets/metadata/education/monotechnics_reconciliation.json`

The static foundation contains 86 NBTE Specialised Institutions (Monotechnics): 32 federal, 4 state and 50 private. It exposes only `id`, `name`, `ownership_type`, `state_id` and `country_code`. Colleges, remedial schools and other wrong-category source entries are excluded and documented. Website/logo enrichment and API integration are intentionally deferred.

### `ng-vocational-enterprise-institutions`

- `datasets/education/vocational_enterprise_institutions.json`
- `datasets/schemas/education/vocational_enterprise_institutions.schema.json`
- `datasets/metadata/education/vocational_enterprise_institutions.json`
- `datasets/metadata/education/vocational_enterprise_institutions_reconciliation.json`

The static foundation contains 25 current Digital NBTE-listed VEIs: 4 state and 21 private across 11 states. It filters the live combined VEI/Skill Training Center/Master Craft Person register to rows explicitly categorized as VEI. IEIs were previously recognized by NBTE, but the category and NID programme were abolished under the newer reform; upgraded and revoked IEIs are not included. This is a roster-only foundation with no websites, logos or API integration.

### `ng-colleges-of-agriculture`

- `datasets/education/colleges_of_agriculture.json`
- `datasets/schemas/education/colleges_of_agriculture.schema.json`
- `datasets/metadata/education/colleges_of_agriculture.json`
- `datasets/metadata/education/colleges_of_agriculture_reconciliation.json`

The static foundation contains 31 NBTE Colleges of Agriculture and Related Disciplines: 23 federal, 7 state and 1 private. It exposes only `id`, `name`, `ownership_type`, `state_id` and `country_code`. The source-marked conversion of College of Agriculture, Zuru is excluded; no universities, polytechnics or monotechnics are merged into this dataset. Website/logo enrichment and API integration are intentionally deferred.

### `ng-primary-and-secondary-schools`

- `datasets/education/primary_and_secondary_schools.json`
- `datasets/schemas/education/primary_and_secondary_schools.schema.json`
- `datasets/metadata/education/primary_and_secondary_schools.json`
- `datasets/metadata/education/primary_and_secondary_schools_reconciliation/index.json` and its 37 state/FCT partitions

This source-observed snapshot contains 166,604 reconciled school/campus records from the official UBEC 2022 Primary and Junior Secondary workbooks: 81,160 public and 85,444 private records across 37 state/FCT values and 772 resolved LGAs. It is not a current active-school, licensing or complete senior-secondary register; 69 ambiguous Osun `ILESHA` rows are explicitly quarantined. Repository, service and HTTP support are deferred.

### `ng-colleges-of-health-sciences-and-technology`

- `datasets/education/colleges_of_health_sciences_and_technology.json`
- `datasets/schemas/education/colleges_of_health_sciences_and_technology.schema.json`
- `datasets/metadata/education/colleges_of_health_sciences_and_technology.json`
- `datasets/metadata/education/colleges_of_health_sciences_and_technology_reconciliation.json`

The static foundation contains 98 NBTE Colleges of Health Sciences and Technology: 4 federal, 31 state and 63 private across 33 states. It exposes only the five roster fields. Hospital-only schools and institutes are excluded, and nursing and midwifery institutions are reserved for a separate future dataset. Website/logo enrichment and API integration are intentionally deferred.

### `ng-colleges-of-nursing-and-midwifery`

- `datasets/education/colleges_of_nursing_and_midwifery.json`
- `datasets/schemas/education/colleges_of_nursing_and_midwifery.schema.json`
- `datasets/metadata/education/colleges_of_nursing_and_midwifery.json`
- `datasets/metadata/education/colleges_of_nursing_and_midwifery_reconciliation.json`

The active package contains 156 institution/campus records: 13 federal, 52 state and 91 private across 32 states. It is the NMCN December 2025 approved-schools snapshot, not a complete live register. The reconciliation preserves 376 extracted numbered-row candidates and explicitly excludes 220 unresolved rows; NMCN separately reports 290 training institutions. Websites, logos and API integration remain intentionally deferred.

### `ng-technical-colleges`

- `datasets/education/technical_colleges.json`
- `datasets/schemas/education/technical_colleges.schema.json`
- `datasets/metadata/education/technical_colleges.json`
- `datasets/metadata/education/technical_colleges_reconciliation.json`

The active package contains 115 institution records retained from 122 numbered entries in the official NBTE Technical Colleges directory. It is a dated row-level directory snapshot, not a complete live national register. NBTE separately reports 153 technical colleges; the 38-entry coverage difference is not filled with inferred institutions. Websites, logos and API integration remain intentionally deferred.

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
