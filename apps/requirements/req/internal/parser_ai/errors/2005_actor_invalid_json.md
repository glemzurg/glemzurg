# Actor Invalid JSON (E2005)

The actor JSON file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your actor file but encountered a JSON syntax error. The file contents are not valid JSON, which means the parser cannot extract any values from it.

## File Location

Actor files are located in the `actors/` directory at the model root:

```
your_model/
├── model.json
├── actors/
│   └── customer.actor.json    <-- This file contains invalid JSON syntax
├── domains/
└── ...
```

## How to Fix

Ensure your actor file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "Customer",
    "type": "human"
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "Customer"
    "type": "human"
}

// CORRECT: Comma separates properties
{
    "name": "Customer",
    "type": "human"
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma after last property
{
    "name": "Customer",
    "type": "human",
}

// CORRECT: No comma after last property
{
    "name": "Customer",
    "type": "human"
}
```

### 3. Single Quotes Instead of Double Quotes

```json
// WRONG: Single quotes are not valid JSON
{
    'name': 'Customer',
    'type': 'human'
}

// CORRECT: Must use double quotes
{
    "name": "Customer",
    "type": "human"
}
```

### 4. Unquoted Property Names

```json
// WRONG: Property names must be quoted
{
    name: "Customer",
    type: "human"
}

// CORRECT: Property names in double quotes
{
    "name": "Customer",
    "type": "human"
}
```

### 5. Missing or Mismatched Braces

```json
// WRONG: Missing closing brace
{
    "name": "Customer",
    "type": "human"

// CORRECT: Matching braces
{
    "name": "Customer",
    "type": "human"
}
```

### 6. Comments in JSON

```json
// WRONG: JSON does not support comments
{
    "name": "Customer",  // This is not allowed
    "type": "human"
}

// CORRECT: No comments
{
    "name": "Customer",
    "type": "human"
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like JSONLint can pinpoint exact errors
2. **Check encoding**: Ensure the file is UTF-8 encoded without BOM
3. **Look for invisible characters**: Copy-paste from documents can introduce hidden characters
4. **Verify file isn't empty**: An empty file is not valid JSON

### Command Line Validation

```bash
# Validate JSON with jq (will show error location)
jq . actors/customer.actor.json

# Check for non-printable characters
cat -A actors/customer.actor.json

# Validate with Python
python3 -m json.tool actors/customer.actor.json
```

## Valid Actor File Template

Start with this template and customize:

```json
{
    "name": "Actor Name",
    "type": "human",
    "details": "Optional description of the actor"
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

- **E2001**: Actor name field is missing (JSON is valid but incomplete)
- **E2003**: Actor type field is missing (JSON is valid but incomplete)
- **E2006**: JSON is valid but doesn't match the schema
