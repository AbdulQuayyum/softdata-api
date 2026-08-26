# Quick Start

SoftData provides public datasets through a JSON REST API.

## Make Your First Request

No account or API key is required:

```bash
curl https://softdata-api.vercel.app/v1/geography/states
```

You can also open the URL directly in a browser:

```text
https://softdata-api.vercel.app/v1/geography/states
```

## JavaScript

```javascript
const response = await fetch(
  "https://softdata-api.vercel.app/v1/geography/states"
)

if (!response.ok) {
  throw new Error(`Request failed: ${response.status}`)
}

const result = await response.json()

console.log(result.data)
```

## Go

```go
response, err := http.Get(
	"https://softdata-api.vercel.app/v1/geography/states",
)
if err != nil {
	log.Fatal(err)
}
defer response.Body.Close()
```

## Pagination

```http
GET /v1/geography/states?page=1&limit=20
```

## Search

```http
GET /v1/geography/states?search=kwara
```

## Optional API Key

```bash
curl \
  -H "X-API-Key: sd_live_your_api_key" \
  https://softdata-api.vercel.app/v1/geography/states
```

API keys provide higher limits and usage analytics but are not required for normal access.

## Next Steps

- Browse `/v1/datasets`.
- Review the OpenAPI documentation.
- Create an optional developer account.
- Generate a key for higher limits.
- Report incorrect data through GitHub.
