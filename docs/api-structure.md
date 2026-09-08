# API Structure

SoftData API uses a layered Go layout that keeps entry points, configuration, persistence, domain models, HTTP handling, and documentation separated.

## Overview

- `cmd/` contains executable entry points.
- `internal/` contains application configuration, infrastructure, domain models, repositories, services, HTTP handlers, middleware, routing, validation, security, and response helpers.
- `datasets/` contains versioned source data, schemas, provenance metadata, licensing notes, and embedded dataset, flag, and regulated-finance logo assets.
- `database/` contains PostgreSQL migrations and handwritten SQL queries.
- `docs/` contains API and project documentation, including the OpenAPI contract.
- `tools/` contains dataset generation and snapshot tooling.
- Root files provide development configuration, build metadata, contribution guidance, and project licensing.
- `api`, `build/`, `dist/`, and `tmp/` are generated or local development artifacts and are not application source files. `.env` is a local ignored configuration file; `.env.example` is the shareable configuration template.

## Project Tree

The tree below reflects the repository layout. The flag directory contains one SVG per supported country or area; the wildcard represents all 248 SVG files.

```text
softdata-api/
├── .air.toml
├── .env
├── .env.example
├── .gitignore
├── api                         # locally built API binary
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── CONTRIBUTING.md
├── DATASETS.md
├── LICENSE
├── Makefile
├── README.md
├── SECURITY.md
├── go.mod
├── go.sum
├── sqlc.yaml
├── cmd/
│   └── api/
│       └── main.go
├── database/
│   ├── migrations/
│   │   ├── 000001_create_accounts.down.sql
│   │   ├── 000001_create_accounts.up.sql
│   │   ├── 000002_create_sessions.down.sql
│   │   ├── 000002_create_sessions.up.sql
│   │   ├── 000003_create_api_keys.down.sql
│   │   ├── 000003_create_api_keys.up.sql
│   │   ├── 000004_create_api_requests.down.sql
│   │   ├── 000004_create_api_requests.up.sql
│   │   ├── 000005_create_usage_daily.down.sql
│   │   ├── 000005_create_usage_daily.up.sql
│   │   ├── 000006_create_datasets.down.sql
│   │   ├── 000006_create_datasets.up.sql
│   │   ├── 000007_create_dataset_sources.down.sql
│   │   ├── 000007_create_dataset_sources.up.sql
│   │   ├── 000008_create_dataset_versions.down.sql
│   │   ├── 000008_create_dataset_versions.up.sql
│   │   ├── 000009_add_api_request_dataset_group.down.sql
│   │   └── 000009_add_api_request_dataset_group.up.sql
│   └── queries/
│       ├── accounts.sql
│       ├── api_keys.sql
│       ├── datasets.sql
│       ├── sessions.sql
│       └── usage.sql
├── datasets/
│   ├── LICENSE.md
│   ├── README.md
│   ├── embedded.go                         # embeds runtime JSON datasets for deployments without a local volume
│   ├── assets/
│   │   ├── banks.go
│   │   ├── banks_test.go
│   │   ├── flags.go
│   │   ├── flags_test.go
│   │   ├── banks/
│   │   │   └── ng/
│   │   │       ├── ATTRIBUTION.md
│   │   │       ├── LICENSES/
│   │   │       │   └── Nigerian-Bank-Logos-MIT.txt
│   │   │       └── *.png                  # 28 vendored commercial-bank assets
│   │   ├── financial_institution_logos.go
│   │   ├── financial_institutions_test.go
│   │   ├── financial-institutions/
│   │   │   ├── LICENSE
│   │   │   └── ng/
│   │   │       ├── ACQUISITION_AUDIT.md
│   │   │       ├── ATTRIBUTION.md
│   │   │       ├── LOGO_PERMISSION_REQUESTS.md
│   │   │       ├── development-finance/*.png
│   │   │       ├── holding-companies/*.png
│   │   │       ├── international-money-transfer-operators/*.png
│   │   │       ├── merchant-banks/*.png
│   │   │       ├── microfinance-banks/*.png
│   │   │       ├── microfinance-banks/ACQUISITION_AUDIT.md
│   │   │       ├── microfinance-banks/ATTRIBUTION.md
│   │   │       ├── microfinance-banks/LOGO_PERMISSION_REQUESTS.md
│   │   │       └── microfinance-banks/SOURCE_RECONCILIATION.md
│   │   │       ├── non-interest/*.png
│   │   │       ├── payment-service-banks/*.png
│   │   │       ├── payment-service-providers/*.png
│   │   │       └── primary-mortgage/*.png  # regulated-finance PNG assets
│   │   └── flags/
│   │       ├── ATTRIBUTION.md
│   │       ├── LICENSE
│   │       └── 4x3/
│   │           └── *.svg                  # 248 vendored flag assets
│   ├── education/
│   │   ├── colleges_of_education.json
│   │   ├── colleges_of_agriculture.json
│   │   ├── colleges_of_health_sciences_and_technology.json
│   │   ├── colleges_of_nursing_and_midwifery.json
│   │   ├── primary_and_secondary_schools.json
│   │   ├── technical_colleges.json
│   │   ├── monotechnics.json
│   │   ├── polytechnics.json
│   │   ├── universities.json
│   │   └── vocational_enterprise_institutions.json
│   ├── healthcare/
│   │   └── health_facilities.json
│   ├── finance/
│   │   ├── development_finance_institutions.json
│   │   ├── financial_holding_companies.json
│   │   ├── microfinance_banks.json
│   │   ├── merchant_banks.json
│   │   ├── non_interest_institutions.json
│   │   ├── payment_service_banks.json
│   │   ├── primary_mortgage_institutions.json
│   │   ├── commercial_banks.json
│   │   ├── currencies.json
│   │   ├── international_money_transfer_operators.json
│   │   └── payment_service_providers.json
│   ├── geography/
│   │   ├── countries_and_areas.json
│   │   ├── country_languages.json
│   │   ├── geopolitical_zones.json
│   │   ├── languages.json
│   │   ├── lgas.json
│   │   ├── states.json
│   │   └── time_zones.json
│   ├── metadata/
│   │   ├── education/
│   │   │   ├── colleges_of_education.json
│   │   │   ├── colleges_of_agriculture.json
│   │   │   ├── colleges_of_agriculture_reconciliation.json
│   │   │   ├── colleges_of_health_sciences_and_technology.json
│   │   │   ├── colleges_of_health_sciences_and_technology_reconciliation.json
│   │   │   ├── colleges_of_nursing_and_midwifery.json
│   │   │   ├── colleges_of_nursing_and_midwifery_reconciliation.json
│   │   │   ├── monotechnics.json
│   │   │   ├── vocational_enterprise_institutions.json
│   │   │   ├── monotechnics_reconciliation.json
│   │   │   ├── polytechnics.json
│   │   │   ├── polytechnics_reconciliation.json
│   │   │   ├── primary_and_secondary_schools.json
│   │   │   ├── primary_and_secondary_schools_checkpoint.json
│   │   │   ├── primary_and_secondary_schools_reconciliation/
│   │   │   │   ├── index.json
│   │   │   │   └── {state-or-fct}.json          # 37 state/FCT partition summaries
│   │   │   ├── technical_colleges.json
│   │   │   ├── technical_colleges_reconciliation.json
│   │   │   ├── universities.json
│   │   │   └── vocational_enterprise_institutions.json
│   │   ├── healthcare/
│   │   │   ├── health_facilities.json
│   │   │   └── health_facilities_reconciliation/
│   │   │       ├── index.json
│   │   │       └── {state_id}.json
│   │   ├── finance/
│   │   │   ├── development_finance_institutions.json
│   │   │   ├── financial_holding_companies.json
│   │   │   ├── microfinance_banks.json
│   │   │   ├── microfinance_banks_enrichment.json
│   │   │   ├── financial_institution_codes_reconciliation.json
│   │   │   ├── merchant_banks.json
│   │   │   ├── non_interest_institutions.json
│   │   │   ├── payment_service_banks.json
│   │   │   ├── primary_mortgage_institutions.json
│   │   │   ├── commercial_banks.json
│   │   │   ├── currencies.json
│   │   │   ├── international_money_transfer_operators.json
│   │   │   ├── international_money_transfer_operators_enrichment.json
│   │   │   ├── microfinance_banks_reconciliation.json
│   │   │   ├── payment_service_providers.json
│   │   │   ├── payment_service_providers_reconciliation.json
│   │   │   ├── payment_service_provider_mobile_money_logo_reconciliation.json
│   │   │   ├── cross_dataset_finance_logo_phase2b_reconciliation.json
│   │   │   ├── cross_dataset_finance_logo_phase2d_reconciliation.json
│   │   │   ├── cross_dataset_finance_logo_reconciliation.json
│   │   │   ├── microfinance_banks_phase2c_logo_reconciliation.json
│   │   │   └── microfinance_banks_unused_source_reconciliation.json
│   │   └── geography/
│   │       ├── countries_and_areas.json
│   │       ├── country_languages.json
│   │       ├── geopolitical_zones.json
│   │       ├── languages.json
│   │       ├── lgas.json
│   │       ├── states.json
│   │       └── time_zones.json
│   └── schemas/
│       ├── education/
│       │   ├── colleges_of_education.schema.json
│       │   ├── colleges_of_agriculture.schema.json
│       │   ├── colleges_of_health_sciences_and_technology.schema.json
│       │   ├── colleges_of_nursing_and_midwifery.schema.json
│       │   ├── primary_and_secondary_schools.schema.json
│       │   ├── technical_colleges.schema.json
│       │   ├── monotechnics.schema.json
│       │   ├── vocational_enterprise_institutions.schema.json
│       │   ├── polytechnics.schema.json
│       │   └── universities.schema.json
│       ├── healthcare/
│       │   └── health_facilities.schema.json
│       ├── finance/
│       │   ├── development_finance_institutions.schema.json
│       │   ├── financial_holding_companies.schema.json
│       │   ├── microfinance_banks.schema.json
│       │   ├── merchant_banks.schema.json
│       │   ├── non_interest_institutions.schema.json
│       │   ├── payment_service_banks.schema.json
│       │   ├── primary_mortgage_institutions.schema.json
│       │   ├── commercial_banks.schema.json
│       │   ├── currencies.schema.json
│       │   ├── international_money_transfer_operators.schema.json
│       │   └── payment_service_providers.schema.json
│       └── geography/
│           ├── countries_and_areas.schema.json
│           ├── country_languages.schema.json
│           ├── geopolitical_zones.schema.json
│           ├── languages.schema.json
│           ├── lgas.schema.json
│           ├── states.schema.json
│           └── time_zones.schema.json
├── docs/
│   ├── api-keys.md
│   ├── api-structure.md
│   ├── authentication.md
│   ├── datasets.md
│   ├── errors.md
│   ├── openapi.yaml
│   ├── quick-start.md
│   ├── rate-limits.md
│   ├── softdata-api.postman_collection.json
│   └── versioning.md
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   ├── app_test.go
│   │   ├── dependencies.go
│   │   ├── education_test.go
│   │   ├── finance_test.go
│   │   ├── microfinance_bootstrap_test.go
│   │   ├── geography_test.go
│   │   └── shutdown.go
│   ├── config/
│   │   ├── config.go
│   │   ├── config_test.go
│   │   ├── database.go
│   │   ├── datasets.go
│   │   ├── rate_limit.go
│   │   ├── redis.go
│   │   ├── security.go
│   │   ├── server.go
│   │   └── usage.go
│   ├── database/
│   │   ├── health.go
│   │   ├── postgres.go
│   │   ├── postgres_test.go
│   │   └── sqlc/
│   │       ├── accounts.sql.go
│   │       ├── api_keys.sql.go
│   │       ├── datasets.sql.go
│   │       ├── db.go
│   │       ├── models.go
│   │       ├── querier.go
│   │       ├── sessions.sql.go
│   │       ├── sessions_rotation_test.go
│   │       └── usage.sql.go
│   ├── handlers/
│   │   ├── account_handler.go
│   │   ├── account_handler_test.go
│   │   ├── api_key_handler.go
│   │   ├── api_key_handler_test.go
│   │   ├── auth_handler.go
│   │   ├── auth_handler_test.go
│   │   ├── dataset_handler.go
│   │   ├── dataset_handler_test.go
│   │   ├── discovery_handler.go
│   │   ├── discovery_handler_test.go
│   │   ├── education_colleges_handler_test.go
│   │   ├── education_handler.go
│   │   ├── education_handler_test.go
│   │   ├── education_extended_handler.go
│   │   ├── education_extended_handler_test.go
│   │   ├── finance_handler.go
│   │   ├── finance_handler_test.go
│   │   ├── finance_commercial_banks_test.go
│   │   ├── finance_microfinance_banks_test.go
│   │   ├── finance_microfinance_banks_openapi_test.go
│   │   ├── geography_handler.go
│   │   ├── geography_handler_test.go
│   │   ├── geography_languages_handler_test.go
│   │   ├── geography_languages_openapi_test.go
│   │   ├── health_handler.go
│   │   ├── health_handler_test.go
│   │   ├── openapi_test.go
│   │   ├── usage_handler.go
│   │   └── usage_handler_test.go
│   ├── middlewares/
│   │   ├── authentication.go
│   │   ├── authentication_test.go
│   │   ├── body_limit.go
│   │   ├── body_limit_test.go
│   │   ├── cors.go
│   │   ├── cors_test.go
│   │   ├── identity.go
│   │   ├── identity_test.go
│   │   ├── logger.go
│   │   ├── logger_test.go
│   │   ├── optional_api_key.go
│   │   ├── optional_api_key_test.go
│   │   ├── rate_limit.go
│   │   ├── rate_limit_test.go
│   │   ├── recovery.go
│   │   ├── recovery_test.go
│   │   ├── request_id.go
│   │   ├── request_id_test.go
│   │   ├── security_headers.go
│   │   ├── security_headers_test.go
│   │   ├── timeout.go
│   │   ├── timeout_test.go
│   │   ├── usage_tracking.go
│   │   └── usage_tracking_test.go
│   ├── models/
│   │   ├── account.go
│   │   ├── api_key.go
│   │   ├── api_request.go
│   │   ├── auth.go
│   │   ├── colleges_of_education_test.go
│   │   ├── colleges_of_agriculture_test.go
│   │   ├── colleges_of_health_sciences_and_technology_test.go
│   │   ├── colleges_of_nursing_and_midwifery_test.go
│   │   ├── countries_and_areas_test.go
│   │   ├── currencies_test.go
│   │   ├── dataset.go
│   │   ├── dataset_public_test.go
│   │   ├── dataset_source.go
│   │   ├── dataset_version.go
│   │   ├── education.go
│   │   ├── finance.go
│   │   ├── microfinance_banks.go
│   │   ├── microfinance_banks_test.go
│   │   ├── commercial_banks.go
│   │   ├── commercial_banks_test.go
│   │   ├── regulated_finance.go
│   │   ├── regulated_finance_test.go
│   │   ├── finance_imto_test.go
│   │   ├── finance_test.go
│   │   ├── geography.go
│   │   ├── geography_test.go
│   │   ├── lgas_test.go
│   │   ├── session.go
│   │   ├── monotechnics_test.go
│   │   ├── polytechnics_test.go
│   │   ├── primary_and_secondary_schools_test.go
│   │   ├── technical_colleges_test.go
│   │   ├── time_zones_test.go
│   │   ├── universities_test.go
│   │   ├── usage_summary.go
│   │   ├── usage_summary_test.go
│   │   └── vocational_enterprise_institutions_test.go
│   ├── redis/
│   │   ├── client.go
│   │   └── client_test.go
│   ├── repository/
│   │   ├── file/
│   │   │   ├── countries_and_areas_test.go
│   │   │   ├── csv_repository.go
│   │   │   ├── csv_repository_test.go
│   │   │   ├── education_colleges_repository.go
│   │   │   ├── education_colleges_repository_test.go
│   │   │   ├── education_bench_test.go
│   │   │   ├── education_cache.go
│   │   │   ├── education_datasets.go
│   │   │   ├── education_extended_repository_test.go
│   │   │   ├── education_repository.go
│   │   │   ├── education_repository_test.go
│   │   │   ├── education_schools.go
│   │   │   ├── education_smalls.go
│   │   │   ├── finance_commercial_banks_test.go
│   │   │   ├── finance_currency_test.go
│   │   │   ├── finance_microfinance_banks.go
│   │   │   ├── finance_microfinance_banks_test.go
│   │   │   ├── finance_regulated.go
│   │   │   ├── finance_regulated_test.go
│   │   │   ├── finance_repository.go
│   │   │   ├── finance_repository_test.go
│   │   │   ├── geography_languages.go
│   │   │   ├── geography_languages_test.go
│   │   │   ├── geography_repository.go
│   │   │   ├── geography_repository_test.go
│   │   │   ├── geography_time_zones.go
│   │   │   ├── geojson_repository.go
│   │   │   ├── geojson_repository_test.go
│   │   │   ├── json_repository.go
│   │   │   ├── json_repository_test.go
│   │   │   ├── store.go
│   │   │   ├── store_test.go
│   │   │   └── time_zones_test.go
│   │   ├── interfaces/
│   │   │   ├── account_repository.go
│   │   │   ├── api_key_repository.go
│   │   │   ├── dataset_repository.go
│   │   │   ├── education_repository.go
│   │   │   ├── errors.go
│   │   │   ├── file_repository.go
│   │   │   ├── finance_repository.go
│   │   │   ├── geography_repository.go
│   │   │   ├── rate_limit_repository.go
│   │   │   ├── session_repository.go
│   │   │   └── usage_repository.go
│   │   ├── postgres/
│   │   │   ├── account_repository.go
│   │   │   ├── api_key_repository.go
│   │   │   ├── dataset_repository.go
│   │   │   ├── dataset_repository_test.go
│   │   │   ├── mappers.go
│   │   │   ├── mappers_test.go
│   │   │   ├── session_repository.go
│   │   │   ├── usage_repository.go
│   │   │   └── usage_repository_test.go
│   │   └── redis/
│   │       ├── rate_limit_repository.go
│   │       └── rate_limit_repository_test.go
│   ├── response/
│   │   ├── errors.go
│   │   ├── errors_test.go
│   │   ├── finance_response_test.go
│   │   ├── geography_languages_response_test.go
│   │   ├── pagination.go
│   │   ├── response.go
│   │   └── response_test.go
│   ├── router/
│   │   ├── bank_assets.go
│   │   ├── bank_assets_test.go
│   │   ├── account_routes.go
│   │   ├── account_routes_test.go
│   │   ├── auth_routes.go
│   │   ├── auth_routes_test.go
│   │   ├── dataset_routes.go
│   │   ├── dataset_routes_test.go
│   │   ├── flag_assets.go
│   │   ├── flag_assets_test.go
│   │   ├── http_router.go
│   │   ├── http_router_test.go
│   │   ├── public_routes.go
│   │   ├── public_routes_test.go
│   │   ├── regulated_finance_assets.go
│   │   ├── regulated_finance_routes.go
│   │   ├── regulated_finance_routes_test.go
│   │   ├── route_catalog.go
│   │   ├── router.go
│   │   └── router_test.go
│   ├── security/
│   │   ├── anonymous_id.go
│   │   ├── api_key.go
│   │   ├── password.go
│   │   ├── random.go
│   │   ├── security_test.go
│   │   └── token.go
│   ├── services/
│   │   ├── account_service.go
│   │   ├── account_service_test.go
│   │   ├── api_key_service.go
│   │   ├── api_key_service_test.go
│   │   ├── auth_service.go
│   │   ├── auth_service_test.go
│   │   ├── countries_and_areas_test.go
│   │   ├── dataset_service.go
│   │   ├── dataset_service_test.go
│   │   ├── education_colleges_service.go
│   │   ├── education_colleges_service_test.go
│   │   ├── education_extended_service.go
│   │   ├── education_extended_service_test.go
│   │   ├── education_service.go
│   │   ├── education_service_test.go
│   │   ├── errors.go
│   │   ├── finance_service.go
│   │   ├── finance_service_test.go
│   │   ├── finance_commercial_banks_test.go
│   │   ├── finance_microfinance_banks_test.go
│   │   ├── geography_country_profile.go
│   │   ├── geography_country_profile_test.go
│   │   ├── geography_languages.go
│   │   ├── geography_service.go
│   │   ├── geography_service_test.go
│   │   ├── geography_time_zones.go
│   │   ├── time_zones_test.go
│   │   ├── usage_service.go
│   │   └── usage_service_test.go
│   └── validators/
│       ├── account_validator.go
│       ├── account_validator_test.go
│       ├── api_key_validator.go
│       ├── api_key_validator_test.go
│       ├── auth_validator.go
│       ├── auth_validator_test.go
│       ├── dataset_validator.go
│       ├── dataset_validator_test.go
│       ├── education_colleges_validator_test.go
│       ├── education_extended_validator.go
│       ├── education_extended_validator_test.go
│       ├── education_validator.go
│       ├── education_validator_test.go
│       ├── finance_validator.go
│       ├── finance_validator_test.go
│       ├── finance_commercial_banks_test.go
│       ├── microfinance_bank_validator_test.go
│       ├── geography_languages_validator_test.go
│       ├── geography_validator.go
│       ├── geography_validator_test.go
│       ├── query_validator.go
│       └── query_validator_test.go
├── tools/
│   ├── generate_education_snapshots.py
│   ├── generate_health_facilities.py
│   └── generate_primary_and_secondary_schools.py
```

