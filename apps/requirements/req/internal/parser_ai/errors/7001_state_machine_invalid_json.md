# State Machine Invalid JSON (E7001)

The state machine JSON file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your state machine file but encountered a JSON syntax error. The file contents are not valid JSON.

## File Location

State machine files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── domain.json
    ├── order.class.json
    └── order.state_machine.json    <-- This file contains invalid JSON
```

## How to Fix

Ensure your state machine file contains valid JSON. A minimal valid file looks like:

```json
{
    "states": {
        "pending": {
            "name": "Pending"
        }
    },
    "events": {
        "submit": {
            "name": "Submit"
        }
    }
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "states"
{
    "states": {
        "pending": { "name": "Pending" }
    }
    "events": {}
}

// CORRECT
{
    "states": {
        "pending": { "name": "Pending" }
    },
    "events": {}
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma
{
    "states": {
        "pending": { "name": "Pending" },
    }
}

// CORRECT
{
    "states": {
        "pending": { "name": "Pending" }
    }
}
```

### 3. Single Quotes Instead of Double Quotes

```json
// WRONG
{
    'states': { 'pending': { 'name': 'Pending' } }
}

// CORRECT
{
    "states": { "pending": { "name": "Pending" } }
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like [JSONLint](https://jsonlint.com/) can pinpoint errors
2. **Check encoding**: Ensure UTF-8 encoding without BOM
3. **Verify nested structure**: State machines have deeply nested objects

### Command Line Validation

```bash
# Validate JSON with jq
jq . order.state_machine.json

# Validate with Python
python3 -m json.tool order.state_machine.json
```

## Related Errors

- **E7002**: JSON is valid but doesn't match the schema
- **E7003**: State name is missing
- **E7004**: State name is empty
