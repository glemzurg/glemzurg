# Action Invalid JSON (E8003)

The action JSON file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your action file but encountered a JSON syntax error. The file contents are not valid JSON.

## File Location

Action files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.actions.json    <-- This file contains invalid JSON
```

## How to Fix

Ensure your action file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "Send Confirmation Email"
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "Send Email"
    "details": "Sends an email"
}

// CORRECT
{
    "name": "Send Email",
    "details": "Sends an email"
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma
{
    "name": "Send Email",
    "details": "Sends an email",
}

// CORRECT
{
    "name": "Send Email",
    "details": "Sends an email"
}
```

### 3. Missing Commas in Arrays

```json
// WRONG: Missing comma in array
{
    "name": "Process Order",
    "requires": [
        "Order exists"
        "Order not empty"
    ]
}

// CORRECT
{
    "name": "Process Order",
    "requires": [
        "Order exists",
        "Order not empty"
    ]
}
```

### 4. Single Quotes Instead of Double Quotes

```json
// WRONG
{
    'name': 'Send Email'
}

// CORRECT
{
    "name": "Send Email"
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like [JSONLint](https://jsonlint.com/) can pinpoint errors
2. **Check encoding**: Ensure UTF-8 encoding without BOM
3. **Check array syntax**: The `requires` and `guarantees` fields are arrays

### Command Line Validation

```bash
# Validate JSON with jq
jq . order.actions.json

# Validate with Python
python3 -m json.tool order.actions.json
```

## Related Errors

- **E8001**: Name field is missing (JSON is valid but incomplete)
- **E8002**: Name is empty (JSON is valid but value is wrong)
- **E8004**: JSON is valid but doesn't match the schema
