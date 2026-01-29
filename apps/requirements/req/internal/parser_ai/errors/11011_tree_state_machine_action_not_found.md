# State Machine Action Not Found (E11011)

A transition or state action references an action that does not exist in the class.

## What Went Wrong

A `state_machine.json` file references an action that does not exist in the class's `actions/` directory. This can occur in two places:

1. **Transition action**: The `action_key` field in a transition
2. **State action**: An entry in a state's `actions` array (entry/exit/do actions)

## How Actions Work

Actions are defined as separate files in the class's `actions/` directory. The filename (without `.json`) becomes the action key.

```
classes/
└── book_order/
    ├── class.json
    ├── state_machine.json
    └── actions/
        ├── calculate_total.json      <-- Key: "calculate_total"
        ├── notify_warehouse.json     <-- Key: "notify_warehouse"
        └── send_confirmation.json    <-- Key: "send_confirmation"
```

## How to Fix

### Option 1: Create the Missing Action

Create an action file in the class's `actions/` directory:

```
classes/{class}/actions/{action_key}.json
```

With content:

```json
{
    "name": "Action Name",
    "details": "Description of what this action does",
    "requires": [],
    "guarantees": []
}
```

### Option 2: Fix the Reference

Update the transition or state to reference an existing action.

**For transitions:**
```json
{
    "transitions": [
        {
            "action_key": "existing_action"
        }
    ]
}
```

**For state actions:**
```json
{
    "states": {
        "confirmed": {
            "name": "Confirmed",
            "actions": [
                {"action_key": "existing_action", "when": "entry"}
            ]
        }
    }
}
```

### Option 3: Remove the Action Reference

Set `action_key` to null (transitions) or remove the action entry (state actions):

```json
{
    "transitions": [
        {
            "action_key": null
        }
    ]
}
```

## Troubleshooting Checklist

1. **Check action file exists**: `actions/{action_key}.json` must exist
2. **Check spelling**: Action keys are case-sensitive
3. **Check file extension**: Action files must end with `.json`

## Common Mistakes

```json
// WRONG: Including file extension in key
{
    "action_key": "calculate_total.json"
}
// Should be:
{
    "action_key": "calculate_total"
}

// WRONG: Using action name instead of key
{
    "action_key": "Calculate Total"
}
```

## Related Errors

- **E11008**: Transition state not found
- **E11009**: Transition event not found
- **E8001**: Action name is required
