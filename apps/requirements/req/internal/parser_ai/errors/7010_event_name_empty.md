# Event Name Empty (E7010)

An event in the state machine has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The event has a `name` field that is either an empty string (`""`) or contains only whitespace characters. Every event must have a meaningful name.

## File Location

State machine files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.state_machine.json    <-- An event in this file has an empty name
```

## How to Fix

Provide a non-empty, meaningful name for each event:

```json
{
    "events": {
        "submit": {
            "name": "Submit"
        }
    }
}
```

## Understanding the Error Field

The error message includes the path to the problematic event:

| Error Field | Meaning |
|-------------|---------|
| `events.submit.name` | The `submit` event's name is empty |
| `events.approve.name` | The `approve` event's name is empty |

## Invalid Examples

```json
// WRONG: Empty event name
{
    "events": {
        "submit": {
            "name": ""
        }
    }
}

// WRONG: Whitespace-only event name
{
    "events": {
        "submit": {
            "name": "   "
        }
    }
}
```

## Valid Examples

```json
// Simple event
{
    "events": {
        "submit": {
            "name": "Submit"
        }
    }
}

// Event with details and parameters
{
    "events": {
        "approve": {
            "name": "Approve",
            "details": "Manager approves the request",
            "parameters": [
                {
                    "name": "approver_id",
                    "source": "User ID of approving manager"
                }
            ]
        }
    }
}
```

## Event Naming Guidelines

Event names should:
- Describe the action or trigger
- Use verb form (Submit, Approve, Cancel)
- Be concise but descriptive

| Key | Good Name | Avoid |
|-----|-----------|-------|
| `submit` | `Submit` | `S`, `Event1` |
| `approve` | `Approve` | `OK`, `Yes` |
| `payment_received` | `Payment Received` | `PR`, `Money` |

## Related Errors

- **E7002**: Schema violation (general)
- **E7009**: Event name is missing entirely
- **E7012**: Event parameter name is empty
- **E7013**: Event parameter source is empty
