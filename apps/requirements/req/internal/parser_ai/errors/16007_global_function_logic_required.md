# Global Function Logic Required (E16007)

A global function has a `logic` object whose `description` is missing or contains only whitespace.

## What Went Wrong

Every global function must have a `logic` object with a non-empty `description` field that explains what the function computes or represents. The parser found a logic description that is empty or whitespace-only.

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

## How to Fix

Provide a meaningful description in the logic object:

```json
// Wrong:
{ "logic": { "description": "   " } }

// Correct:
{ "logic": { "description": "Returns the maximum of two values" } }
```

## Writing Good Descriptions

- Explain **what** the function computes: "Returns the maximum of two values"
- Describe **constraints**: "Clamps a value between min and max inclusive"
- State the **purpose**: "Defines the set of valid order statuses"

## Related Errors

- **E16001**: Global function name is missing
- **E16005**: Global function name does not start with underscore
