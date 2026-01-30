# State Machine Has No Transitions (E11024)

Every state machine must have at least one transition to define how the class moves between states.

## What Went Wrong

A state machine has been defined with states and/or events but the `transitions` array is empty. Transitions are essential - they define how and when the class changes from one state to another.

## Context

Transitions connect states via events. Without transitions, states exist but there's no way to move between them.

```json
{
    "states": { ... },
    "events": { ... },
    "transitions": [           <-- Must have at least one transition
        {
            "from_state_key": "pending",
            "to_state_key": "confirmed",
            "event_key": "confirm"
        }
    ]
}
```

## How to Fix

### Step 1: Define Your Initial Transition

Every class needs at least one way to be created:

```json
{
    "transitions": [
        {
            "from_state_key": null,
            "to_state_key": "active",
            "event_key": "create"
        }
    ]
}
```

### Step 2: Map Your Lifecycle

Add transitions for each state change:

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
            "to_state_key": "pending",
            "event_key": "submit"
        },
        {
            "from_state_key": "pending",
            "to_state_key": "approved",
            "event_key": "approve"
        },
        {
            "from_state_key": "pending",
            "to_state_key": "rejected",
            "event_key": "reject"
        }
    ]
}
```

## Transition Components

### Required Fields

- **event_key**: The event that triggers this transition

### State Fields (at least one required)

- **from_state_key**: State before transition (null for initial)
- **to_state_key**: State after transition (null for final)

### Optional Fields

- **guard_key**: Condition that must be true
- **action_key**: Action to execute during transition
- **uml_comment**: Documentation note

## Transition Patterns

### Initial Transition (Creation)
```json
{
    "from_state_key": null,
    "to_state_key": "draft",
    "event_key": "create"
}
```

### Standard Transition
```json
{
    "from_state_key": "pending",
    "to_state_key": "confirmed",
    "event_key": "confirm"
}
```

### Guarded Transition
```json
{
    "from_state_key": "pending",
    "to_state_key": "confirmed",
    "event_key": "confirm",
    "guard_key": "payment_valid"
}
```

### Transition with Action
```json
{
    "from_state_key": "pending",
    "to_state_key": "confirmed",
    "event_key": "confirm",
    "action_key": "send_confirmation_email"
}
```

### Final Transition (Termination)
```json
{
    "from_state_key": "completed",
    "to_state_key": null,
    "event_key": "archive"
}
```

### Self-Transition (Same State)
```json
{
    "from_state_key": "active",
    "to_state_key": "active",
    "event_key": "update",
    "action_key": "record_modification"
}
```

## Common Lifecycle Patterns

### Simple CRUD
```json
{
    "transitions": [
        { "from_state_key": null, "to_state_key": "active", "event_key": "create" },
        { "from_state_key": "active", "to_state_key": "active", "event_key": "update" },
        { "from_state_key": "active", "to_state_key": null, "event_key": "delete" }
    ]
}
```

### Draft-Publish
```json
{
    "transitions": [
        { "from_state_key": null, "to_state_key": "draft", "event_key": "create" },
        { "from_state_key": "draft", "to_state_key": "published", "event_key": "publish" },
        { "from_state_key": "published", "to_state_key": "draft", "event_key": "unpublish" },
        { "from_state_key": "draft", "to_state_key": null, "event_key": "delete" }
    ]
}
```

### Approval Workflow
```json
{
    "transitions": [
        { "from_state_key": null, "to_state_key": "draft", "event_key": "create" },
        { "from_state_key": "draft", "to_state_key": "pending_review", "event_key": "submit" },
        { "from_state_key": "pending_review", "to_state_key": "approved", "event_key": "approve" },
        { "from_state_key": "pending_review", "to_state_key": "rejected", "event_key": "reject" },
        { "from_state_key": "rejected", "to_state_key": "draft", "event_key": "revise" }
    ]
}
```

### Order Fulfillment
```json
{
    "transitions": [
        { "from_state_key": null, "to_state_key": "pending", "event_key": "place_order" },
        { "from_state_key": "pending", "to_state_key": "confirmed", "event_key": "confirm" },
        { "from_state_key": "confirmed", "to_state_key": "processing", "event_key": "start_processing" },
        { "from_state_key": "processing", "to_state_key": "shipped", "event_key": "ship" },
        { "from_state_key": "shipped", "to_state_key": "delivered", "event_key": "deliver" },
        { "from_state_key": "pending", "to_state_key": "cancelled", "event_key": "cancel" },
        { "from_state_key": "confirmed", "to_state_key": "cancelled", "event_key": "cancel" }
    ]
}
```

## Validation Rules

1. Every transition must reference valid state and event keys
2. Cannot have both `from_state_key` and `to_state_key` as null
3. Guard keys must reference defined guards
4. Action keys must reference defined actions

## Related Errors

- **E11008**: Transition references non-existent state
- **E11009**: Transition references non-existent event
- **E11010**: Transition references non-existent guard
- **E11011**: Transition references non-existent action
- **E11012**: Transition has neither from nor to state
