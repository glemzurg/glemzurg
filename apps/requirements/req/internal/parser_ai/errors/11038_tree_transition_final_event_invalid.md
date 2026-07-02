# Final Transition Must Use _destroy (E11038)

A final transition (`to_state_key: null`) must fire the system finalization event `_destroy`.

## What Went Wrong

A transition reaches the final pseudo-state but its `event_key` does not reference an event named `_destroy`.

## How to Fix

Declare `_destroy` in `events` and reference it on the finalization transition:

```json
{
  "events": {
    "_destroy": { "name": "_destroy" }
  },
  "transitions": [
    {
      "from_state_key": "active",
      "to_state_key": null,
      "event_key": "_destroy"
    }
  ]
}
```

Use `_destroy` only for finalization transitions. Other events belong on transitions between ordinary states.