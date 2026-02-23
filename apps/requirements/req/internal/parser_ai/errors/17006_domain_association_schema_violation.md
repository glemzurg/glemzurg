# Domain Association Schema Violation (E17006)

The domain association JSON is valid JSON but does not conform to the expected schema.

## What Went Wrong

The JSON was parsed successfully, but its structure does not match what is expected for a domain association. This typically means:
- A required field is missing (`problem_domain_key` or `solution_domain_key`)
- A field has the wrong type (e.g., a key is a number instead of a string)
- An unknown/extra field is present
- A required string field is empty

## Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `problem_domain_key` | string | Yes | Key of the problem domain |
| `solution_domain_key` | string | Yes | Key of the solution domain |
| `uml_comment` | string | No | Optional UML comment |

## Correct Format

```json
{
    "problem_domain_key": "billing",
    "solution_domain_key": "payment_processing"
}
```

## Related Errors

- **E17005**: The JSON itself is malformed (syntax error)
- **E17001**: The `problem_domain_key` field is specifically missing
- **E17003**: The `solution_domain_key` field is specifically missing
