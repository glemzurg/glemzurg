# Event Name Required (E7009)

An event in the state machine is missing the required `name` field.

## How to Fix

Add a `name` field to every event:

```json
{
    "events": {
        "confirm": {
            "name": "Confirm"
        }
    }
}
```
