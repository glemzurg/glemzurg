# Model Invalid JSON (E1003)

The `model.json` file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your `model.json` file but encountered a JSON syntax error. The file contents are not valid JSON, which means the parser cannot extract any values from it.

## File Location

The `model.json` file must exist at the **root** of your model directory:

```
your_model/
├── model.json          <-- This file contains invalid JSON syntax
├── actors/
├── domains/
├── associations/
└── generalizations/
```

## How to Fix

Ensure your `model.json` file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "Your Model Name"
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "My Model"
    "details": "Description"
}

// CORRECT: Comma separates properties
{
    "name": "My Model",
    "details": "Description"
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma after last property
{
    "name": "My Model",
    "details": "Description",
}

// CORRECT: No comma after last property
{
    "name": "My Model",
    "details": "Description"
}
```

### 3. Single Quotes Instead of Double Quotes

```json
// WRONG: Single quotes are not valid JSON
{
    'name': 'My Model'
}

// CORRECT: Must use double quotes
{
    "name": "My Model"
}
```

### 4. Unquoted Property Names

```json
// WRONG: Property names must be quoted
{
    name: "My Model"
}

// CORRECT: Property names in double quotes
{
    "name": "My Model"
}
```

### 5. Unquoted String Values

```json
// WRONG: String values must be quoted
{
    "name": My Model
}

// CORRECT: String values in double quotes
{
    "name": "My Model"
}
```

### 6. Missing or Mismatched Braces

```json
// WRONG: Missing closing brace
{
    "name": "My Model"

// WRONG: Extra closing brace
{
    "name": "My Model"
}}

// CORRECT: Matching braces
{
    "name": "My Model"
}
```

### 7. Comments in JSON

```json
// WRONG: JSON does not support comments
{
    "name": "My Model"  // This is not allowed
}

// WRONG: Multi-line comments also invalid
{
    /* comment */
    "name": "My Model"
}

// CORRECT: No comments
{
    "name": "My Model"
}
```

### 8. JavaScript/Python Syntax Confusion

```json
// WRONG: Python True/False/None
{
    "name": "My Model",
    "active": True,
    "deprecated": None
}

// CORRECT: JSON uses true/false/null (lowercase)
{
    "name": "My Model",
    "active": true,
    "deprecated": null
}
```

### 9. Unescaped Special Characters in Strings

```json
// WRONG: Unescaped newline in string
{
    "name": "My
Model"
}

// WRONG: Unescaped backslash
{
    "path": "C:\Users\name"
}

// CORRECT: Escape special characters
{
    "name": "My Model",
    "path": "C:\\Users\\name"
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like [JSONLint](https://jsonlint.com/) can pinpoint exact errors
2. **Check encoding**: Ensure the file is UTF-8 encoded without BOM
3. **Look for invisible characters**: Copy-paste from documents can introduce hidden characters
4. **Verify file isn't empty**: An empty file is not valid JSON
5. **Check for BOM**: Some editors add a byte-order mark that breaks JSON parsing

### Command Line Validation

```bash
# Validate JSON with jq (will show error location)
jq . model.json

# Check for non-printable characters
cat -A model.json

# Validate with Python
python3 -m json.tool model.json

# Check file encoding
file model.json
```

## Valid model.json Template

Start with this template and customize:

```json
{
    "name": "Your Model Name",
    "details": "Optional description of what this model represents"
}
```

## JSON Quick Reference

| Element | Syntax | Example |
|---------|--------|---------|
| Object | `{ }` | `{"key": "value"}` |
| Array | `[ ]` | `["a", "b", "c"]` |
| String | `" "` | `"hello"` |
| Number | digits | `42`, `3.14`, `-5` |
| Boolean | `true`/`false` | `true` |
| Null | `null` | `null` |

## Related Errors

- **E1001**: Model name field is missing (JSON is valid but incomplete)
- **E1002**: Model name is empty (JSON is valid but value is wrong)
- **E1004**: JSON is valid but doesn't match the schema
