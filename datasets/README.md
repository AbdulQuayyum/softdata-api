# Dataset Package Layout

This directory contains independently compiled dataset packages, JSON Schemas and provenance metadata.

## Layout

```text
datasets/
├── geography/
│   ├── countries_and_areas.json
│   ├── geopolitical_zones.json
│   ├── lgas.json
│   ├── time_zones.json
│   └── states.json
├── education/
│   ├── colleges_of_education.json
│   ├── monotechnics.json
│   ├── polytechnics.json
│   └── universities.json
├── finance/
│   ├── development_finance_institutions.json
│   ├── financial_holding_companies.json
│   ├── microfinance_banks.json
│   ├── merchant_banks.json
│   ├── non_interest_institutions.json
│   ├── payment_service_banks.json
│   ├── primary_mortgage_institutions.json
│   ├── currencies.json
│   ├── international_money_transfer_operators.json
│   └── payment_service_providers.json
├── metadata/
│   ├── geography/
│   │   ├── countries_and_areas.json
│   │   ├── geopolitical_zones.json
│   │   ├── lgas.json
│   │   ├── time_zones.json
│   │   └── states.json
│   ├── education/
│   │   ├── colleges_of_education.json
│   │   ├── monotechnics.json
│   │   ├── monotechnics_reconciliation.json
│   │   ├── polytechnics.json
│   │   └── universities.json
│   └── finance/
│       ├── development_finance_institutions.json
│       ├── financial_holding_companies.json
│       ├── microfinance_banks.json
│       ├── microfinance_banks_enrichment.json
│       ├── merchant_banks.json
│       ├── non_interest_institutions.json
│       ├── payment_service_banks.json
│       ├── primary_mortgage_institutions.json
│       ├── currencies.json
│       ├── international_money_transfer_operators.json
│       └── payment_service_providers.json
├── schemas/
│   ├── geography/
│   │   ├── countries_and_areas.schema.json
│   │   ├── geopolitical_zones.schema.json
│   │   ├── lgas.schema.json
│   │   ├── time_zones.schema.json
│   │   └── states.schema.json
│   ├── education/
│   │   ├── colleges_of_education.schema.json
│   │   ├── monotechnics.schema.json
│   │   ├── polytechnics.schema.json
│   │   └── universities.schema.json
│   └── finance/
│       ├── development_finance_institutions.schema.json
│       ├── financial_holding_companies.schema.json
│       ├── microfinance_banks.schema.json
│       ├── merchant_banks.schema.json
│       ├── non_interest_institutions.schema.json
│       ├── payment_service_banks.schema.json
│       ├── primary_mortgage_institutions.schema.json
│       ├── currencies.schema.json
│       ├── international_money_transfer_operators.schema.json
│       └── payment_service_providers.schema.json
└── LICENSE.md
```

Regulated-finance logo assets are embedded under `assets/financial-institutions/ng/` by dataset category. Microfinance marks may come from identity-verified first-party, archived first-party, official-social or curated repository sources; source URLs, retrieval timestamps, hashes and provenance are recorded in the category attribution and reconciliation files. Missing or identity-ambiguous logos remain omitted. Microfinance enrichment progress is recorded per deterministic batch in `metadata/finance/microfinance_banks_enrichment.json`.

## What Each File Is For

