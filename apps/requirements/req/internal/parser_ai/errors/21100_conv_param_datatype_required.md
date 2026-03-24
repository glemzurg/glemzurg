# Parameter Data Type Rules Required (E21100)

A parameter's `data_type_rules` field is empty or missing.

## What Went Wrong

Every parameter on an event, action, or query must have a non-empty `data_type_rules` value that describes its data type.

## How to Fix

Add a `data_type_rules` value to the parameter. For the complete data type syntax, see **E5011**.

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

- **E5011**: Attribute data type unparseable (full syntax reference)
- **E15005**: Parameter data type unparseable
- **E21101**: Parameter name is required
