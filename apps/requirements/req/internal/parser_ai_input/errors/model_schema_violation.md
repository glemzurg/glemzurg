# Model Schema Violation (E1004)

The model.json file does not conform to the expected schema.

## Required Fields

- **name** (string): The name of the model. Must not be empty.

## Optional Fields

- **details** (string): A description of the model.

## Rules

- No additional properties are allowed beyond `name` and `details`
- The `name` field must be a non-empty string

## Example of Valid model.json

```json
{
    "name": "Web Store",
    "details": "An online marketplace for selling products"
}
```

## Common Mistakes

1. Adding extra fields not defined in the schema
2. Using wrong types (e.g., number instead of string for name)
3. Missing the required `name` field
