# Use Case Level Required (E18005)

The use case is missing the required `level` field.

## How to Fix

Add a `level` field. Valid values: `"summary"`, `"user_goal"`, `"subfunction"`.

```json
{
    "name": "Place Order",
    "level": "user_goal"
}
```
