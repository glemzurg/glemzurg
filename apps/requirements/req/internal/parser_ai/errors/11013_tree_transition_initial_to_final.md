# Transition Is Both Initial And Final (E11013)

A transition specifies neither a source state nor a target state, making it both an initial and final transition simultaneously.

## What Went Wrong

A transition in `state_machine.json` has both `from_state_key` and `to_state_key` set to null. This would create a meaningless transition that enters and exits the state machine without visiting any state.

## How to Fix

A valid transition must have at least one state. Choose the appropriate pattern:

### Initial Transition (entering the state machine)

```json
{
    "from_state_key": null,
    "to_state_key": "first_state",
    "event_key": "create"
}
```

### Final Transition (leaving the state machine)

```json
{
    "from_state_key": "last_state",
    "to_state_key": null,
    "event_key": "complete"
}
```

### Normal Transition (between states)

```json
{
    "from_state_key": "state_a",
    "to_state_key": "state_b",
    "event_key": "proceed"
}
```

## Related Errors

- **E11012**: Transition has no states
- **E11008**: State not found
