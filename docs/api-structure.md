# API Structure

SoftData API uses a layered Go layout that keeps entry points, configuration, database access, domain models, and documentation separated.

## Overview

- `cmd/` holds executable entry points.
- `internal/` contains the core application logic, including configuration, database access, repositories, security helpers, request validators, response helpers, and domain/API models.
- `database/` contains PostgreSQL migrations and handwritten SQL queries.
- `docs/` holds API and project documentation.
- Root tooling files such as `Makefile`, `sqlc.yaml`, `.air.toml`, and `.env.example` support local development and database generation.

## Project Tree

```text
softdata-api/
├── .air.toml
├── .env.example
├── .gitignore
├── api
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── CONTRIBUTING.md
├── DATASETS.md
├── LICENSE
├── Makefile
├── README.md
├── SECURITY.md
├── cmd
│   └── api
│       └── main.go
├── database
│   ├── migrations
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
│   └── queries
│       ├── accounts.sql
│       ├── api_keys.sql
│       ├── datasets.sql
│       ├── sessions.sql
│       └── usage.sql
├── datasets
│   ├── geography
│   │   └── states.json
│   ├── LICENSE.md
│   ├── metadata
│   │   └── geography
│   │       └── states.json
│   ├── README.md
│   └── schemas
│       └── geography
│           └── states.schema.json
├── docs
│   ├── api-keys.md
│   ├── api-structure.md
│   ├── authentication.md
│   ├── datasets.md
│   ├── errors.md
│   ├── openapi.yaml
│   ├── quick-start.md
│   ├── rate-limits.md
│   └── versioning.md
├── go.mod
├── go.sum
├── internal
│   ├── app
│   │   ├── app.go
│   │   ├── app_test.go
│   │   ├── dependencies.go
│   │   ├── geography_test.go
│   │   └── shutdown.go
│   ├── config
│   │   ├── config.go
│   │   ├── config_test.go
│   │   ├── database.go
│   │   ├── datasets.go
│   │   ├── rate_limit.go
│   │   ├── redis.go
│   │   ├── security.go
│   │   ├── server.go
│   │   └── usage.go
│   ├── database
│   │   ├── health.go
│   │   ├── postgres.go
│   │   ├── postgres_test.go
│   │   └── sqlc
│   │       ├── accounts.sql.go
│   │       ├── api_keys.sql.go
│   │       ├── datasets.sql.go
│   │       ├── db.go
│   │       ├── models.go
│   │       ├── querier.go
│   │       ├── sessions.sql.go
│   │       ├── sessions_rotation_test.go
│   │       └── usage.sql.go
│   ├── handlers
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
│   │   ├── geography_handler.go
│   │   ├── geography_handler_test.go
│   │   ├── health_handler.go
│   │   ├── health_handler_test.go
│   │   ├── usage_handler.go
│   │   └── usage_handler_test.go
│   ├── middlewares
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
│   ├── models
│   │   ├── account.go
│   │   ├── api_key.go
│   │   ├── api_request.go
│   │   ├── auth.go
│   │   ├── dataset.go
│   │   ├── dataset_public_test.go
│   │   ├── dataset_source.go
│   │   ├── dataset_version.go
│   │   ├── geography.go
│   │   ├── geography_test.go
│   │   ├── session.go
│   │   ├── usage_summary.go
│   │   └── usage_summary_test.go
│   ├── redis
│   │   ├── client.go
│   │   └── client_test.go
│   ├── repository
│   │   ├── file
│   │   │   ├── csv_repository.go
│   │   │   ├── csv_repository_test.go
│   │   │   ├── geography_repository.go
│   │   │   ├── geography_repository_test.go
│   │   │   ├── geojson_repository.go
│   │   │   ├── geojson_repository_test.go
│   │   │   ├── json_repository.go
│   │   │   ├── json_repository_test.go
│   │   │   ├── store.go
│   │   │   └── store_test.go
│   │   ├── interfaces
│   │   │   ├── account_repository.go
│   │   │   ├── api_key_repository.go
│   │   │   ├── dataset_repository.go
│   │   │   ├── errors.go
│   │   │   ├── file_repository.go
│   │   │   ├── geography_repository.go
│   │   │   ├── rate_limit_repository.go
│   │   │   ├── session_repository.go
│   │   │   └── usage_repository.go
│   │   ├── postgres
│   │   │   ├── account_repository.go
│   │   │   ├── api_key_repository.go
│   │   │   ├── dataset_repository.go
│   │   │   ├── dataset_repository_test.go
│   │   │   ├── mappers.go
│   │   │   ├── mappers_test.go
│   │   │   ├── session_repository.go
│   │   │   ├── usage_repository.go
│   │   │   └── usage_repository_test.go
│   │   └── redis
│   │       ├── rate_limit_repository.go
│   │       └── rate_limit_repository_test.go
│   ├── response
│   │   ├── errors.go
│   │   ├── errors_test.go
│   │   ├── pagination.go
│   │   ├── response.go
│   │   └── response_test.go
│   ├── router
│   │   ├── account_routes.go
│   │   ├── account_routes_test.go
│   │   ├── auth_routes.go
│   │   ├── auth_routes_test.go
│   │   ├── dataset_routes.go
│   │   ├── dataset_routes_test.go
│   │   ├── http_router.go
│   │   ├── http_router_test.go
│   │   ├── public_routes.go
│   │   ├── public_routes_test.go
│   │   ├── route_catalog.go
│   │   ├── router.go
│   │   └── router_test.go
│   ├── security
│   │   ├── anonymous_id.go
│   │   ├── api_key.go
│   │   ├── password.go
│   │   ├── random.go
│   │   ├── security_test.go
│   │   └── token.go
│   ├── services
│   │   ├── account_service.go
│   │   ├── account_service_test.go
│   │   ├── api_key_service.go
│   │   ├── api_key_service_test.go
│   │   ├── auth_service.go
│   │   ├── auth_service_test.go
│   │   ├── dataset_service.go
│   │   ├── dataset_service_test.go
│   │   ├── errors.go
│   │   ├── geography_service.go
│   │   ├── geography_service_test.go
│   │   ├── usage_service.go
│   │   └── usage_service_test.go
│   └── validators
│       ├── account_validator.go
│       ├── account_validator_test.go
│       ├── api_key_validator.go
│       ├── api_key_validator_test.go
│       ├── auth_validator.go
│       ├── auth_validator_test.go
│       ├── dataset_validator.go
│       ├── dataset_validator_test.go
│       ├── geography_validator.go
│       ├── geography_validator_test.go
│       ├── query_validator.go
│       └── query_validator_test.go
├── sqlc.yaml
└── tmp
    ├── build-errors.log
    └── main
```

