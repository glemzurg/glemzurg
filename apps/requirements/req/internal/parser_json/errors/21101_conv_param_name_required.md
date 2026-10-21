# Parameter Name Required (E21101)

A parameter's `name` field is empty or missing.

## What Went Wrong

Every parameter on an event, action, or query must have a non-empty `name` value.

## How to Fix

Add a `name` value to the parameter:

```json
{
    "parameters": {
        "amount": {
            "name": "Amount",
            "data_type_rules": "[0 .. unconstrained] at 0.01 dollars"
        }
    }
}
```

## Related Errors

- **E21100**: Parameter data type rules required
