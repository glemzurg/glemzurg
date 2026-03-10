# Duplicate Let Target (E21103)

Two or more `let` logic items in the same list use the same `target` name.

## What Went Wrong

Within a single logic list (e.g., invariants, requires, guarantees), each `let` item must have a unique `target`. The same target name was used more than once.

## How to Fix

Rename one of the duplicate `let` targets to be unique within the list:

```json
{
    "requires": [
        {
            "type": "let",
            "target": "total",
            "description": "Calculate total",
            "specification": "SumOf(items, amount)"
        },
        {
            "type": "let",
            "target": "tax",
            "description": "Calculate tax",
            "specification": "total * tax_rate"
        }
    ]
}
```

## Related Errors

- **E21102**: Logic type invalid for context
- **E21104**: Duplicate guarantee target
