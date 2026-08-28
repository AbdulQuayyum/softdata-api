# Dataset Package Layout

This directory contains independently compiled dataset packages, JSON Schemas and provenance metadata.

## Layout

```text
datasets/
├── geography/
│   ├── geopolitical_zones.json
│   └── states.json
├── metadata/
│   └── geography/
│       ├── geopolitical_zones.json
│       └── states.json
├── schemas/
│   └── geography/
│       ├── geopolitical_zones.schema.json
│       └── states.schema.json
└── LICENSE.md
```

## What Each File Is For

- `geography/geopolitical_zones.json` is the six-zone catalogue.
- `geography/states.json` is the data file.
- `schemas/geography/geopolitical_zones.schema.json` describes the zone record contract.
- `schemas/geography/states.schema.json` describes the record contract.
- `metadata/geography/geopolitical_zones.json` records provenance, versioning and licensing details for the zone catalogue.
- `metadata/geography/states.json` records provenance, versioning and licensing details.
- `LICENSE.md` explains the dataset-content licence.

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
