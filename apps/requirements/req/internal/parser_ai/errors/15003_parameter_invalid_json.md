# Parameter Invalid JSON (E15003)

A parameter object within an action, query, or event JSON file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read a parameter object but encountered a JSON syntax error. The parameter contents are not valid JSON.

## How to Fix

Ensure your parameter object contains valid JSON. A minimal valid parameter looks like:

```json
{
    "name": "amount"
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "amount"
    "data_type_rules": "positive integer"
}

// CORRECT: Comma separates properties
{
    "name": "amount",
    "data_type_rules": "positive integer"
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma after last property
{
    "name": "amount",
    "data_type_rules": "positive integer",
}

// CORRECT: No trailing commas
{
    "name": "amount",
    "data_type_rules": "positive integer"
}
```

### 3. Single Quotes Instead of Double Quotes

```json
// WRONG: Single quotes
{
    'name': 'amount'
}

// CORRECT: Double quotes
{
    "name": "amount"
}
```

### 4. Unquoted String Values

```json
// WRONG: Unquoted value
{
    "name": amount
}

// CORRECT: Quoted value
{
    "name": "amount"
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like [JSONLint](https://jsonlint.com/) can pinpoint errors
2. **Check encoding**: Ensure UTF-8 encoding without BOM
3. **Look for invisible characters**: Copy-paste can introduce hidden characters
4. **Check the parent file**: The parameter is embedded inside an action, query, or event file -- verify the entire file is valid JSON

### Command Line Validation

```bash
# Validate the parent JSON file with jq
jq . your_action_file.json

# Validate with Python
python3 -m json.tool your_action_file.json
```

## Complete Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Human-readable name for the parameter |
| `data_type_rules` | string | No | Describes the expected data type or validation rules |

## Valid Examples

```json
// Minimal valid parameter
{"name": "amount"}

// Parameter with data type rules
{"name": "email_address", "data_type_rules": "valid email address"}
```

## Related Errors

- **E15001**: Name field is missing (JSON is valid but incomplete)
- **E15002**: Name is empty (JSON is valid but value is wrong)
- **E15004**: JSON is valid but doesn't match the schema
