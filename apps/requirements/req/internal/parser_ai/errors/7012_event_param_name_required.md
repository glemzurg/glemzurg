# Event Parameter Name Required (E7012)

An event parameter has a `name` field that is empty or contains only whitespace.

## What Went Wrong

An event has a `parameters` array, and one of the parameters has a `name` field that is either an empty string (`""`) or contains only whitespace characters. Every parameter must have a meaningful name.

## File Location

State machine files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.state_machine.json    <-- An event parameter has an empty name
```

## How to Fix

Provide a non-empty name for each event parameter:

```json
{
    "events": {
        "approve": {
            "name": "Approve",
            "parameters": [
                {
                    "name": "approver_id",
                    "source": "User ID of the approving manager"
                }
            ]
        }
    }
}
```

## Understanding the Error Field

The error message includes the path to the problematic parameter:

| Error Field | Meaning |
|-------------|---------|
| `events.approve.parameters[0].name` | First parameter of `approve` event has empty name |
| `events.reject.parameters[1].name` | Second parameter of `reject` event has empty name |

## Invalid Examples

```json
// WRONG: Empty parameter name
{
    "events": {
        "approve": {
            "name": "Approve",
            "parameters": [
                {
                    "name": "",
                    "source": "User input"
                }
            ]
        }
    }
}

// WRONG: Whitespace-only parameter name
{
    "events": {
        "approve": {
            "name": "Approve",
            "parameters": [
                {
                    "name": "   ",
                    "source": "User input"
                }
            ]
        }
    }
}
```

## Valid Examples

```json
// Single parameter
{
    "events": {
        "approve": {
            "name": "Approve",
            "parameters": [
                {
                    "name": "approver_id",
                    "source": "User ID of the approving manager"
                }
            ]
        }
    }
}

// Multiple parameters
{
    "events": {
        "reject": {
            "name": "Reject",
            "parameters": [
                {
                    "name": "reason",
                    "source": "Text explaining why the request was rejected"
                },
                {
                    "name": "rejection_date",
                    "source": "System timestamp"
                }
            ]
        }
    }
}
```

## Parameter Naming Guidelines

Parameter names should:
- Use snake_case
- Be descriptive of the data
- Match the terminology used in guards and actions

| Good Names | Avoid |
|------------|-------|
| `approver_id` | `id`, `a` |
| `rejection_reason` | `reason1`, `r` |
| `amount` | `amt`, `$` |

## Related Errors

- **E7002**: Schema violation (general)
- **E7010**: Event name is empty
- **E7013**: Event parameter source is empty
