# Query Invalid JSON (E9003)

The query JSON file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your query file but encountered a JSON syntax error. The file contents are not valid JSON.

## File Location

Query files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.queries.json    <-- This file contains invalid JSON
```

## How to Fix

Ensure your query file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "Get Order Total"
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "Get Order Total"
    "details": "Returns the total"
}

// CORRECT
{
    "name": "Get Order Total",
    "details": "Returns the total"
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma
{
    "name": "Get Order Total",
    "details": "Returns the total",
}

// CORRECT
{
    "name": "Get Order Total",
    "details": "Returns the total"
}
```

### 3. Missing Commas in Arrays

```json
// WRONG: Missing comma in array
{
    "name": "Find Orders",
    "requires": [
        "User exists"
        "User has access"
    ]
}

// CORRECT
{
    "name": "Find Orders",
    "requires": [
        "User exists",
        "User has access"
    ]
}
```

### 4. Single Quotes Instead of Double Quotes

```json
// WRONG
{
    'name': 'Get Order Total'
}

// CORRECT
{
    "name": "Get Order Total"
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like [JSONLint](https://jsonlint.com/) can pinpoint errors
2. **Check encoding**: Ensure UTF-8 encoding without BOM
3. **Check array syntax**: The `requires` and `guarantees` fields are arrays

### Command Line Validation

```bash
# Validate JSON with jq
jq . order.queries.json

# Validate with Python
python3 -m json.tool order.queries.json
```

## Related Errors

- **E9001**: Name field is missing (JSON is valid but incomplete)
- **E9002**: Name is empty (JSON is valid but value is wrong)
- **E9004**: JSON is valid but doesn't match the schema
