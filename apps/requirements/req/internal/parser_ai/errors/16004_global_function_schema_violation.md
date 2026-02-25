# Global Function Schema Violation (E16004)

The global function JSON is valid JSON but does not conform to the expected schema.

## What Went Wrong

The JSON was parsed successfully, but its structure does not match what is expected for a global function. This typically means:
- A required field is missing (`name` or `logic`)
- A field has the wrong type (e.g., `name` is a number instead of a string)
- An unknown/extra field is present
- The `logic` object is missing its required `description` field
- The `parameters` array contains a non-string or empty string

## Required Fields

Global function files must contain:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Function name starting with underscore |
| `logic` | object | Yes | Logic specification with description |
| `parameters` | array of strings | No | Named parameters for the function |

The `logic` object must contain:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | Yes | Human-readable description |
| `notation` | string | No | Formal notation system |
| `specification` | string | No | Formal specification |

## Correct Format

```json
{
    "name": "_Max",
    "parameters": ["x", "y"],
    "logic": {
        "description": "Returns the maximum of two values",
        "notation": "tla_plus",
        "specification": "IF x > y THEN x ELSE y"
    }
}
```

## Related Errors

- **E16003**: The JSON itself is malformed (syntax error)
- **E16001**: The `name` field is specifically missing