## What Each Area Does

### `cmd/`

Executable entry points for the API server and future tooling.

### `internal/app/`

Application startup, dependency construction, dataset verification, and graceful shutdown.

### `internal/config/`

Environment-driven application configuration for the server, database, security, rate limits, Redis, usage tracking, and dataset paths.

### `datasets/`

Versioned geography, education, healthcare and finance datasets, schemas, provenance metadata, reconciliation manifests, licensing notes, and embedded runtime assets. `embedded.go` embeds the JSON dataset directories used when a deployment cannot provide the configured filesystem dataset path. Education includes paginated primary and secondary schools plus institution snapshots for universities, colleges of education, polytechnics, monotechnics, colleges of agriculture, health sciences and technology, nursing and midwifery, technical colleges, and vocational enterprise institutions. The healthcare foundation contains the source-verified unified health-facility snapshot and deterministic state/FCT reconciliation partitions; it is not runtime-integrated in this pass. Regulated-finance logo provenance is recorded in `datasets/assets/financial-institutions/ng/ATTRIBUTION.md`; the FHA Homes asset is explicitly a Federal Housing Authority parent-brand representative mark.

### `internal/database/`

PostgreSQL pool creation, readiness checks, and generated sqlc persistence code.

### `internal/handlers/`

HTTP handlers that validate requests, call services, and produce shared response envelopes. Education handlers are split between the existing education handler and `education_extended_handler.go`, which serves the additional institution categories and paginated schools.

