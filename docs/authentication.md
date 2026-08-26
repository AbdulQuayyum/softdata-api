# Authentication

SoftData supports two forms of access:

1. Anonymous dataset access
2. Optional authenticated developer access

## Anonymous Access

Dataset endpoints can be called without authentication.

```bash
curl https://softdata-api.vercel.app/v1/geography/states
```

Anonymous requests use IP-based rate limits.

## Developer Accounts

Developer accounts are only needed for:

- Creating API keys
- Receiving higher limits
- Viewing usage analytics
- Managing application keys

## Register

```http
POST /v1/auth/register
Content-Type: application/json
```

```json
{
  "username": "abdulquayyum",
  "password": "strong-password",
  "email": "optional@example.com"
}
```

The email field is optional. Password recovery is only available to accounts with an email address.

## Login

```http
POST /v1/auth/login
Content-Type: application/json
```

```json
{
  "username": "abdulquayyum",
  "password": "strong-password"
}
```

## Account Authentication

Account routes use a bearer access token:

```http
Authorization: Bearer your_access_token
```

Bearer tokens are not used for dataset access. Dataset access uses no key or an optional API key.

## Refreshing a Session

```http
POST /v1/auth/refresh
Content-Type: application/json
```

```json
{
  "refresh_token": "your_refresh_token"
}
```

## Logging Out

```http
POST /v1/auth/logout
Authorization: Bearer your_access_token
Content-Type: application/json
```

```json
{
  "refresh_token": "your_refresh_token"
}
```

## Security

- Store access and refresh tokens securely.
- Never send credentials over plain HTTP.
- Never commit credentials to source control.
- Rotate credentials that may have been exposed.
