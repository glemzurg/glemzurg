# State Action When Required (E7007)

A state action is missing the required `when` field.

## How to Fix

Add a `when` field: `"entry"`, `"exit"`, or `"do"`:

```json
{
    "actions": [
        {
            "action_key": "calculate_total",
            "when": "entry"
        }
    ]
}
```