### `internal/middlewares/`

HTTP cross-cutting concerns such as request IDs, logging, recovery, timeouts, CORS, security headers, body limits, authentication, rate limiting, and usage tracking.

### `internal/models/`

Domain and API-facing models kept separate from sqlc-generated persistence structs. `healthcare.go` defines the unified health-facility snapshot model, with `health_facilities_test.go` validating its dataset and reconciliation package.

### `internal/repository/`

Repository interfaces plus PostgreSQL, Redis, and file-backed implementations. The file repository includes lazy, synchronized, validated loading for education snapshots, paginated school indexing, and dedicated validation/loading for regulated-finance datasets.

### Runtime/Data Loading

- Local development reads JSON files from `DATASETS_PATH` (normally `datasets/`).
- Production/serverless startup can fall back to the embedded JSON filesystem exposed by `datasets.Files()` when the configured path is unavailable.
- Flag assets and regulated-finance logo assets are embedded by `datasets/assets`; public asset handlers read them through category- and identifier-scoped helpers and expose only the supported public formats.
- The source tree includes `248` flag SVGs, `28` commercial-bank PNGs, and `422` regulated-finance PNGs. The wildcard entries in the tree represent those complete inventories; together the finance asset trees contain `450` PNG files.

### `internal/services/`

Application use cases and business rules for accounts, authentication, datasets, education, finance, geography, and usage. Education services expose list/detail operations for institution snapshots and paginated filtering/detail lookup for primary and secondary schools.

