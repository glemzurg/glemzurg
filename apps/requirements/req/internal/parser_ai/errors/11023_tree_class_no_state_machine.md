# Class Has No State Machine (E11023)

Every class must have a state machine defined to describe its lifecycle and behavior.

## What Went Wrong

A class directory does not contain a `state_machine.json` file. Every class requires one.

```
classes/
└── order/
    ├── class.json
    └── state_machine.json    <-- Required file
```

## How to Fix

Create a `state_machine.json` file in the class directory with states, events, and transitions:

```json
{
    "states": {
        "draft": {
            "name": "Draft",
            "details": "Order is being prepared"
        },
        "confirmed": {
            "name": "Confirmed",
            "details": "Order has been confirmed"
        }
    },
    "events": {
        "create": {
            "name": "Create",
            "details": "Create a new order"
        },
        "confirm": {
            "name": "Confirm",
            "details": "Confirm the order"
        }
    },
    "guards": {
        "has_items": {
            "name": "Has Items",
            "details": "Order must have at least one line item"
        }
    },
    "transitions": [
        {
            "from_state_key": null,
            "to_state_key": "draft",
            "event_key": "create"
        },
        {
            "from_state_key": "draft",
            "to_state_key": "confirmed",
            "event_key": "confirm",
            "guard_key": "has_items"
        }
    ]
}
```

Use `from_state_key: null` for the initial creation transition. Use `to_state_key: null` for a final/termination transition. Guards and actions are optional per transition.

### Classes Without an Obvious Lifecycle

If a class has no status or lifecycle states, create a minimal state machine with a single `existing` state and a creation event:

```json
{
    "states": {
        "existing": {
            "name": "Existing",
            "details": "The entity exists in the system"
        }
    },
    "events": {
        "existing": {
            "name": "existing",
            "details": "Initial event that creates the entity"
        }
    },
    "transitions": [
        {
            "from_state_key": null,
            "to_state_key": "existing",
            "event_key": "existing"
        }
    ]
}
```

## Related Errors

- **E7001**: State machine JSON is invalid
- **E11024**: State machine has no transitions
