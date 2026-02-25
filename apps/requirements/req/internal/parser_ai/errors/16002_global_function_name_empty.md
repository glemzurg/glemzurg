# Global Function Name Empty (E16002)

A global function has a `name` field that contains only whitespace characters.

## What Went Wrong

The parser found a global function with a `name` that is present but consists entirely of spaces, tabs, or other whitespace. The name must contain visible characters and start with an underscore.

## Correct Format

```json
{
    "name": "_Max",
    "logic": {
        "description": "Returns the maximum of two values"
    }
}
```

## How to Fix

Replace the whitespace-only name with a valid function name starting with underscore:

```json
{
    "name": "_YourFunctionName",
    "logic": {
        "description": "What this function does"
    }
}
```

## Choosing a Good Global Function Name

- Start with underscore: `_Max`, not `Max`
- Use PascalCase: `_SetOfValues`, not `_setofvalues`
- Be descriptive: `_ClampValue` rather than `_CV`
- Reflect the computation: `_SumOfItems`, `_IsValidState`

## Related Errors

- **E16001**: Global function name field is missing entirely
- **E16005**: Global function name does not start with underscore