- `geography/geopolitical_zones.json` is the six-zone catalogue.
- `geography/lgas.json` is the compiled catalogue of Local Government Areas and FCT Area Councils.
- `geography/states.json` is the data file.
- `geography/time_zones.json` is the compiled catalogue of canonical IANA time zones from `zone1970.tab`.
- `education/colleges_of_education.json` is the compiled catalogue of current NCCE-listed colleges of education.
- `education/polytechnics.json` is the 168-record static catalogue of current NBTE-listed Nigerian polytechnics.
- `education/monotechnics.json` is the 86-record static catalogue of NBTE Specialised Institutions (Monotechnics).
- `education/universities.json` is the compiled catalogue of current NUC-listed Nigerian universities.
- `finance/international_money_transfer_operators.json` is the compiled register snapshot of current CBN-listed IMTO entries, with optional verified website URLs.
- `finance/currencies.json` is the compiled snapshot of current ISO 4217 monetary currencies.
- `finance/payment_service_providers.json` is the compiled register snapshot of payment-service-provider memberships.
- `metadata/finance/payment_service_providers_reconciliation.json` records the 255-record code, website and logo reconciliation for payment-service providers. The pinned community catalogue maps `scCode` to the optional three-digit `cbn_code` field and `bankCode` to the optional six-digit `nip_code` field; Paystack data is supporting evidence only.
- `metadata/finance/payment_service_provider_mobile_money_logo_reconciliation.json` records the complete 55-file Phase 2A mobile-money source review, including source hashes, dimensions, target IDs, evidence and explicit rejection decisions.
- `metadata/finance/cross_dataset_finance_logo_phase2b_reconciliation.json` records the complete 29-file Phase 2B merchant-bank, mortgage-bank and payment-service-bank review, including source and existing-target hashes.
- `metadata/finance/cross_dataset_finance_logo_phase2d_reconciliation.json` records the complete 32-file Phase 2D commercial-bank review, including cross-category target identities and source/target hashes.
- `metadata/finance/cross_dataset_finance_logo_reconciliation.json` records the final classified inventory of all 610 archive files; no archive file is unclassified.
- `assets/financial-institutions/ng/payment-service-providers/` contains the 23 curated PNG marks accepted in that reconciliation. Logos remain third-party trademarks and are not relicensed as SoftData content.
- `finance/microfinance_banks.json` is the 790-record snapshot of active Nigerian microfinance banks after the reconciled CBN/NDIC exclusions. `cbn_code`, `nip_code`, `website_url` and `logo_url` are optional and are added only after identity and value/asset verification. `bankCode` values are accepted only as six-digit NIP institution codes and `scCode` values only as three-digit legacy CBN/sort-code fields; most values come from the pinned community catalogue, Paystack is supporting evidence only, and future CBN/NIBSS confirmation may be required. Batch evidence is in `metadata/finance/microfinance_banks_enrichment.json`.
- `metadata/finance/microfinance_banks_phase2c_logo_reconciliation.json` records the complete 315-file Phase 2C MFB archive review, including code-first decisions, source hashes, target candidates and existing-target comparisons.
- The microfinance enrichment currently contains 330 embedded PNG logos for 790 records; 316 are curated-repository imports and 14 source entries from the corrected unused-source review matched active records, with the existing first-party `b-c-kash-microfinance-bank` asset preserved.
- The 138 unused source filenames from the import bundle remain manual-review cases because their PNG bytes and source metadata were not included; see `metadata/finance/microfinance_banks_unused_source_reconciliation.json`.
- `finance/non_interest_institutions.json`, `finance/merchant_banks.json`, `finance/payment_service_banks.json`, `finance/financial_holding_companies.json`, `finance/development_finance_institutions.json` and `finance/primary_mortgage_institutions.json` are CBN category snapshots from the supplied August 2026 exports. CBN/NIP fields are optional only in the categories that expose them; holding-company and DFI contracts intentionally exclude bank identifiers. Websites and logos are optional and omitted when not verified.
- `schemas/geography/geopolitical_zones.schema.json` describes the zone record contract.
- `schemas/geography/lgas.schema.json` describes the LGA and Area Council record contract.
- `schemas/geography/states.schema.json` describes the record contract.
- `schemas/geography/time_zones.schema.json` describes the canonical IANA time-zone record contract.
- `schemas/education/colleges_of_education.schema.json` describes the college-of-education record contract.
- `schemas/education/polytechnics.schema.json` describes the five-field polytechnic record contract.
- `schemas/education/monotechnics.schema.json` describes the five-field monotechnic record contract.
- `schemas/education/universities.schema.json` describes the university record contract.
- `schemas/finance/international_money_transfer_operators.schema.json` describes the IMTO record contract.
- `schemas/finance/currencies.schema.json` describes the currency record contract.
- `schemas/finance/payment_service_providers.schema.json` describes the payment-service-provider record contract.
- The six related Nigerian regulated-finance schemas use Draft 2020-12 arrays with fixed snapshot counts, `NG` country codes, deterministic IDs and optional code, website and logo fields.
- `metadata/education/universities.json` records provenance, versioning and licensing details for the university catalogue.
- `metadata/education/colleges_of_education.json` records provenance, versioning and licensing details for the colleges-of-education catalogue.
- `metadata/education/polytechnics.json` and `metadata/education/polytechnics_reconciliation.json` record source provenance and roster decisions for polytechnics.
- `metadata/education/monotechnics.json` and `metadata/education/monotechnics_reconciliation.json` record source provenance and roster decisions for specialised institutions.
- `metadata/geography/geopolitical_zones.json` records provenance, versioning and licensing details for the zone catalogue.
- `metadata/geography/lgas.json` records provenance, versioning and licensing details for the LGA catalogue.
- `metadata/geography/states.json` records provenance, versioning and licensing details.
- `metadata/geography/time_zones.json` records provenance, versioning and licensing details for the time-zone catalogue.
- `metadata/finance/international_money_transfer_operators.json` records provenance, versioning and licensing details for the IMTO catalogue; `international_money_transfer_operators_enrichment.json` records website/logo research decisions.
- `metadata/finance/currencies.json` records provenance, versioning and licensing details for the currency catalogue.
- `metadata/finance/payment_service_providers.json` records provenance, versioning and licensing details for the payment-service-provider catalogue.
- `metadata/finance/microfinance_banks.json` records the pinned CBN snapshot, NDIC status sources, exclusion arithmetic, optional identifier coverage, deferred fields and licensing boundary for the microfinance-bank roster. Cross-dataset identifier evidence is in `metadata/finance/financial_institution_codes_reconciliation.json`.
- `metadata/finance/microfinance_banks_reconciliation.json` preserves the 39-row source-ID exclusion manifest used to reproduce the public 790-record result.
- `LICENSE.md` explains the dataset-content licence.

