# Subdomain Invalid JSON (E4003)

The `subdomain.json` file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your `subdomain.json` file but encountered a JSON syntax error. The file contents are not valid JSON, which means the parser cannot extract any values from it.

## File Location

Subdomain files are located within domain directories:

```
your_model/
├── model.json
└── order_management/
    ├── domain.json
    └── fulfillment/
        ├── subdomain.json            <-- This file contains invalid JSON syntax
        └── ... (classes, etc.)
```

## How to Fix

Ensure your `subdomain.json` file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "Order Fulfillment"
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "Order Fulfillment"
    "details": "Description"
}

// CORRECT: Comma separates properties
{
    "name": "Order Fulfillment",
    "details": "Description"
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma after last property
{
    "name": "Order Fulfillment",
    "details": "Description",
}

// CORRECT: No comma after last property
{
    "name": "Order Fulfillment",
    "details": "Description"
}
```

### 3. Single Quotes Instead of Double Quotes

```json
// WRONG: Single quotes are not valid JSON
{
    'name': 'Order Fulfillment'
}

// CORRECT: Must use double quotes
{
    "name": "Order Fulfillment"
}
```

### 4. Unquoted Property Names

```json
// WRONG: Property names must be quoted
{
    name: "Order Fulfillment"
}

// CORRECT: Property names in double quotes
{
    "name": "Order Fulfillment"
}
```

### 5. Missing or Mismatched Braces

```json
// WRONG: Missing closing brace
{
    "name": "Order Fulfillment"

// CORRECT: Matching braces
{
    "name": "Order Fulfillment"
}
```

### 6. Comments in JSON

```json
// WRONG: JSON does not support comments
{
    "name": "Order Fulfillment"  // This is not allowed
}

// CORRECT: No comments
{
    "name": "Order Fulfillment"
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like [JSONLint](https://jsonlint.com/) can pinpoint exact errors
2. **Check encoding**: Ensure the file is UTF-8 encoded without BOM
3. **Look for invisible characters**: Copy-paste from documents can introduce hidden characters
4. **Verify file isn't empty**: An empty file is not valid JSON

### Command Line Validation

```bash
# Validate JSON with jq (will show error location)
jq . subdomain.json

# Check for non-printable characters
cat -A subdomain.json

# Validate with Python
python3 -m json.tool subdomain.json
```

## Valid subdomain.json Template

Start with this template and customize:

```json
{
    "name": "Your Subdomain Name",
    "details": "Optional description of what this subdomain covers"
}
```

## JSON Quick Reference

| Element | Syntax | Example |
|---------|--------|---------|
| Object | `{ }` | `{"key": "value"}` |
| String | `" "` | `"hello"` |
| Null | `null` | `null` |

## Related Errors

- **E4001**: Subdomain name field is missing (JSON is valid but incomplete)
- **E4002**: Subdomain name is empty (JSON is valid but value is wrong)
- **E4004**: JSON is valid but doesn't match the schema
