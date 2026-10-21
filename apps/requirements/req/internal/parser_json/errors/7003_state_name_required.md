# State Name Required (E7003)

A state in the state machine is missing the required `name` field.

## How to Fix

Add a `name` field to every state:

```json
{
    "states": {
        "pending": {
            "name": "Pending"
        }
    }
}
```