## What Each Area Does

### `cmd/`

Executable entry points for the API server and future tooling.

### `internal/config/`

Environment-driven application configuration, including server, database, security, rate-limit, and dataset path settings.

### `internal/database/`

PostgreSQL pool creation, readiness checks, and generated sqlc output.

### `internal/handlers/`

HTTP handlers that bridge routes to services and shared response helpers.

### `internal/models/`

Domain and API-facing models that stay separate from sqlc-generated persistence structs.

### `internal/repository/`

Repository interfaces plus PostgreSQL, Redis, and file-backed implementations that bridge services to persistence and dataset assets.

### `internal/repository/file/`

Safe file-backed dataset readers for JSON, CSV, and GeoJSON content, plus shared path and size guards.

### `internal/redis/`

Redis client setup and low-level access helpers used by repository implementations.

### `internal/security/`

Token, password, API key, and anonymous-identifier helpers used by services and repositories.

### `internal/response/`

Shared HTTP response and error formatting helpers.

### `internal/validators/`

Request validation and normalization helpers for auth, account, API key, dataset, and query inputs.

### `database/`

Database schema migrations and handwritten SQL query files.

### `docs/`

User-facing and contributor-facing documentation, including the OpenAPI spec.

## Design Notes

- Public dataset access stays anonymous by default.
- Optional API keys add higher limits and usage analytics.
- Configuration is loaded once at startup and passed down explicitly.
- PostgreSQL readiness is separated from HTTP liveness.
- sqlc-generated persistence models stay isolated from `internal/models`.
