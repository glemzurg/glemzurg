# State Name Empty (E7004)

A state in the state machine has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The state has a `name` field that is either an empty string (`""`) or contains only whitespace characters. Every state must have a meaningful name.

## File Location

State machine files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.state_machine.json    <-- A state in this file has an empty name
```

## How to Fix

Provide a non-empty, meaningful name for each state:

```json
{
    "states": {
        "pending": {
            "name": "Pending"
        }
    }
}
```

## Understanding the Error Field

The error message includes the path to the problematic state:

| Error Field | Meaning |
|-------------|---------|
| `states.pending.name` | The `pending` state's name is empty |
| `states.approved.name` | The `approved` state's name is empty |

## Invalid Examples

```json
// WRONG: Empty state name
{
    "states": {
        "pending": {
            "name": ""
        }
    }
}

// WRONG: Whitespace-only state name
{
    "states": {
        "pending": {
            "name": "   "
        }
    }
}
```

## Valid Examples

```json
// Simple state
{
    "states": {
        "pending": {
            "name": "Pending"
        }
    }
}

// State with details
{
    "states": {
        "pending": {
            "name": "Pending Approval",
            "details": "Order is waiting for manager approval"
        }
    }
}
```

## State Naming Guidelines

State names should:
- Describe the condition or status of the object
- Use title case for multi-word names
- Be unique and distinguishable

| Key | Good Name | Avoid |
|-----|-----------|-------|
| `pending` | `Pending` | `P`, `State1` |
| `in_progress` | `In Progress` | `IP`, `Working` |
| `completed` | `Completed` | `Done`, `C` |

## Related Errors

- **E7002**: Schema violation (general)
- **E7003**: State name is missing entirely
- **E7006**: State action key is empty