### `internal/validators/`

Request validation and normalization helpers for authentication, accounts, API keys, datasets, geography, education institution IDs, school IDs, pagination, filters, and query inputs.

### `internal/router/`

HTTP router construction, route registration, public and authenticated route groups, route cataloging, and embedded flag, commercial-bank, and regulated-finance logo serving.

### `internal/redis/`

Redis client setup and low-level access helpers used by repository implementations.

### `internal/security/`

Token, password, API key, random-value, and anonymous-identifier helpers.

### `internal/response/`

Shared HTTP response, pagination, and error formatting helpers.

### `database/`

Database schema migrations and handwritten SQL query files used to generate persistence code.

### `docs/`

User-facing and contributor-facing documentation, including the OpenAPI specification and this structure reference.

### `tools/`

Dataset generation and snapshot tooling. `generate_education_snapshots.py` builds verified institution snapshots, `generate_primary_and_secondary_schools.py` builds the partitioned UBEC school dataset and reconciliation summaries, and `generate_health_facilities.py` builds the deterministic GRID3/NHFR-derived health-facility snapshot and state/FCT manifests. Generators are not run during API startup.

### `tmp/`

Local scratch outputs used during development and verification. This directory is not required at runtime.

## Design Notes

### UBEC School Snapshot

The education data tree contains `datasets/education/primary_and_secondary_schools.json`, its Draft 2020-12 schema, source metadata, checkpoint summary, and 37-file state/FCT reconciliation index under `datasets/metadata/education/primary_and_secondary_schools_reconciliation/`. The generator is `tools/generate_primary_and_secondary_schools.py`. This is a source-observed UBEC 2022 snapshot exposed through a paginated repository, service, handler, and production HTTP route.

- Public dataset access stays anonymous by default.
- Optional API keys add higher limits and usage analytics.
- Configuration and dependencies are constructed once at startup and passed down explicitly.
- PostgreSQL readiness is separated from HTTP liveness.
- sqlc-generated persistence models stay isolated from `internal/models`.
- Dataset readers use explicit paths, bounded file access, validation, and defensive copies before data reaches handlers.
