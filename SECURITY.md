# Security Policy

SoftData takes the security of its API, contributors and users seriously.

## Supported Versions

| Version             | Supported |
| ------------------- | --------: |
| Latest `v1` release |       Yes |
| Development builds  |   Limited |
| Deprecated releases |        No |

## Reporting a Vulnerability

Do not report security vulnerabilities through public GitHub issues.

Send a private report to:

```text
alaoabdulquayyumm@gmail.com
```

Use the subject:

```text
SoftData Security Report
```

Include:

- A description of the vulnerability
- The affected endpoint or component
- Steps required to reproduce it
- Its possible impact
- Suggested mitigation, if available
- Your preferred contact information

Do not include real user passwords, complete API keys or unnecessary personal data.

## Responsible Disclosure

When researching a possible vulnerability:

- Do not access other users’ accounts.
- Do not download data belonging to other accounts.
- Do not expose credentials publicly.
- Do not intentionally disrupt the hosted API.
- Do not perform denial-of-service testing.
- Do not repeatedly exceed published rate limits.
- Stop testing if private information becomes accessible.

Allow maintainers a reasonable period to investigate and release a fix before public disclosure.

## In Scope

Examples of security issues include:

- Authentication bypass
- API-key validation bypass
- Account takeover
- Exposure of raw API keys
- SQL injection
- Unauthorized access to usage records
- Rate-limit bypass with material impact
- Sensitive information exposure
- Remote code execution
- Dependency vulnerabilities affecting the deployed API

## Out of Scope

The following are generally not vulnerabilities:

- Public access to intentionally public datasets
- Missing authentication on documented public endpoints
- Published anonymous rate limits
- Username enumeration without additional impact
- Self-inflicted rate limiting
- Missing security headers without demonstrated impact
- Vulnerabilities in unsupported versions
- Social-engineering attacks
- Denial-of-service testing

## API-Key Handling

SoftData:

- Generates API keys using cryptographically secure randomness.
- Shows the complete key only once.
- Stores only a hash and short prefix.
- Never returns existing complete keys.
- Supports revocation and rotation.
- Never logs complete API keys.
- Uses the documented `sd_live_` public prefix.

Users should:

- Store keys in environment variables.
- Never commit keys to Git.
- Avoid embedding private keys in public frontend code.
- Rotate keys that may have been exposed.
- Use separate keys for separate applications.

## Password Handling

Passwords must:

- Be hashed using Argon2id.
- Never be stored or logged as plain text.
- Never be returned by an API response.
- Be checked against reasonable minimum requirements.

## Token And Privacy Handling

- Access tokens are HS256 JWTs signed with `AUTH_TOKEN_SECRET`.
- Refresh tokens are opaque random values stored only as SHA-256 hashes.
- Anonymous identifiers use a separate HMAC-SHA256 secret and rotate daily using a UTC date bucket.

## Privacy

Usage tracking must not store:

- Complete API keys
- Passwords
- Authorization headers
- Response bodies
- Unnecessary request bodies
- Raw IP addresses beyond operational necessity

Anonymous identifiers should be rotated and should not be treated as permanent identities.

## Security Updates

Security-related changes will be documented without publishing details that would place users at unnecessary risk.
