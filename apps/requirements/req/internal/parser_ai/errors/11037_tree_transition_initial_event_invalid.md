# Initial Transition Must Use _new (E11037)

An initial transition (`from_state_key: null`) must fire the system creation event `_new`.

## What Went Wrong

A transition leaves the initial pseudo-state but its `event_key` does not reference an event named `_new`.

## How to Fix

Declare `_new` in `events` and reference it on the creation transition:

```json
{
  "events": {
    "_new": { "name": "_new" }
  },
  "transitions": [
    {
      "from_state_key": null,
      "to_state_key": "active",
      "event_key": "_new"
    }
  ]
}
```

Use `_new` only for creation transitions. Other events belong on transitions between ordinary states.