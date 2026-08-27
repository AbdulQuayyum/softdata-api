# Rate Limits

SoftData uses rate limits to protect service availability.

## Default Limits

| Access                    |                           Limit |
| ------------------------- | ------------------------------: |
| Anonymous                 |   60 requests per minute per IP |
| API key                   | 300 requests per minute per key |
| API-key monthly allowance |     50,000 requests per account |
| Dataset downloads         |          10 requests per minute |

## Response Headers

```http
RateLimit-Limit: 300
RateLimit-Remaining: 281
RateLimit-Reset: 1787756400
```

## Exceeded Limit

SoftData returns `429 Too Many Requests`:

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "The request limit has been exceeded.",
    "retry_after_seconds": 30
  }
}
```

A `Retry-After` header may also be included.

## Fair Use

Applications should:

- Cache stable reference data.
- Avoid repeatedly downloading unchanged datasets.
- Use pagination.
- Request only necessary fields.
- Use bulk downloads when the entire dataset is needed.
- Avoid aggressive retry loops.

Rate-limit policies may change as traffic and infrastructure evolve.

The `remaining_allowance` value returned by usage summary endpoints is the monthly account allowance. It is separate from the short-window `RateLimit-*` headers and should not be treated as the same counter.
