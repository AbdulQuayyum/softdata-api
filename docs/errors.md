# API Errors

SoftData uses a consistent error response format.

## Format

```json
{
  "success": false,
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "The requested resource was not found.",
    "details": null,
    "request_id": "req_01J..."
  }
}
```

Every error response includes a `request_id` so support requests can be traced quickly.

## Error Response Rules

- `success` is always `false` for error responses.
- `error.code` is a stable machine-readable identifier.
- `error.message` is a short human-readable summary.
- `error.details` contains field-level or context-specific information when available.
- `request_id` should be included in logs and bug reports.

## Common Status Codes

| Status | Meaning                           |
| -----: | --------------------------------- |
|    400 | Invalid request                   |
|    401 | Invalid authentication or API key |
|    403 | Action not permitted              |
|    404 | Resource not found                |
|    409 | Resource already exists           |
|    422 | Validation failed                 |
|    429 | Rate limit exceeded               |
|    500 | Unexpected server error           |
|    503 | Service temporarily unavailable   |

## Typical Error Categories

- Authentication failures return `401 Unauthorized`.
- Permission failures return `403 Forbidden`.
- Missing resources return `404 Not Found`.
- Validation failures return `422 Unprocessable Entity`.
- Rate limiting returns `429 Too Many Requests`.
- Unexpected server failures return `500 Internal Server Error`.

## Validation Error

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "One or more fields are invalid.",
    "details": [
      {
        "field": "limit",
        "message": "Limit must not exceed 100."
      }
    ],
    "request_id": "req_01J..."
  }
}
```

## Error Codes

```text
INVALID_REQUEST
VALIDATION_FAILED
INVALID_CREDENTIALS
INVALID_API_KEY
EXPIRED_API_KEY
REVOKED_API_KEY
RESOURCE_NOT_FOUND
RESOURCE_CONFLICT
RATE_LIMIT_EXCEEDED
INTERNAL_ERROR
SERVICE_UNAVAILABLE
```

Include the request ID when reporting an API problem.

## Client Guidance

- Treat `401` and `403` as non-retryable until credentials or permissions change.
- Retry `429` only after the suggested delay.
- Retry `503` with backoff if the request is safe to repeat.
- Do not assume every error has structured `details`.
