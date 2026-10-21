# Transition Initial To Final (E7025)

A transition has both `from_state_key: null` (initial) and `to_state_key: null` (final). A transition cannot be both initial and final.

## How to Fix

An initial transition must go TO a state. A final transition must come FROM a state:

```json
// Initial transition (entering the state machine)
{"from_state_key": null, "to_state_key": "pending", "event_key": "create"}

// Final transition (leaving the state machine)
{"from_state_key": "completed", "to_state_key": null, "event_key": "archive"}
```
