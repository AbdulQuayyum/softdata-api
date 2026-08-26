# API Structure

SoftData API uses a layered Go layout that keeps HTTP concerns, business logic, data access, and dataset assets cleanly separated.

## Overview

- `cmd/` holds executable entry points.
- `internal/` contains the core application logic.
- `external/` wraps third-party integrations.
- `datasets/` stores versioned data, schemas, and metadata.
- `database/` contains migrations, SQL, and seed data.
- `workers/` runs background jobs and maintenance tasks.
- `docs/` holds API and project documentation.

## Project Tree

```text
softdata-api/
├── cmd/
│   ├── api/
│   │   └── main.go
│   ├── worker/
│   │   └── main.go
│   └── cli/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   ├── server.go
│   │   ├── database.go
│   │   ├── redis.go
│   │   ├── security.go
│   │   ├── rate_limit.go
│   │   └── datasets.go
│   ├── router/
│   │   ├── router.go
│   │   ├── public_routes.go
│   │   ├── auth_routes.go
│   │   ├── account_routes.go
│   │   └── dataset_routes.go
│   ├── handlers/
│   │   ├── health_handler.go
│   │   ├── discovery_handler.go
│   │   ├── auth_handler.go
│   │   ├── account_handler.go
│   │   ├── api_key_handler.go
│   │   ├── usage_handler.go
│   │   ├── dataset_handler.go
│   │   ├── geography_handler.go
│   │   ├── finance_handler.go
│   │   ├── education_handler.go
│   │   ├── healthcare_handler.go
│   │   ├── emergency_handler.go
│   │   ├── infrastructure_handler.go
│   │   └── statistics_handler.go
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── account_service.go
│   │   ├── api_key_service.go
│   │   ├── usage_service.go
│   │   ├── dataset_service.go
│   │   ├── geography_service.go
│   │   ├── finance_service.go
│   │   ├── education_service.go
│   │   ├── healthcare_service.go
│   │   ├── emergency_service.go
│   │   ├── infrastructure_service.go
│   │   ├── statistics_service.go
│   │   ├── search_service.go
│   │   ├── download_service.go
│   │   ├── cache_service.go
│   │   └── import_service.go
│   ├── repository/
│   │   ├── interfaces/
│   │   │   ├── account_repository.go
│   │   │   ├── session_repository.go
│   │   │   ├── api_key_repository.go
│   │   │   ├── usage_repository.go
│   │   │   ├── dataset_repository.go
│   │   │   ├── geography_repository.go
│   │   │   ├── finance_repository.go
│   │   │   ├── education_repository.go
│   │   │   ├── healthcare_repository.go
│   │   │   ├── emergency_repository.go
│   │   │   ├── infrastructure_repository.go
│   │   │   └── statistics_repository.go
│   │   ├── postgres/
│   │   │   ├── account_repository.go
│   │   │   ├── session_repository.go
│   │   │   ├── api_key_repository.go
│   │   │   ├── usage_repository.go
│   │   │   ├── dataset_repository.go
│   │   │   ├── geography_repository.go
│   │   │   ├── finance_repository.go
│   │   │   ├── education_repository.go
│   │   │   ├── healthcare_repository.go
│   │   │   ├── emergency_repository.go
│   │   │   ├── infrastructure_repository.go
│   │   │   └── statistics_repository.go
│   │   ├── file/
│   │   │   ├── json_repository.go
│   │   │   ├── csv_repository.go
│   │   │   └── geojson_repository.go
│   │   └── redis/
│   │       ├── cache_repository.go
│   │       ├── session_repository.go
│   │       └── rate_limit_repository.go
│   ├── models/
│   │   ├── account.go
│   │   ├── session.go
│   │   ├── api_key.go
│   │   ├── api_request.go
│   │   ├── usage_summary.go
│   │   ├── dataset.go
│   │   ├── dataset_source.go
│   │   ├── dataset_version.go
│   │   ├── geography.go
│   │   ├── finance.go
│   │   ├── education.go
│   │   ├── healthcare.go
│   │   ├── emergency.go
│   │   ├── infrastructure.go
│   │   └── statistics.go
│   ├── middlewares/
│   │   ├── recovery.go
│   │   ├── request_id.go
│   │   ├── logger.go
│   │   ├── security_headers.go
│   │   ├── cors.go
│   │   ├── body_limit.go
│   │   ├── timeout.go
│   │   ├── authentication.go
│   │   ├── optional_api_key.go
│   │   ├── rate_limit.go
│   │   └── usage_tracking.go
│   ├── validators/
│   │   ├── auth_validator.go
│   │   ├── account_validator.go
│   │   ├── api_key_validator.go
│   │   ├── query_validator.go
│   │   ├── pagination_validator.go
│   │   └── dataset_validator.go
│   ├── security/
│   │   ├── password.go
│   │   ├── token.go
│   │   ├── api_key.go
│   │   ├── anonymous_id.go
│   │   └── random.go
│   ├── response/
│   │   ├── response.go
│   │   ├── pagination.go
│   │   ├── metadata.go
│   │   └── errors.go
│   ├── cache/
│   │   ├── keys.go
│   │   ├── ttl.go
│   │   └── invalidation.go
│   └── constants/
│       ├── errors.go
│       ├── datasets.go
│       ├── limits.go
│       └── context.go
├── external/
│   ├── clients/
│   │   ├── http_client.go
│   │   ├── retry.go
│   │   └── response.go
├── datasets/
│   ├── geography/
│   │   ├── countries.json
│   │   ├── states.json
│   │   ├── lgas.json
│   │   ├── wards.json
│   │   ├── cities.json
│   │   ├── geopolitical_zones.json
│   │   ├── postal_codes.json
│   │   ├── plate_numbers.json
│   │   └── boundaries.geojson
│   ├── finance/
│   │   ├── institutions.json
│   │   ├── bank_codes.json
│   │   ├── ussd_codes.json
│   │   ├── payment_channels.json
│   │   └── currencies.json
│   ├── education/
│   │   ├── universities.json
│   │   ├── polytechnics.json
│   │   ├── colleges.json
│   │   └── courses.json
│   ├── healthcare/
│   │   ├── facilities.json
│   │   ├── facility_types.json
│   │   └── services.json
│   ├── emergency/
│   │   ├── emergency_numbers.json
│   │   ├── police_commands.json
│   │   ├── fire_stations.json
│   │   └── road_safety_commands.json
│   ├── infrastructure/
│   │   ├── airports.json
│   │   ├── seaports.json
│   │   ├── railway_stations.json
│   │   ├── electricity_discos.json
│   │   └── phone_prefixes.json
│   ├── statistics/
│   │   ├── population.json
│   │   ├── inflation.json
│   │   ├── food_prices.json
│   │   ├── fuel_prices.json
│   │   └── economic_indicators.json
│   ├── schemas/
│   │   ├── geography.schema.json
│   │   ├── finance.schema.json
│   │   ├── education.schema.json
│   │   ├── healthcare.schema.json
│   │   ├── emergency.schema.json
│   │   ├── infrastructure.schema.json
│   │   └── statistics.schema.json
│   └── metadata/
│       ├── datasets.json
│       ├── sources.json
│       ├── licences.json
│       └── versions.json
├── database/
│   ├── migrations/
│   │   ├── 000001_create_accounts.up.sql
│   │   ├── 000001_create_accounts.down.sql
│   │   ├── 000002_create_sessions.up.sql
│   │   ├── 000002_create_sessions.down.sql
│   │   ├── 000003_create_api_keys.up.sql
│   │   ├── 000003_create_api_keys.down.sql
│   │   ├── 000004_create_api_requests.up.sql
│   │   ├── 000004_create_api_requests.down.sql
│   │   ├── 000005_create_usage_daily.up.sql
│   │   ├── 000005_create_usage_daily.down.sql
│   │   ├── 000006_create_datasets.up.sql
│   │   ├── 000006_create_datasets.down.sql
│   │   ├── 000007_create_dataset_sources.up.sql
│   │   ├── 000007_create_dataset_sources.down.sql
│   │   ├── 000008_create_dataset_versions.up.sql
│   │   └── 000008_create_dataset_versions.down.sql
│   ├── queries/
│   │   ├── accounts.sql
│   │   ├── sessions.sql
│   │   ├── api_keys.sql
│   │   ├── usage.sql
│   │   └── datasets.sql
│   └── seeds/
│       └── datasets.sql
├── workers/
│   ├── dataset_importer.go
│   ├── dataset_validator.go
│   ├── dataset_refresher.go
│   ├── cache_warmer.go
│   ├── usage_aggregator.go
│   ├── request_log_cleaner.go
│   └── expired_session_cleaner.go
├── utilities/
│   ├── pagination.go
│   ├── slug.go
│   ├── file.go
│   ├── checksum.go
│   ├── dates.go
│   └── coordinates.go
├── scripts/
│   ├── import-datasets.sh
│   ├── validate-datasets.sh
│   ├── generate-checksums.sh
│   └── seed-database.sh
├── docs/
│   ├── api-structure.md
│   ├── openapi.yaml
│   ├── quick-start.md
│   ├── authentication.md
│   ├── api-keys.md
│   ├── rate-limits.md
│   ├── datasets.md
│   ├── versioning.md
│   └── errors.md
├── tests/
│   ├── unit/
│   │   ├── services/
│   │   ├── middlewares/
│   │   └── validators/
│   ├── integration/
│   │   ├── auth_test.go
│   │   ├── api_key_test.go
│   │   ├── dataset_test.go
│   │   └── usage_test.go
│   ├── fixtures/
│   └── testdata/
├── .env.example
├── .gitignore
├── .golangci.yml
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── CONTRIBUTING.md
├── DATASETS.md
├── SECURITY.md
├── CODE_OF_CONDUCT.md
├── LICENSE
└── CHANGELOG.md
```

