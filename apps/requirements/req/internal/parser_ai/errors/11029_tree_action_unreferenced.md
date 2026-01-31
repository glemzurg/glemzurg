# Action Unreferenced (E11029)

An action JSON file exists but is not referenced by any state action or transition in the state machine.

## What Went Wrong

Every action defined in a class must be used in the state machine. Actions exist to be executed either:
- As **state actions** (entry, exit, or do actions when entering, leaving, or while in a state)
- As **transition actions** (executed when a transition fires)

If an action file exists but is never referenced, it serves no purpose and may indicate:
- A forgotten reference in the state machine
- An action that should be deleted
- A typo in the action key

## How to Fix

### Option 1: Reference the Action in a State

Add the action to a state's entry, exit, or do actions:

```json
{
    "states": {
        "pending": {
            "name": "Pending",
            "actions": [
                {
                    "action_key": "calculate_total",
                    "when": "entry"
                }
            ]
        }
    }
}
```

Valid `when` values:
- `"entry"` - Execute when entering the state
- `"exit"` - Execute when leaving the state
- `"do"` - Execute while in the state (continuous activity)

### Option 2: Reference the Action in a Transition

Add the action to a transition:

```json
{
    "transitions": [
        {
            "from_state_key": "pending",
            "to_state_key": "confirmed",
            "event_key": "confirm",
            "action_key": "calculate_total"
        }
    ]
}
```

### Option 3: Delete the Unreferenced Action

If the action is no longer needed, delete the action file:

```bash
rm domains/{domain}/subdomains/{subdomain}/classes/{class}/actions/{action}.json
```

## Example

### File Structure

```
classes/book_order/
├── class.json
├── state_machine.json
└── actions/
    ├── calculate_total.json    <- Referenced in state
    ├── send_notification.json  <- Referenced in transition
    └── unused_action.json      <- ERROR: Not referenced anywhere
```

### State Machine with References

```json
{
    "states": {
        "pending": {
            "name": "Pending",
            "actions": [
                {
                    "action_key": "calculate_total",
                    "when": "entry"
                }
            ]
        },
        "confirmed": {
            "name": "Confirmed"
        }
    },
    "events": {
        "confirm": {
            "name": "confirm"
        }
    },
    "transitions": [
        {
            "from_state_key": "pending",
            "to_state_key": "confirmed",
            "event_key": "confirm",
            "action_key": "send_notification"
        }
    ]
}
```

In this example:
- `calculate_total` is referenced as a state entry action ✓
- `send_notification` is referenced as a transition action ✓
- `unused_action` is not referenced anywhere ✗

## Why This Validation Exists

### 1. Model Completeness

A well-formed model has no orphaned elements. Every action should have a purpose and be connected to the state machine's behavior.

### 2. Code Generation

Generated code may create methods for each action. Unreferenced actions result in dead code that:
- Clutters the codebase
- May confuse developers
- Requires maintenance for no benefit

### 3. Documentation Accuracy

Actions describe what the system does. Unreferenced actions mislead readers about the class's actual behavior.

### 4. Early Error Detection

An unreferenced action often indicates a mistake:
- A typo in the action key reference
- A forgotten transition or state action
- An incomplete refactoring

Catching this early prevents bugs and confusion later.

## Common Mistakes

### Typo in Action Key

```json
// In state_machine.json - WRONG
"action_key": "calculate_totl"

// Action file is named calculate_total.json
// The reference has a typo, so calculate_total appears unreferenced
```

### Forgetting to Add State Action

```json
// State without actions - the action was intended to run on entry
{
    "states": {
        "pending": {
            "name": "Pending"
            // Missing: "actions": [{"action_key": "initialize", "when": "entry"}]
        }
    }
}
```

### Removing Transition but Keeping Action

After refactoring, a transition was removed but its action file remained:

```bash
# The transition that used "old_action" was removed
# But the file still exists:
actions/old_action.json  # Should be deleted
```

## Related Errors

- **E11011**: State machine action not found (the reverse - reference exists but action doesn't)
- **E8001**: Action name required
- **E8002**: Action name empty
