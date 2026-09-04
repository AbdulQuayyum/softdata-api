# API Structure

SoftData API uses a layered Go layout that keeps entry points, configuration, persistence, domain models, HTTP handling, and documentation separated.

## Overview

- `cmd/` contains executable entry points.
- `internal/` contains application configuration, infrastructure, domain models, repositories, services, HTTP handlers, middleware, routing, validation, security, and response helpers.
- `datasets/` contains versioned source data, schemas, provenance metadata, licensing notes, and embedded flag assets.
- `database/` contains PostgreSQL migrations and handwritten SQL queries.
- `docs/` contains API and project documentation, including the OpenAPI contract.
- Root files provide development configuration, build metadata, contribution guidance, and project licensing.
- `api` and `tmp/` are generated or local development artifacts and are not application source files.

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
│   ├── assets/
│   │   ├── flags.go
│   │   ├── flags_test.go
│   │   └── flags/
│   │       ├── ATTRIBUTION.md
│   │       ├── LICENSE
│   │       └── 4x3/
│   │           └── *.svg                  # 248 vendored flag assets
│   ├── banks.go
│   └── banks/
│       └── ng/
│           ├── ATTRIBUTION.md
│           ├── LICENSES/
│           │   └── Nigerian-Bank-Logos-MIT.txt
│           └── *.png                      # 28 vendored bank assets
│   ├── education/
│   │   ├── colleges_of_education.json
│   │   └── universities.json
│   ├── finance/
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
│   │   │   └── universities.json
│   │   ├── finance/
│   │   │   ├── commercial_banks.json
│   │   │   ├── currencies.json
│   │   │   ├── international_money_transfer_operators.json
│   │   │   └── payment_service_providers.json
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
│       │   └── universities.schema.json
│       ├── finance/
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
│   └── versioning.md
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   ├── app_test.go
│   │   ├── dependencies.go
│   │   ├── education_test.go
│   │   ├── finance_test.go
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
│   │   ├── finance_handler.go
│   │   ├── finance_handler_test.go
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
│   │   ├── countries_and_areas_test.go
│   │   ├── currencies_test.go
│   │   ├── dataset.go
│   │   ├── dataset_public_test.go
│   │   ├── dataset_source.go
│   │   ├── dataset_version.go
│   │   ├── education.go
│   │   ├── finance.go
│   │   ├── finance_imto_test.go
│   │   ├── finance_test.go
│   │   ├── geography.go
│   │   ├── geography_test.go
│   │   ├── lgas_test.go
│   │   ├── session.go
│   │   ├── time_zones_test.go
│   │   ├── universities_test.go
│   │   ├── usage_summary.go
│   │   └── usage_summary_test.go
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
│   │   │   ├── education_repository.go
│   │   │   ├── education_repository_test.go
│   │   │   ├── finance_currency_test.go
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
│   │   │   └── store_test.go
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
│   │   ├── education_service.go
│   │   ├── education_service_test.go
│   │   ├── errors.go
│   │   ├── finance_service.go
│   │   ├── finance_service_test.go
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
│       ├── education_validator.go
│       ├── education_validator_test.go
│       ├── finance_validator.go
│       ├── finance_validator_test.go
│       ├── geography_languages_validator_test.go
│       ├── geography_validator.go
│       ├── geography_validator_test.go
│       ├── query_validator.go
│       └── query_validator_test.go
└── tmp/
    ├── build-errors.log
    └── main
```

## What Each Area Does

### `cmd/`

Executable entry points for the API server and future tooling.

### `internal/app/`

Application startup, dependency construction, dataset verification, and graceful shutdown.

### `internal/config/`

Environment-driven application configuration for the server, database, security, rate limits, Redis, usage tracking, and dataset paths.

### `datasets/`

Versioned geography, education, and finance datasets, schemas, provenance metadata, licensing notes, and embedded flag assets.

### `internal/database/`

PostgreSQL pool creation, readiness checks, and generated sqlc persistence code.

### `internal/handlers/`

HTTP handlers that validate requests, call services, and produce shared response envelopes.

### `internal/middlewares/`

HTTP cross-cutting concerns such as request IDs, logging, recovery, timeouts, CORS, security headers, body limits, authentication, rate limiting, and usage tracking.

### `internal/models/`

Domain and API-facing models kept separate from sqlc-generated persistence structs.

### `internal/repository/`

Repository interfaces plus PostgreSQL, Redis, and file-backed implementations.

### `internal/services/`

Application use cases and business rules for accounts, authentication, datasets, education, finance, geography, and usage.

### `internal/validators/`

Request validation and normalization helpers for authentication, accounts, API keys, datasets, geography, and query inputs.

### `internal/router/`

HTTP router construction, route registration, public and authenticated route groups, route cataloging, and embedded flag and bank-logo serving.

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

### `tmp/`

Local scratch outputs used during development and verification. This directory is not required at runtime.

## Design Notes

- Public dataset access stays anonymous by default.
- Optional API keys add higher limits and usage analytics.
- Configuration and dependencies are constructed once at startup and passed down explicitly.
- PostgreSQL readiness is separated from HTTP liveness.
- sqlc-generated persistence models stay isolated from `internal/models`.
- Dataset readers use explicit paths, bounded file access, validation, and defensive copies before data reaches handlers.
