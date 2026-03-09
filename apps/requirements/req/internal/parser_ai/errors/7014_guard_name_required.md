# Guard Name Required (E7014)

A guard in the state machine is missing the required `name` field.

## How to Fix

Add a `name` field to every guard:

```json
{
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "Check that the order is valid"
        }
    }
}
```
