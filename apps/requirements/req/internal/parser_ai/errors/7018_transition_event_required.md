# Transition Event Required (E7018)

A transition in the state machine has an `event_key` field that is empty or contains only whitespace.

## What Went Wrong

A transition has an `event_key` field that is either an empty string (`""`) or contains only whitespace characters. Every transition must specify which event triggers it.

## File Location

State machine files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.state_machine.json    <-- A transition has an empty event_key
```

## How to Fix

Provide a valid event key that references an event defined in the same state machine:

```json
{
    "events": {
        "approve": {
            "name": "Approve"
        }
    },
    "transitions": [
        {
            "from_state_key": "pending",
            "to_state_key": "approved",
            "event_key": "approve"
        }
    ]
}
```

## Understanding the Error Field

The error message includes the index of the problematic transition:

| Error Field | Meaning |
|-------------|---------|
| `transitions[0].event_key` | First transition has empty event_key |
| `transitions[2].event_key` | Third transition has empty event_key |

## Invalid Examples

```json
// WRONG: Empty event_key
{
    "transitions": [
        {
            "from_state_key": "pending",
            "to_state_key": "approved",
            "event_key": ""
        }
    ]
}

// WRONG: Whitespace-only event_key
{
    "transitions": [
        {
            "from_state_key": "pending",
            "to_state_key": "approved",
            "event_key": "   "
        }
    ]
}
```

## Valid Examples

```json
// Simple transition
{
    "events": {
        "submit": {
            "name": "Submit"
        }
    },
    "transitions": [
        {
            "from_state_key": null,
            "to_state_key": "pending",
            "event_key": "submit"
        }
    ]
}

// Transition with guard and action
{
    "events": {
        "approve": {
            "name": "Approve"
        }
    },
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "Passes validation"
        }
    },
    "transitions": [
        {
            "from_state_key": "pending",
            "to_state_key": "approved",
            "event_key": "approve",
            "guard_key": "is_valid",
            "action_key": "send_notification"
        }
    ]
}
```

## Transition Structure

Every transition requires an event that triggers it:

```
   [from_state] --( event [guard] / action )--> [to_state]
                      ↑
                 REQUIRED
```

- `from_state_key`: null for initial transition, or a state key
- `to_state_key`: null for final transition, or a state key
- `event_key`: **Required** - must reference an event
- `guard_key`: Optional - condition that must be true
- `action_key`: Optional - action to execute during transition

## Related Errors

- **E7002**: Schema violation (general)
- **E7009**: Event name is missing
- **E7010**: Event name is empty
- **E7022**: Event key references non-existent event (reference validation)
