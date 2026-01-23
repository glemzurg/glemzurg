# Model Invalid JSON (E1003)

The model.json file contains invalid JSON syntax.

## Common Causes

1. **Missing commas** between properties
2. **Trailing commas** after the last property
3. **Unquoted strings** - all string values must be in double quotes
4. **Single quotes** - JSON requires double quotes, not single quotes
5. **Missing closing braces** or brackets

## Example of Valid JSON

```json
{
    "name": "My Model",
    "details": "Optional description of the model"
}
```

## How to Fix

1. Use a JSON validator to identify syntax errors
2. Ensure all strings are wrapped in double quotes
3. Remove any trailing commas
4. Verify all braces and brackets are properly matched
