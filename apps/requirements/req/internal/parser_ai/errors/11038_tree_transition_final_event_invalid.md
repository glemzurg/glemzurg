# Final Transition Must Use _delete (E11038)

A final transition (`to_state_key: null`) must fire the system finalization event `_delete`.

## What Went Wrong

A transition reaches the final pseudo-state but its `event_key` does not reference an event named `_delete`.

## How to Fix

Declare `_delete` in `events` and reference it on the finalization transition:

```json
{
  "events": {
    "_delete": { "name": "_delete" }
  },
  "transitions": [
    {
      "from_state_key": "active",
      "to_state_key": null,
      "event_key": "_delete"
    }
  ]
}
```

Use `_delete` only for finalization transitions. Other events belong on transitions between ordinary states.