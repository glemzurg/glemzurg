# Global Function Invalid JSON (E16003)

The global function file contains invalid JSON that cannot be parsed.

## What Went Wrong

The file content is not valid JSON. This could be due to:
- Missing or extra commas
- Unquoted property names
- Single quotes instead of double quotes
- Missing closing braces or brackets
- Trailing commas after the last property
- Comments (JSON does not support comments)

## Common Mistakes

### Missing comma
```json
{
    "name": "_Max"
    "logic": { "description": "Returns maximum" }
}
```
Fix: Add comma after `"_Max"`.

### Trailing comma
```json
{
    "name": "_Max",
    "logic": { "description": "Returns maximum" },
}
```
Fix: Remove the comma after the last property.

### Single quotes
```json
{
    'name': '_Max',
    'logic': { 'description': 'Returns maximum' }
}
```
Fix: Use double quotes for all strings and property names.

## Valid Global Function Template

```json
{
    "name": "_YourFunctionName",
    "parameters": ["param1", "param2"],
    "logic": {
        "description": "What this function does",
        "notation": "tla_plus",
        "specification": "formal expression here"
    }
}
```

## Related Errors

- **E16004**: JSON is valid but does not match the global function schema
