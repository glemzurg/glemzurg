# Use Case Schema Violation (E18004)

The use case JSON file does not conform to the expected schema.

## What Went Wrong

The file contains valid JSON but violates the schema rules. Common causes:
- A required field is missing
- A field has the wrong type (e.g., number instead of string)
- An unknown field is present (no additional properties allowed)
- A string field is empty when a minimum length is required

## How to Fix

1. Run `req_check --schema use_case` to see the full JSON schema
2. Compare your file against the schema requirements
3. Fix the specific violation mentioned in the error message

## File Location

Use case files: `use_cases/{key}/use_case.json`.

## How to Read the Error Message

The error message includes the JSON path to the failing field and the specific rule violated:

| Error Pattern | Meaning | Fix |
|--------------|---------|-----|
| `missing properties: 'X'` | Required field X is absent | Add the field |
| `expected string, but got number` | Wrong value type | Use a string in double quotes |
| `length must be >= 1, but got 0` | Empty string not allowed | Provide at least one character |
| `additionalProperties 'X' not allowed` | Unknown field present | Remove the field or check spelling |

## Related

- Run `req_check --schema use_case` for the full schema
- Run `req_check --format-docs` for model format documentation
- Run `req_check --tree` for expected directory structure
