# Transition Has No States (E11012)

A transition specifies neither a source state nor a target state.

## What Went Wrong

A transition in `state_machine.json` has both `from_state_key` and `to_state_key` set to null or omitted. Every transition must have at least one state reference.

## Valid Transition Types

| Type | from_state_key | to_state_key | Meaning |
|------|----------------|--------------|---------|
| Normal | set | set | Transition between two states |
| Initial | null | set | Entry point into the state machine |
| Final | set | null | Exit point from the state machine |
| Invalid | null | null | **Not allowed** |

## How to Fix

Ensure the transition has at least one state:

### Normal Transition

```json
{
    "from_state_key": "pending",
    "to_state_key": "confirmed",
    "event_key": "confirm"
}
```

### Initial Transition

```json
{
    "from_state_key": null,
    "to_state_key": "pending",
    "event_key": "create"
}
```

### Final Transition

```json
{
    "from_state_key": "completed",
    "to_state_key": null,
    "event_key": "archive"
}
```

## Troubleshooting Checklist

1. **Check for missing state**: Ensure at least one state is specified
2. **Check for typos**: A misspelled key might be treated as missing
3. **Check JSON syntax**: Ensure proper null vs omitted handling

## Common Mistakes

```json
// WRONG: Both states are null
{
    "from_state_key": null,
    "to_state_key": null,
    "event_key": "some_event"
}

// WRONG: Both states omitted
{
    "event_key": "some_event"
}
```

## Related Errors

- **E11008**: State not found
- **E11013**: Transition is both initial and final (redundant with this error)
