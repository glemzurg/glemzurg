# Transition No States (E7019)

A transition has neither `from_state_key` nor `to_state_key`.

## How to Fix

Every transition must have at least one of:
- `from_state_key`: null for initial transitions (state machine entry)
- `to_state_key`: null for final transitions (state machine exit)

```json
{
    "from_state_key": null,
    "to_state_key": "pending",
    "event_key": "create"
}
```
