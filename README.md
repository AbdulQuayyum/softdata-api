# SoftData API

SoftData is an open-source API providing structured, normalized and developer-ready public datasets.

The hosted API supports immediate anonymous access. Developers who need higher request limits and personal usage analytics can create a free account and generate an optional API key.

## Features

- No signup required for public dataset access
- Optional API keys with higher limits
- Account-level usage analytics
- JSON, CSV and GeoJSON datasets
- Search, filtering, sorting and pagination
- Versioned datasets
- Dataset source and licence information
- OpenAPI documentation
- Rate limiting and abuse protection
- Open-source dataset contributions

## Dataset Groups

SoftData is organized into the following groups:

- Geography
- Finance
- Education
- Healthcare
- Emergency services
- Infrastructure
- Statistics

The first release will focus primarily on geography and dataset discovery.

## Quick Start

Make an anonymous request:

```bash
curl https://softdata-api.vercel.app/v1/geography/states
```

No API key is required.

### Using an Optional API Key

```bash
curl \
  -H "X-API-Key: sd_live_your_api_key" \
  https://softdata-api.vercel.app/v1/geography/states
```

An API key provides higher rate limits and account-level usage analytics.

## Base URL

```text
https://softdata-api.vercel.app
```

Current API version:

```text
v1
```

## Example Response

```json
{
  "success": true,
  "data": [
    {
      "code": "NG-KW",
      "name": "Kwara",
      "capital": "Ilorin",
      "country_code": "NG",
      "geopolitical_zone": "North Central"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 37,
    "total_pages": 2
  }
}
```

## Authentication

Authentication is optional for dataset endpoints.

### Anonymous access

Call any public dataset endpoint without an authentication header.

```bash
curl https://softdata-api.vercel.app/v1/geography/states
```

### API-key access

Create a free account, generate an API key and include it in the request:

```http
X-API-Key: sd_live_your_api_key
```

API keys must never be committed to source control or exposed in public repositories.

## Rate Limits

| Access            |    Per minute |  Monthly allowance |
| ----------------- | ------------: | -----------------: |
| Anonymous         |     60 per IP |     Not applicable |
| API key           |   300 per key | 50,000 per account |
| Dataset downloads | 10 per minute |           Included |

Limits may change as the project grows.

Rate-limit information is returned through response headers:

```http
RateLimit-Limit: 300
RateLimit-Remaining: 281
RateLimit-Reset: 1787756400
```

## Public Endpoints

```http
GET /health
GET /v1
GET /v1/datasets
GET /v1/datasets/{dataset_id}
GET /v1/datasets/{dataset_id}/versions
GET /v1/datasets/{dataset_id}/sources
GET /v1/datasets/{dataset_id}/download
```

Dataset endpoints:

```http
GET /v1/geography/*
GET /v1/finance/*
GET /v1/education/*
GET /v1/healthcare/*
GET /v1/emergency/*
GET /v1/infrastructure/*
GET /v1/statistics/*
```

## Account Endpoints

Accounts are only required for API keys and personal analytics.

```http
POST /v1/auth/register
POST /v1/auth/login
POST /v1/auth/refresh
POST /v1/auth/logout

GET    /v1/account
PATCH  /v1/account
DELETE /v1/account

POST   /v1/account/api-keys
GET    /v1/account/api-keys
DELETE /v1/account/api-keys/{key_id}
POST   /v1/account/api-keys/{key_id}/rotate

GET /v1/account/usage
GET /v1/account/usage/daily
GET /v1/account/usage/endpoints
GET /v1/account/usage/dataset-groups
```

## Query Parameters

List endpoints use consistent query parameters:

```text
page
limit
search
sort
order
fields
format
```

Example:

```http
GET /v1/geography/states?page=1&limit=20&sort=name&order=asc
```

The default limit is `20`. The maximum limit is `100`.

## Running Locally

### Requirements

- Go
- PostgreSQL
- Redis, optional

### Clone the repository

```bash
git clone https://github.com/AbdulQuayyum/softdata-api.git
cd softdata-api
```

### Configure the environment

```bash
cp .env.example .env
```

Update the environment variables before running the application.

### Start directly

```bash
go mod download
go run ./cmd/api
```

The API will be available at:

```text
http://localhost:8080
```

Check its health:

```bash
curl http://localhost:8080/health
```

## Dataset Storage

Canonical datasets are stored under:

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

Every dataset must have:

- A stable identifier
- A documented schema
- A source
- A licence
- A version
- A last-verified date
- A maintainer
- Validation checks

See [DATASETS.md](DATASETS.md) before adding or modifying data.

## Documentation

- [Quick Start](docs/quick-start.md)
- [Authentication](docs/authentication.md)
- [API Keys](docs/api-keys.md)
- [Rate Limits](docs/rate-limits.md)
- [Datasets](docs/datasets.md)
- [Data Sources](docs/data-sources.md)
- [Versioning](docs/versioning.md)
- [Errors](docs/errors.md)
- [OpenAPI Specification](docs/openapi.yaml)

## Contributing

Contributions are welcome. You can contribute:

- New datasets
- Dataset corrections
- Additional sources
- Documentation
- Validation rules
- API improvements
- Tests
- Bug fixes

Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

## Security

Do not report security vulnerabilities in public issues.

Follow the private reporting instructions in [SECURITY.md](SECURITY.md).

## Licence

The SoftData source code is licensed under the MIT Licence.

Individual datasets may use different licences. Dataset-specific licences and attribution requirements are documented in:

```text
datasets/metadata/licences.json
```

Using the SoftData source code does not automatically grant rights to datasets supplied by third parties.

## Maintainer

Created and maintained by Abdul-Quayyum Alao.

## Project Status

SoftData is under active development. Endpoints and schemas may change before the stable `v1.0.0` release.
