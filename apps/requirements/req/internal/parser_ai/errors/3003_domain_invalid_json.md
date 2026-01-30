# Domain Invalid JSON (E3003)

The `domain.json` file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your `domain.json` file but encountered a JSON syntax error. The file contents are not valid JSON, which means the parser cannot extract any values from it.

## File Location

Domain files are located in directories named after the domain:

```
your_model/
├── model.json
└── order_management/
    ├── domain.json             <-- This file contains invalid JSON syntax
    └── ... (classes, etc.)
```

## How to Fix

Ensure your `domain.json` file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "Order Management"
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "Order Management"
    "details": "Description"
}

// CORRECT: Comma separates properties
{
    "name": "Order Management",
    "details": "Description"
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma after last property
{
    "name": "Order Management",
    "details": "Description",
}

// CORRECT: No comma after last property
{
    "name": "Order Management",
    "details": "Description"
}
```

### 3. Single Quotes Instead of Double Quotes

```json
// WRONG: Single quotes are not valid JSON
{
    'name': 'Order Management'
}

// CORRECT: Must use double quotes
{
    "name": "Order Management"
}
```

### 4. Unquoted Property Names

```json
// WRONG: Property names must be quoted
{
    name: "Order Management"
}

// CORRECT: Property names in double quotes
{
    "name": "Order Management"
}
```

### 5. Boolean Values Must Be Lowercase

```json
// WRONG: True/False with capital letters
{
    "name": "Order Management",
    "realized": True
}

// CORRECT: Lowercase true/false
{
    "name": "Order Management",
    "realized": true
}
```

### 6. Missing or Mismatched Braces

```json
// WRONG: Missing closing brace
{
    "name": "Order Management"

// CORRECT: Matching braces
{
    "name": "Order Management"
}
```

### 7. Comments in JSON

```json
// WRONG: JSON does not support comments
{
    "name": "Order Management"  // This is not allowed
}

// CORRECT: No comments
{
    "name": "Order Management"
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
jq . domain.json

# Check for non-printable characters
cat -A domain.json

# Validate with Python
python3 -m json.tool domain.json
```

## Valid domain.json Template

Start with this template and customize:

```json
{
    "name": "Your Domain Name",
    "details": "Optional description of what this domain covers"
}
```

## JSON Quick Reference

| Element | Syntax | Example |
|---------|--------|---------|
| Object | `{ }` | `{"key": "value"}` |
| String | `" "` | `"hello"` |
| Boolean | `true`/`false` | `true` |
| Null | `null` | `null` |

## Related Errors

- **E3001**: Domain name field is missing (JSON is valid but incomplete)
- **E3002**: Domain name is empty (JSON is valid but value is wrong)
- **E3004**: JSON is valid but doesn't match the schema
