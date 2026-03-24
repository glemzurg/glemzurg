# Logic Target Must Not Start with Underscore (E21107)

A logic item's `target` starts with an underscore, which is reserved for internal use.

## What Went Wrong

The `target` field for `let` and `query` logic items must not begin with an underscore (`_`). Names starting with underscore are reserved.

## How to Fix

Rename the target to not start with an underscore:

```json
{
    "type": "let",
    "target": "calculated_total",
    "description": "Calculate the total",
    "specification": "SumOf(items, amount)"
}
```

## Related Errors

- **E21105**: Logic target is required
