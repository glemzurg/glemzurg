# State Machine State Not Found (E11008)

A transition references a state that does not exist in the state machine.

## What Went Wrong

A `state_machine.json` file has a transition with a `from_state_key` or `to_state_key` that references a state key, but no state with that key exists in the `states` map.

## How State Machines Work

State machines define states and transitions between them. Transitions reference states by their key in the `states` map.

```json
{
    "states": {
        "pending": {"name": "Pending"},
        "confirmed": {"name": "Confirmed"},
        "shipped": {"name": "Shipped"}
    },
    "transitions": [
        {
            "from_state_key": "pending",      // Must exist in states
            "to_state_key": "confirmed",      // Must exist in states
            "event_key": "confirm"
        }
    ]
}
```

## How to Fix

### Option 1: Add the Missing State

Add the state to the `states` map:

```json
{
    "states": {
        "missing_state": {
            "name": "Missing State",
            "details": "Description of this state"
        }
    }
}
```

### Option 2: Fix the Reference

Update the transition to reference an existing state:

```json
{
    "transitions": [
        {
            "from_state_key": "existing_state",
            "to_state_key": "another_state",
            "event_key": "some_event"
        }
    ]
}
```

### Option 3: Use Initial/Final Transitions

- Set `from_state_key` to `null` for an **initial transition** (entering a state from outside)
- Set `to_state_key` to `null` for a **final transition** (leaving to outside)

```json
{
    "transitions": [
        {
            "from_state_key": null,           // Initial transition
            "to_state_key": "pending",
            "event_key": "create"
        }
    ]
}
```

## Troubleshooting Checklist

1. **Check spelling**: State keys are case-sensitive
2. **Check state exists**: The state must be defined in the `states` map
3. **Check null vs missing**: Use `null` intentionally for initial/final transitions

## Common Mistakes

```json
// WRONG: Using state name instead of key
{
    "states": {
        "order_pending": {"name": "Order Pending"}
    },
    "transitions": [
        {"from_state_key": "Order Pending", ...}
    ]
}
// Should be:
{
    "transitions": [
        {"from_state_key": "order_pending", ...}
    ]
}
```

## Related Errors

- **E11009**: Transition event not found
- **E11010**: Transition guard not found
- **E11012**: Transition has no states (neither from nor to)