The world countries-and-areas catalogue is compiled from the current English UN M49 overview table, uses the lowercase alpha-2 code as its public `id`, and preserves the source names, ISO alpha codes, numeric codes, calling codes, flag emoji, flag SVG URLs and available region hierarchy fields. Its current boundary is 248 countries or areas, and SoftData's independent compilation, schema and metadata are CC BY 4.0 while the UN source material retains its own rights and is used for statistical reference only. Flag SVG assets are vendored separately from MIT-licensed flag-icons v7.5.0.

This global package is separate from the Nigerian geography datasets, which continue to use `country_code: NG`.

The world time-zones catalogue is compiled from the canonical IANA tzdb `zone1970.tab` manifest in release `2026c`, uses the exact IANA identifier as its public `id`, and maps approved `world-countries-and-areas` IDs through `country_area_ids`. Its current boundary is 312 time zones, `Asia/Taipei` intentionally carries an empty `country_area_ids` array, and `bv` plus `hm` are the two M49 country/area IDs with zero canonical zones. Aliases, `zone.tab`-only records, `backzone`, static offsets, DST fields and coordinates are excluded from v1. SoftData's independent compilation, schema and metadata are CC BY 4.0 while the IANA source files retain their own rights and are cited for reference only.

The language catalogue is pinned to Unicode CLDR JSON 48.2.0 and contains 633 current base-language IDs with English display names. Its country-language companion contains 1,289 unique relationships, covers all 248 approved country-area IDs, references 523 languages, and intentionally leaves 110 catalog languages without relationships. Status counts are 833 `used`, 319 `official`, 117 `official_regional` and 20 `de_facto_official`. The approved alias remaps are `fat -> ak`, `sh -> sr-Latn -> sr`, `tl -> fil` and `tw -> ak`; base-language rows win status conflicts after script-qualified collapse. SoftData's compilation, schemas and metadata are CC BY 4.0; Unicode CLDR source material retains its own licence.

Country Profiles are derived API views, not static dataset records. The profile endpoint combines `world-countries-and-areas`, `world-currencies`, `world-time-zones`, and `world-country-languages`; `language_ids` references `world-languages.id`, is unique and sorted, and includes relationships across every published status. Use the country-language endpoint for status details. Empty language arrays are valid.

The IMTO catalogue is names-only, excludes addresses and inferred country data, and remains separate from the payment-service-provider catalogue. No HTTP routes are added by the dataset package itself.

The world-currencies catalogue is compiled from the current ISO 4217 monetary-currency snapshot published in the SIX List One XML, uses the lowercase alphabetic code as its public `id`, keeps `name`, `alphabetic_code`, `numeric_code`, `minor_unit` and `country_area_ids` as the public fields, and maps approved `world-countries-and-areas` IDs through `country_area_ids`. The package contains 155 current monetary currencies, excludes special-purpose and historical codes, keeps TWD with an empty `country_area_ids` array, and omits reverse mappings for Antarctica, State of Palestine and South Georgia and the South Sandwich Islands. SoftData's independent compilation, schema and metadata are CC BY 4.0, while the ISO and SIX source publications retain their own rights.

The colleges-of-education catalogue is compiled from the current NCCE accredited-colleges register, maps each record to `ng-states` through `state_id`, excludes the stale Cross River row that has a successor university in `ng-universities`, and keeps the public model compact at `id`, `name`, `ownership_type`, `state_id` and `country_code`. The package version is `1.0.0`, and the current boundary is 244 records split into 28 federal, 48 state and 168 private colleges. SoftData's independent compilation, schema and metadata are CC BY 4.0; the NCCE and other official publications retain their own rights.

The university catalogue is compiled from the current National Universities Commission federal, state and private registers, and preserves `ownership_type` and `state_id` as the public grouping fields.

## Licensing

The dataset package is independently compiled by SoftData and is available under CC BY 4.0.

Attribution is required. Suggested attribution:

`SoftData API contributors, “Nigeria States and Federal Capital Territory”, version 1.0.0.`

The source organizations listed in the metadata retain the rights they hold in their own publications.

## Versioning

Dataset versions use semantic versioning.

- Bump the major version for breaking changes.
- Bump the minor version for additive changes.
- Bump the patch version for corrections that do not change the contract.

## Contribution Rules

- Keep records independently compiled and cross-verified.
- Do not merge unverified additions.
- Correct records only with an authoritative citation.
- Keep metadata, schema and data files in sync.
