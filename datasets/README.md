# Dataset Package Layout

This directory contains independently compiled dataset packages, JSON Schemas and provenance metadata.

## Layout

```text
datasets/
├── geography/
│   ├── countries_and_areas.json
│   ├── geopolitical_zones.json
│   ├── lgas.json
│   └── states.json
├── education/
│   ├── colleges_of_education.json
│   └── universities.json
├── finance/
│   ├── international_money_transfer_operators.json
│   └── payment_service_providers.json
├── metadata/
│   ├── geography/
│   │   ├── countries_and_areas.json
│   │   ├── geopolitical_zones.json
│   │   ├── lgas.json
│   │   └── states.json
│   ├── education/
│   │   └── universities.json
│   └── finance/
│       ├── international_money_transfer_operators.json
│       └── payment_service_providers.json
├── schemas/
│   ├── geography/
│   │   ├── countries_and_areas.schema.json
│   │   ├── geopolitical_zones.schema.json
│   │   ├── lgas.schema.json
│   │   └── states.schema.json
│   ├── education/
│   │   └── universities.schema.json
│   └── finance/
│       ├── international_money_transfer_operators.schema.json
│       └── payment_service_providers.schema.json
└── LICENSE.md
```

## What Each File Is For

- `geography/geopolitical_zones.json` is the six-zone catalogue.
- `geography/lgas.json` is the compiled catalogue of Local Government Areas and FCT Area Councils.
- `geography/states.json` is the data file.
- `education/colleges_of_education.json` is the compiled catalogue of current NCCE-listed colleges of education.
- `education/universities.json` is the compiled catalogue of current NUC-listed Nigerian universities.
- `finance/international_money_transfer_operators.json` is the compiled register snapshot of current CBN-listed IMTO entries.
- `finance/payment_service_providers.json` is the compiled register snapshot of payment-service-provider memberships.
- `schemas/geography/geopolitical_zones.schema.json` describes the zone record contract.
- `schemas/geography/lgas.schema.json` describes the LGA and Area Council record contract.
- `schemas/geography/states.schema.json` describes the record contract.
- `schemas/education/colleges_of_education.schema.json` describes the college-of-education record contract.
- `schemas/education/universities.schema.json` describes the university record contract.
- `schemas/finance/international_money_transfer_operators.schema.json` describes the IMTO record contract.
- `schemas/finance/payment_service_providers.schema.json` describes the payment-service-provider record contract.
- `metadata/education/universities.json` records provenance, versioning and licensing details for the university catalogue.
- `metadata/education/colleges_of_education.json` records provenance, versioning and licensing details for the colleges-of-education catalogue.
- `metadata/geography/geopolitical_zones.json` records provenance, versioning and licensing details for the zone catalogue.
- `metadata/geography/lgas.json` records provenance, versioning and licensing details for the LGA catalogue.
- `metadata/geography/states.json` records provenance, versioning and licensing details.
- `metadata/finance/international_money_transfer_operators.json` records provenance, versioning and licensing details for the IMTO catalogue.
- `metadata/finance/payment_service_providers.json` records provenance, versioning and licensing details for the payment-service-provider catalogue.
- `LICENSE.md` explains the dataset-content licence.

The world countries-and-areas catalogue is compiled from the current English UN M49 overview table, uses the lowercase alpha-2 code as its public `id`, and preserves the source names, ISO alpha codes, numeric codes and available region hierarchy fields. Its current boundary is 248 countries or areas, and SoftData's independent compilation, schema and metadata are CC BY 4.0 while the UN source material retains its own rights and is used for statistical reference only.

This global package is separate from the Nigerian geography datasets, which continue to use `country_code: NG`.

The IMTO catalogue is names-only, excludes addresses and inferred country data, and remains separate from the payment-service-provider catalogue. No HTTP routes are added by the dataset package itself.

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
