# Contributing to SoftData

Thank you for contributing to SoftData.

SoftData accepts contributions to its source code, documentation and public datasets.

## Ways to Contribute

You can contribute by:

- Reporting incorrect records
- Updating outdated information
- Adding missing records
- Suggesting a new dataset
- Adding source attribution
- Improving documentation
- Creating validation rules
- Fixing bugs
- Adding tests
- Improving API performance

## Before Contributing Data

Every dataset contribution must be legally reusable.

Do not submit:

- Private or confidential information
- Leaked information
- Personal information about private individuals
- Data obtained through unauthorized access
- Data with unclear redistribution rights
- Copied commercial datasets
- Unsourced claims
- Generated data presented as official data

A publicly accessible webpage does not automatically mean its content can be redistributed.

## Dataset Requirements

Every new dataset must include:

- Dataset name
- Dataset identifier
- Dataset group
- Description
- Country or geographical coverage
- Original source name
- Original source URL
- Licence or usage terms
- Version
- Last-verified date
- Update frequency
- Maintainer
- Machine-readable data file
- Matching schema
- Validation results

## Dataset File Naming

Use lowercase snake case:

```text
states.json
financial_institutions.json
health_facilities.json
economic_indicators.json
```

Avoid:

```text
StatesData.json
final-data.json
new_dataset_2.json
```

## Dataset Identifiers

Dataset identifiers should be lowercase and separated with hyphens:

```text
ng-states
ng-lgas
ng-financial-institutions
ng-health-facilities
```

Identifiers must remain stable after publication.

## Data Formatting

Use UTF-8 encoding.

JSON files should:

- Use lowercase snake-case field names
- Use stable identifiers
- Avoid duplicate records
- Use `null` for genuinely unavailable values
- Avoid empty placeholder strings
- Use ISO date formats
- Use ISO country and currency codes where possible

Example:

```json
{
  "id": "NG-KW",
  "name": "Kwara",
  "country_code": "NG",
  "capital": "Ilorin",
  "geopolitical_zone": "North Central",
  "created_at": null
}
```

## Source Information

Every contribution must explain where the data came from.

Example:

```json
{
  "dataset_id": "ng-states",
  "source_name": "Official source name",
  "source_url": "https://example.gov.ng/dataset",
  "publisher": "Publishing organization",
  "accessed_at": "2026-08-26",
  "licence": "Open Government Licence",
  "notes": "Additional source information"
}
```

If multiple sources were used, list all of them.

## Making a Contribution

1. Fork the repository.
2. Create a branch.
3. Make the change.
4. Add or update tests.
5. Run dataset validation.
6. Update the relevant metadata.
7. Update the changelog when appropriate.
8. Open a pull request.

Example branch names:

```text
dataset/add-ng-states
dataset/fix-lga-spelling
feature/add-csv-download
fix/rate-limit-headers
docs/update-quick-start
```

## Validation

Run:

```bash
make validate-datasets
```

Run the tests:

```bash
make test
```

Run linting:

```bash
make lint
```

A pull request should not be merged when validation or tests fail.

## Dataset Correction Pull Requests

A correction should explain:

- What is incorrect
- What the correct value should be
- Which records are affected
- Which reliable source supports the correction
- Whether the change is breaking

Do not replace valid information without providing a source.

## Code Contributions

Code contributions should:

- Follow existing project structure
- Keep handlers thin
- Place business rules in services
- Access storage through repositories
- Avoid unnecessary dependencies
- Include appropriate tests
- Return standardized API responses
- Avoid exposing secrets or sensitive information

## Commit Messages

Use clear commit messages:

```text
feat: add geography state endpoints
fix: correct duplicate LGA records
data: update financial institution codes
docs: explain anonymous rate limits
test: add API key middleware tests
```

## Pull Request Checklist

Before submitting, confirm:

- [ ] The change has a clear purpose.
- [ ] The data source is documented.
- [ ] Redistribution is permitted.
- [ ] Dataset metadata was updated.
- [ ] Dataset version was updated when required.
- [ ] Validation passes.
- [ ] Tests pass.
- [ ] Documentation was updated.
- [ ] No secrets or credentials were committed.
- [ ] No unrelated changes are included.

## Review

Maintainers may request:

- Stronger source evidence
- Formatting corrections
- Additional tests
- Schema changes
- Licence clarification
- Smaller pull requests

A contribution may be declined if its accuracy, source or redistribution rights cannot be verified.

## Code of Conduct

All contributors must follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
