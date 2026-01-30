# State Action Key Required (E7006)

A state action in the state machine has an `action_key` field that is empty or contains only whitespace.

## What Went Wrong

A state has an `actions` array, and one of the actions has an `action_key` field that is either an empty string (`""`) or contains only whitespace characters. Every action reference must have a valid action key.

## File Location

State machine files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    ├── order.actions.json          <-- Actions are defined here
    └── order.state_machine.json    <-- An action reference has an empty key
```

## How to Fix

Provide a valid action key that references an action defined in the corresponding actions file:

```json
{
    "states": {
        "pending": {
            "name": "Pending",
            "actions": [
                {
                    "action_key": "notify_manager",
                    "when": "entry"
                }
            ]
        }
    }
}
```

## Understanding the Error Field

The error message includes the path to the problematic action:

| Error Field | Meaning |
|-------------|---------|
| `states.pending.actions[0].action_key` | First action in `pending` state has empty key |
| `states.approved.actions[1].action_key` | Second action in `approved` state has empty key |

## Invalid Examples

```json
// WRONG: Empty action key
{
    "states": {
        "pending": {
            "name": "Pending",
            "actions": [
                {
                    "action_key": "",
                    "when": "entry"
                }
            ]
        }
    }
}

// WRONG: Whitespace-only action key
{
    "states": {
        "pending": {
            "name": "Pending",
            "actions": [
                {
                    "action_key": "   ",
                    "when": "entry"
                }
            ]
        }
    }
}
```

## Valid Examples

```json
// Single action on state entry
{
    "states": {
        "pending": {
            "name": "Pending",
            "actions": [
                {
                    "action_key": "notify_manager",
                    "when": "entry"
                }
            ]
        }
    }
}

// Multiple actions with different triggers
{
    "states": {
        "processing": {
            "name": "Processing",
            "actions": [
                {
                    "action_key": "start_processing",
                    "when": "entry"
                },
                {
                    "action_key": "check_status",
                    "when": "do"
                },
                {
                    "action_key": "cleanup",
                    "when": "exit"
                }
            ]
        }
    }
}
```

## Action Timing

The `when` field determines when the action executes:

| Value | Meaning |
|-------|---------|
| `entry` | Execute once when entering the state |
| `exit` | Execute once when leaving the state |
| `do` | Execute continuously while in the state |

## Related Errors

- **E7002**: Schema violation (general)
- **E7004**: State name is empty
- **E7007**: Action when is missing
- **E7008**: Action when has invalid value
