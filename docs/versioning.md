# Versioning

SoftData versions its API and datasets separately.

## API Versions

API versions appear in the URL:

```text
/v1/geography/states
```

API versioning rules:

- Major API versions are part of the path.
- Breaking changes require a new major version.
- Non-breaking additions should remain in the current version.
- Older versions may remain available for a transition period.

The initial public release uses `v1`.

## Dataset Versions

Dataset versions track the content and schema of a specific dataset.

Examples:

- `1.0.0`
- `1.1.0`
- `2.0.0`

Dataset version changes should reflect:

- added or removed fields
- corrected records
- schema changes
- source changes that affect the output

## When To Bump Versions

Use a major bump when a change breaks existing clients or response contracts.

Use a minor bump when you add backwards-compatible fields or records.

Use a patch bump when you correct data without changing the contract.

## Deprecation

When a dataset or API route is being replaced:

- keep the old version available for a reasonable period
- document the replacement clearly
- mark the older version as deprecated in metadata and docs
- avoid removing a version without a migration path

## Compatibility

SoftData should aim to preserve:

- field names
- response shapes
- status codes
- sorting and filtering behavior

If compatibility must change, document it clearly in the changelog and the relevant docs page.

## Version Metadata

Version metadata should record:

- the version number
- the release date
- the source or schema change that caused the update
- whether the version is active, deprecated, or archived

## Practical Guidance

- Prefer additive changes over breaking ones.
- Keep version numbers human-readable.
- Treat dataset versioning as independent from API versioning.
- Update docs whenever a versioned response changes.
