# Parameter Schema Violation (E15004)

A parameter object within an action, query, or event JSON file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read a parameter object as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing (`name`)
- A field has the wrong type
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string)

## How to Fix

Ensure your parameter object matches the expected schema. A minimal valid parameter looks like:

```json
{
    "name": "amount"
}
```

## Schema Requirements

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | Human-readable name for the parameter |

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `data_type_rules` | string | None | Describes the expected data type or validation rules |

## Common Schema Violations

### 1. Missing Required Fields

```json
// WRONG: Missing 'name'
{
    "data_type_rules": "positive integer"
}

// CORRECT: Name is present
{
    "name": "amount",
    "data_type_rules": "positive integer"
}
```

### 2. Empty Values

```json
// WRONG: Empty name
{
    "name": ""
}

// WRONG: Whitespace-only name
{
    "name": "   "
}

// CORRECT: Non-empty name
{
    "name": "amount"
}
```

### 3. Wrong Types

```json
// WRONG: name is a number, not a string
{
    "name": 42
}

// WRONG: data_type_rules is an array, not a string
{
    "name": "amount",
    "data_type_rules": ["positive", "integer"]
}

// CORRECT: Proper types
{
    "name": "amount",
    "data_type_rules": "positive integer"
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'type' is not in the schema
{
    "name": "amount",
    "type": "integer"
}

// WRONG: 'required' is not in the schema
{
    "name": "amount",
    "required": true
}

// CORRECT: Only allowed fields
{
    "name": "amount",
    "data_type_rules": "positive integer"
}
```

## Troubleshooting Checklist

1. **Check required fields**: The `name` field must be present and non-empty
2. **Check field types**: `name` and `data_type_rules` must both be strings
3. **Check for extra fields**: Only `name` and `data_type_rules` are allowed
4. **Check the parent file**: The parameter is embedded inside an action, query, or event file -- verify the parent structure is correct

## Common Mistakes

```json
// WRONG: Extra field "description"
{
    "name": "amount",
    "description": "The payment amount"
}

// WRONG: Using "type" instead of "data_type_rules"
{
    "name": "amount",
    "type": "integer"
}

// WRONG: Nested object for data_type_rules
{
    "name": "amount",
    "data_type_rules": {"type": "integer", "min": 0}
}
```

## Valid Examples

```json
// Minimal valid parameter
{"name": "amount"}

// Parameter with data type rules
{"name": "email_address", "data_type_rules": "valid email address"}
```

## Related Errors

- **E15001**: Name field is missing
- **E15002**: Name field is empty
- **E15003**: JSON syntax is invalid
