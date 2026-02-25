# Parameter Name Required (E15001)

The parameter object is missing the required `name` field.

## What Went Wrong

The parser found a parameter object within an action, query, or event JSON file, but it does not contain a `name` property. Every parameter must have a name that identifies the input.

## How to Fix

Add a `name` field to your parameter object:

```json
{
    "name": "amount"
}
```

Or with the optional `data_type_rules` field:

```json
{
    "name": "email_address",
    "data_type_rules": "valid email address"
}
```

## Complete Schema

The parameter object accepts these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Human-readable name for the parameter |
| `data_type_rules` | string | No | Describes the expected data type or validation rules |

## Troubleshooting Checklist

1. **Check the parameter object exists**: Ensure the parameter is inside the `parameters` array of the parent action, query, or event file
2. **Check JSON syntax**: The parameter object must be valid JSON
3. **Check field name spelling**: The field must be exactly `"name"` (lowercase)
4. **Check the value exists**: Ensure the name has a value, not just the key

## Common Mistakes

```json
// WRONG: Missing name entirely
{
    "data_type_rules": "positive integer"
}

// WRONG: Typo in field name
{
    "Name": "amount"
}

// WRONG: Empty parameter object
{}
```

## Valid Examples

```json
// Minimal valid parameter
{"name": "amount"}

// Parameter with data type rules
{"name": "email_address", "data_type_rules": "valid email address"}
```

## Related Errors

- **E15002**: Parameter name is present but empty or whitespace
- **E15003**: Invalid JSON syntax in parameter object
- **E15004**: Parameter schema violation
