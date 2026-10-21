# Global Function Parameter Empty (E16006)

A global function has a parameter that is empty or contains only whitespace.

## What Went Wrong

The `parameters` array contains an entry that is an empty string or consists only of whitespace characters. Each parameter must have a meaningful name.

## Correct Format

```json
{
    "name": "_Max",
    "parameters": ["x", "y"],
    "logic": {
        "description": "Returns the maximum of two values"
    }
}
```

## How to Fix

Ensure every parameter in the array is a non-empty string:

```json
// Wrong:
{ "parameters": ["x", "", "z"] }

// Correct:
{ "parameters": ["x", "y", "z"] }
```

## Related Errors

- **E16001**: Global function name is missing
- **E16007**: Global function logic description is missing or empty
