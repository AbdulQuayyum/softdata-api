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

Each record uses the lowercase alpha-2 code as its public `id`, preserves the source English name and published codes, and includes `calling_codes`, `flag_emoji`, `flag_svg_url`, `flag_url` and region hierarchy fields when the source table provides them. `flag_url` is qualified with `PUBLIC_API_URL` in API responses when configured. The package follows the current UN statistical manifest boundary, keeps territories and other areas where they appear in the official table, and states in metadata that the designations are statistical references only and do not imply political recognition or legal status. SoftData's independent compilation, schema and metadata are CC BY 4.0, while the UN source material retains its own rights. Flag SVG assets are vendored separately from MIT-licensed flag-icons v7.5.0.

This global package is separate from the Nigerian geography datasets, which continue to use `country_code: NG`.

### `world-time-zones`

The canonical IANA tzdb `zone1970.tab` manifest from release `2026c`, compiled into 312 time-zone records.

- Data: `datasets/geography/time_zones.json`
- Schema: `datasets/schemas/geography/time_zones.schema.json`
- Metadata: `datasets/metadata/geography/time_zones.json`

Each record preserves the exact IANA identifier as its public `id` and maps approved `world-countries-and-areas` IDs through `country_area_ids`. The package excludes aliases, `zone.tab`-only records, `backzone`, fixed-offset identifiers and other non-canonical entries, keeps `Asia/Taipei` with an empty `country_area_ids` array, records `bv` and `hm` as the two M49 country/area IDs with zero canonical zones, and omits static offsets, DST flags and coordinates. SoftData's independent compilation, schema and metadata are CC BY 4.0, while the IANA source files retain their own rights and are cited for reference only.

### `world-languages` and `world-country-languages`

Pinned to Unicode CLDR JSON 48.2.0, accessed 2026-09-04. The language catalogue has 633 current English base-language records. The relationship catalogue has 1,289 unique country-language pairs, covers all 248 approved country-area IDs, references 523 language IDs, and leaves 110 catalog languages without approved relationships by design. Status counts are `used`: 833, `official`: 319, `official_regional`: 117 and `de_facto_official`: 20.

- Data: `datasets/geography/languages.json`, `datasets/geography/country_languages.json`
- Schema: `datasets/schemas/geography/languages.schema.json`, `datasets/schemas/geography/country_languages.schema.json`
- Metadata: `datasets/metadata/geography/languages.json`, `datasets/metadata/geography/country_languages.json`

The compiler applies `fat -> ak`, `sh -> sr-Latn -> sr`, `tl -> fil` and `tw -> ak`, excludes script-qualified public IDs and non-country aggregates, and applies the base-row-wins rule for collapsed relationship conflicts. `used` is a normalized relationship value, not a legal, exhaustive or census statement. SoftData's compilation, schemas and metadata are CC BY 4.0; Unicode CLDR source material retains its own licence.

### Derived Country Profiles

`GET /v1/geography/countries/{country_id}/profile` is a derived view, not a separate static dataset. It combines `world-countries-and-areas`, `world-currencies`, `world-time-zones`, and `world-country-languages`. Its `language_ids` field contains unique, lexicographically sorted IDs from `world-languages` for every relationship status, including `official`, `de_facto_official`, `official_regional`, and `used`; inclusion does not mean that every language is nationally official. Empty arrays are valid. Use `GET /v1/geography/country-languages?country_area_id={country_id}` when relationship status details are required.

### `world-currencies`

The current ISO 4217 monetary-currency snapshot compiled from the SIX List One XML.

- Data: `datasets/finance/currencies.json`
- Schema: `datasets/schemas/finance/currencies.schema.json`
- Metadata: `datasets/metadata/finance/currencies.json`

Each record uses the lowercase alphabetic code as its public `id`, preserves `name`, `alphabetic_code`, `numeric_code`, `minor_unit` and `country_area_ids`, and links approved `world-countries-and-areas` IDs through `country_area_ids`. The package contains 155 current monetary currencies, excludes special-purpose and historical codes, keeps TWD with an empty `country_area_ids` array, and omits reverse mappings for Antarctica, State of Palestine and South Georgia and the South Sandwich Islands. SoftData's independent compilation, schema and metadata are CC BY 4.0, while the ISO and SIX source publications retain their own rights.

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
- Reconciliation: `datasets/metadata/finance/payment_service_providers_reconciliation.json`
- Phase 2A logo review: `datasets/metadata/finance/payment_service_provider_mobile_money_logo_reconciliation.json`
- Phase 2B logo review: `datasets/metadata/finance/cross_dataset_finance_logo_phase2b_reconciliation.json`
- Phase 2D commercial-bank logo review: `datasets/metadata/finance/cross_dataset_finance_logo_phase2d_reconciliation.json`
- Complete 610-file logo archive reconciliation: `datasets/metadata/finance/cross_dataset_finance_logo_reconciliation.json`
- Assets: `datasets/assets/financial-institutions/ng/payment-service-providers/`