## What Each Area Does

### `cmd/`

Entry points for the API server, workers, and CLI tools.

### `internal/`

The application core. This is where request handling, service logic, storage access, validation, security, and shared helpers live.

### `external/`

Adapters for external clients, kept separate so third-party integrations do not leak into the core domain.

### `datasets/`

Source-controlled data assets, schemas, and metadata for the public API.

### `database/`

Database schema management, reusable queries, and seed files.

### `workers/`

Asynchronous or scheduled jobs that should not run inside request handlers.

### `utilities/`

Small reusable helpers that are not tied to HTTP or storage concerns.

### `scripts/`

Operational scripts for imports, validation, checksums, and seeding.

### `docs/`

User-facing and contributor-facing documentation, including the OpenAPI spec.

### `tests/`

Unit, integration, and test fixtures.

## Request Flow

```text
request -> router -> middleware -> handler -> service -> repository -> storage/source
```

## Design Notes

- Public dataset access stays anonymous by default.
- Optional API keys add higher limits and usage analytics.
- Account-only features remain separate from public dataset routes.
- Background data processing belongs in workers, not handlers.
- Repository interfaces keep data storage swappable.

## Why This Layout Works

- It keeps the API easy to navigate.
- It reduces coupling between layers.
- It makes dataset growth manageable.
- It supports the hybrid anonymous-plus-account access model.
- It leaves room for background jobs, versioning, and future integrations.
