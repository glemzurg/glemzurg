# State Machine Has No Transitions (E11024)

Every state machine must have at least one transition to define how the class moves between states.

## What Went Wrong

A state machine has states and/or events defined but the `transitions` array is empty. Without transitions, there is no way to move between states.

## How to Fix

Add at least one transition to the `transitions` array in `state_machine.json`. A transition connects a source state to a destination state via an event:

```json
{
    "transitions": [
        {
            "from_state_key": null,
            "to_state_key": "draft",
            "event_key": "create"
        },
        {
            "from_state_key": "draft",
            "to_state_key": "active",
            "event_key": "activate",
            "guard_key": "is_valid",
            "action_key": "send_notification"
        }
    ]
}
```

### Transition Fields

- **from_state_key**: State before transition (`null` for initial/creation transitions)
- **to_state_key**: State after transition (`null` for final/termination transitions)
- **event_key**: The event that triggers this transition (required)
- **guard_key**: Optional condition that must be true for the transition to fire
- **action_key**: Optional action to execute during the transition

All keys must reference states, events, guards, and actions defined elsewhere in the same state machine.

## Related Errors

- **E11023**: Class has no state machine
- **E11008**: Transition references non-existent state
- **E11009**: Transition references non-existent event