The package contains 255 provider-category memberships across seven approved PSP categories. Each record is one provider-category membership, and a provider may appear in multiple categories. Optional enrichment currently covers 11 legacy CBN/sort-code values, 16 NIP institution-code values, 109 verified websites and 23 curated PNG marks; unresolved values are omitted. `scCode` is represented as `cbn_code` and `bankCode` as `nip_code` in the pinned community catalogue. Paystack data is supporting evidence only, and the community mappings may require future CBN/NIBSS confirmation. Phase 2A reviewed all 55 provisional unmatched mobile-money assets, accepting three target assignments and preserving stronger existing assets; the complete decision manifest records the remaining explicit rejections.

### `ng-international-money-transfer-operators`

Nigeria's current Central Bank of Nigeria IMTO register snapshot.

- Data: `datasets/finance/international_money_transfer_operators.json`
- Schema: `datasets/schemas/finance/international_money_transfer_operators.schema.json`
- Metadata: `datasets/metadata/finance/international_money_transfer_operators.json`

The package contains 108 CBN-listed IMTO entries. Each record represents one current register listing, with the source-side concatenation defect at SN 63 normalized transparently in metadata. Optional HTTPS website URLs and verified logo URLs are included where the global or Nigerian legal-entity relationship and asset provenance were established. IMTOs remain separate from the payment-service-provider register.

### `ng-commercial-banks`

Nigeria's 28 CBN-listed commercial banks with official websites, vendored identification logos, and partially verified CBN/NIP identifiers.

- Data: `datasets/finance/commercial_banks.json`
- Schema: `datasets/schemas/finance/commercial_banks.schema.json`
- Metadata: `datasets/metadata/finance/commercial_banks.json`
- Assets: `datasets/assets/banks/ng/`

Logo assets are served from the embedded application binary at `GET /v1/assets/banks/ng/{bank_id}.{ext}`. They are preserved unchanged from the pinned upstream source. `cbn_code` and `nip_code` are optional strings with three- and six-digit patterns respectively. `cbn_code` represents source `scCode`, the legacy CBN/sort-code field; `nip_code` represents source `bankCode`, the NIP institution-code field. Most values come from the pinned community catalogue, Paystack was supporting evidence only, and community mappings may require future CBN/NIBSS confirmation. Unresolved values are omitted. Bank names and logos are trademarks of their respective owners. They are provided for identification and directory purposes only; inclusion does not imply sponsorship, affiliation or endorsement.

### `ng-microfinance-banks`

Nigeria's roster-only v1 of 790 active microfinance banks from the reconciled CBN MFB register.

- Data: `datasets/finance/microfinance_banks.json`
- Schema: `datasets/schemas/finance/microfinance_banks.schema.json`
- Metadata: `datasets/metadata/finance/microfinance_banks.json`
- Reconciliation: `datasets/metadata/finance/microfinance_banks_reconciliation.json`

The public contract requires `id`, `name` and `country_code`; `cbn_code`, `nip_code`, `website_url` and `logo_url` are optional fields. Codes are strings sourced primarily from the pinned community `ng-bank-logos` catalogue: `bankCode` is accepted only as a six-digit NIP institution code and `scCode` only as a three-digit legacy CBN/sort-code field. Provider-specific Paystack codes, malformed values and ambiguous identities remain omitted. Community mappings may require future confirmation against a current CBN/NIBSS register. The 829-row CBN snapshot is reduced by 39 evidence-backed exclusions: four duplicate/stale rows, 33 matched revocations, the AKPO predecessor and Verdant-Capital. Website/logo provenance and batch progress are recorded in `datasets/metadata/finance/microfinance_banks_enrichment.json`; the complete 315-file Phase 2C archive review is recorded in `datasets/metadata/finance/microfinance_banks_phase2c_logo_reconciliation.json`. Logo acceptance requires exact identity and byte validation; explicit redistribution permission is not a prerequisite, and institutional rights remain with the respective owners. Current enrichment coverage is 330 logos for 790 records. The public API exposes list and detail routes at `/v1/finance/microfinance-banks` and `/v1/finance/microfinance-banks/{bank_id}` without pagination or filters.

