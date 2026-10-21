# Parameter Name Empty (E15002)

The parameter object has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The parser found a `name` field in a parameter object within an action, query, or event file, but its value is either an empty string (`""`) or contains only whitespace characters. The parameter name must contain at least one visible character.

## How to Fix

Provide a meaningful, non-empty name for your parameter:

```json
{
    "name": "amount"
}
```

## Invalid Examples

These values will all trigger this error:

```json
{"name": ""}              // Empty string
{"name": "   "}           // Spaces only
{"name": "\t"}            // Tab only
```

## Valid Examples

The name must contain at least one non-whitespace character:

```json
// Minimal valid parameter
{"name": "amount"}

// Parameter with data type rules
{"name": "email_address", "data_type_rules": "valid email address"}
```

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `data_type_rules` | string | No | None |

## Troubleshooting Checklist

1. **Check for invisible characters**: The name may appear present but contain only whitespace
2. **Check for copy-paste issues**: Pasting from rich text can introduce invisible characters
3. **Check the value is a string**: Ensure the name is a quoted string, not `null` or a number
4. **Check field name spelling**: The field must be exactly `"name"` (lowercase)

## Common Mistakes

```json
// WRONG: Empty string
{
    "name": "",
    "data_type_rules": "positive integer"
}

// WRONG: Whitespace only
{
    "name": "   "
}

// WRONG: Null value
{
    "name": null
}
```

## Related Errors

- **E15001**: Parameter name field is missing entirely
- **E15003**: Invalid JSON syntax in parameter object
- **E15004**: Parameter schema violation
