# State Action When Invalid (E7008)

The `when` field of a state action has an invalid value.

## How to Fix

Valid values are: `"entry"`, `"exit"`, `"do"`.

```json
{
    "action_key": "calculate_total",
    "when": "entry"
}
```
