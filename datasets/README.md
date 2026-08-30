# Dataset Package Layout

This directory contains independently compiled dataset packages, JSON Schemas and provenance metadata.

## Layout

```text
datasets/
├── geography/
│   ├── geopolitical_zones.json
│   ├── lgas.json
│   └── states.json
├── education/
│   └── universities.json
├── finance/
│   ├── international_money_transfer_operators.json
│   └── payment_service_providers.json
├── metadata/
│   ├── education/
│   │   └── universities.json
│   ├── geography/
│   │   ├── geopolitical_zones.json
│   │   ├── lgas.json
│   │   └── states.json
│   └── finance/
│       ├── international_money_transfer_operators.json
│       └── payment_service_providers.json
├── schemas/
│   ├── education/
│   │   └── universities.schema.json
│   ├── geography/
│   │   ├── geopolitical_zones.schema.json
│   │   ├── lgas.schema.json
│   │   └── states.schema.json
│   └── finance/
│       ├── international_money_transfer_operators.schema.json
│       └── payment_service_providers.schema.json
└── LICENSE.md
```

## What Each File Is For

- `geography/geopolitical_zones.json` is the six-zone catalogue.
- `geography/lgas.json` is the compiled catalogue of Local Government Areas and FCT Area Councils.
- `geography/states.json` is the data file.
- `education/universities.json` is the compiled catalogue of current NUC-listed Nigerian universities.
- `finance/international_money_transfer_operators.json` is the compiled register snapshot of current CBN-listed IMTO entries.
- `finance/payment_service_providers.json` is the compiled register snapshot of payment-service-provider memberships.
- `schemas/geography/geopolitical_zones.schema.json` describes the zone record contract.
- `schemas/geography/lgas.schema.json` describes the LGA and Area Council record contract.
- `schemas/geography/states.schema.json` describes the record contract.
- `schemas/education/universities.schema.json` describes the university record contract.
- `schemas/finance/international_money_transfer_operators.schema.json` describes the IMTO record contract.
- `schemas/finance/payment_service_providers.schema.json` describes the payment-service-provider record contract.
- `metadata/education/universities.json` records provenance, versioning and licensing details for the university catalogue.
- `metadata/geography/geopolitical_zones.json` records provenance, versioning and licensing details for the zone catalogue.
- `metadata/geography/lgas.json` records provenance, versioning and licensing details for the LGA catalogue.
- `metadata/geography/states.json` records provenance, versioning and licensing details.
- `metadata/finance/international_money_transfer_operators.json` records provenance, versioning and licensing details for the IMTO catalogue.
- `metadata/finance/payment_service_providers.json` records provenance, versioning and licensing details for the payment-service-provider catalogue.
- `LICENSE.md` explains the dataset-content licence.

The IMTO catalogue is names-only, excludes addresses and inferred country data, and remains separate from the payment-service-provider catalogue. No HTTP routes are added by the dataset package itself.

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