### Nigerian regulated-finance category snapshots

The following dataset packages preserve the six supplied CBN category exports as independent finance datasets. CBN/NIP codes are optional only for the applicable bank categories; holding-company and DFI contracts intentionally exclude bank identifiers. Websites and logos are optional; no category-specific HTTP routes are introduced by these packages.

- `ng-non-interest-financial-institutions`: `datasets/finance/non_interest_institutions.json` (6 records)
- `ng-merchant-banks`: `datasets/finance/merchant_banks.json` (6 records)
- `ng-payment-service-banks`: `datasets/finance/payment_service_banks.json` (5 records)
- `ng-financial-holding-companies`: `datasets/finance/financial_holding_companies.json` (7 records)
- `ng-development-finance-institutions`: `datasets/finance/development_finance_institutions.json` (8 records)
- `ng-primary-mortgage-institutions`: `datasets/finance/primary_mortgage_institutions.json` (31 records)

The source workbooks are `Export (9).xlsx`, `Export (6).xlsx`, `Export (10).xlsx`, `Export (5).xlsx`, `Export (2).xlsx` and `Export (11).xlsx`, respectively. The non-interest package uses the broader financial-institutions key because its CBN snapshot includes Mint Microfinance Bank; the source classification is preserved rather than silently dropping the row. The PMI snapshot excludes ASO Savings and Loans after liquidation and reconciles documented successor names. Holding companies remain separate from their bank subsidiaries and never inherit subsidiary identifiers.

Verified PNG logos are embedded under `datasets/assets/financial-institutions/ng/` where exact institution identity and source hashes are recorded. Baseline assets use the pinned upstream MIT source; official-source marks are separately documented and are not claimed under SoftData CC BY 4.0. The FHA Homes asset is explicitly a Federal Housing Authority parent-brand representative mark. Current coverage is 6/6 non-interest, 6/6 merchant, 5/5 payment-service, 7/7 holding-company, 8/8 DFI and 31/31 primary-mortgage records.

Across the six datasets, the current decoded coverage is 63/63 websites and 63/63 logos. This comprises 62 independently sourced institution marks and one disclosed Federal Housing Authority parent-brand representative mark for the wholly owned FHA Homes subsidiary. Akwa Savings was reconciled to current Ibom Mortgage Bank while retaining the stable dataset ID; TrustBond Mortgage Bank was removed after the FirstTrust merger was confirmed.

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

### `ng-polytechnics`

Nigeria's reconciled active polytechnic roster from the NBTE 25th Edition 2025 directory.

- Data: `datasets/education/polytechnics.json`
- Schema: `datasets/schemas/education/polytechnics.schema.json`
- Metadata: `datasets/metadata/education/polytechnics.json`
- Reconciliation: `datasets/metadata/education/polytechnics_reconciliation.json`

The package contains 168 records: 35 federal, 43 state and 90 private polytechnics. The compact public contract contains only `id`, `name`, `ownership_type`, `state_id` and `country_code`; specialist institutes, colleges and transition records from the mixed source section are excluded and documented. This pass adds only the static dataset foundation; websites, logos, repository methods and HTTP routes are intentionally deferred.

### `ng-monotechnics`

Nigeria's reconciled roster of NBTE **Specialised Institutions (Monotechnics)**.

- Data: `datasets/education/monotechnics.json`
- Schema: `datasets/schemas/education/monotechnics.schema.json`
- Metadata: `datasets/metadata/education/monotechnics.json`
- Reconciliation: `datasets/metadata/education/monotechnics_reconciliation.json`

The package contains 86 records: 32 federal, 4 state and 50 private institutions across 27 states. It uses the same five-field static contract as the university, college-of-education and Polytechnic datasets. Colleges of education, general schools, remedial institutions and other wrong-category entries are excluded in the reconciliation manifest. Website/logo enrichment and API integration are intentionally deferred.

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
- International money-transfer operators include optional HTTPS website URLs where the Nigerian/global legal-entity relationship was verified. Logo decisions and research provenance are recorded in `datasets/metadata/finance/international_money_transfer_operators_enrichment.json`.
