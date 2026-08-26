# API Keys

API keys are optional. They provide higher rate limits and usage analytics.

## Creating a Key

```http
POST /v1/account/api-keys
Authorization: Bearer your_access_token
Content-Type: application/json
```

```json
{
  "name": "Portfolio Application"
}
```

The complete API key is returned only once.

SoftData API keys use the documented `sd_live_` prefix and store only a hash plus safe display metadata.

## Using a Key

```http
X-API-Key: sd_live_your_api_key
```

Example:

```bash
curl \
  -H "X-API-Key: sd_live_your_api_key" \
  https://softdata-api.vercel.app/v1/geography/states
```

## Key Limits

An account can create up to three active API keys in the initial release.

## Listing Keys

```http
GET /v1/account/api-keys
Authorization: Bearer your_access_token
```

The response contains key metadata and prefixes, not complete keys.

## Revoking a Key

```http
DELETE /v1/account/api-keys/{key_id}
Authorization: Bearer your_access_token
```

## Rotating a Key

```http
POST /v1/account/api-keys/{key_id}/rotate
Authorization: Bearer your_access_token
```

Rotation revokes the previous key and returns a replacement.

## Invalid Keys

If an `X-API-Key` header is present but invalid, SoftData returns `401 Unauthorized`.

It does not silently process the request as anonymous.

## Key Safety

- Do not commit API keys to Git.
- Use environment variables.
- Use separate keys for separate applications.
- Revoke exposed keys immediately.
- Avoid placing private keys directly in public frontend code.
