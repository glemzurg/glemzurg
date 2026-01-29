# Event Parameter Source Required (E7013)

An event parameter has a `source` field that is empty or contains only whitespace.

## What Went Wrong

An event has a `parameters` array, and one of the parameters has a `source` field that is either an empty string (`""`) or contains only whitespace characters. Every parameter must describe its source.

## File Location

State machine files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.state_machine.json    <-- An event parameter has an empty source
```

## How to Fix

Provide a descriptive source for each event parameter:

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
| `events.approve.parameters[0].source` | First parameter of `approve` event has empty source |
| `events.reject.parameters[1].source` | Second parameter of `reject` event has empty source |

## Invalid Examples

```json
// WRONG: Empty parameter source
{
    "events": {
        "approve": {
            "name": "Approve",
            "parameters": [
                {
                    "name": "approver_id",
                    "source": ""
                }
            ]
        }
    }
}

// WRONG: Whitespace-only parameter source
{
    "events": {
        "approve": {
            "name": "Approve",
            "parameters": [
                {
                    "name": "approver_id",
                    "source": "   "
                }
            ]
        }
    }
}
```

## Valid Examples

```json
// User input source
{
    "events": {
        "approve": {
            "name": "Approve",
            "parameters": [
                {
                    "name": "notes",
                    "source": "Optional approval notes entered by the manager"
                }
            ]
        }
    }
}

// System-generated source
{
    "events": {
        "timeout": {
            "name": "Timeout",
            "parameters": [
                {
                    "name": "timestamp",
                    "source": "System timestamp when timeout occurred"
                },
                {
                    "name": "elapsed_time",
                    "source": "Duration in seconds since state entry"
                }
            ]
        }
    }
}
```

## What to Include in Source

The source field should describe:
- Where the value comes from (user input, system, external service)
- The data type or format if relevant
- Any constraints on the value

| Parameter | Good Source Description |
|-----------|-------------------------|
| `amount` | `Positive decimal representing currency amount in USD` |
| `user_id` | `User ID from the authentication context` |
| `reason` | `Free-form text explaining the rejection` |

## Related Errors

- **E7002**: Schema violation (general)
- **E7010**: Event name is empty
- **E7012**: Event parameter name is empty
