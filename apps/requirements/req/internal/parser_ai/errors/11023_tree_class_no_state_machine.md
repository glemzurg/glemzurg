# Class Has No State Machine (E11023)

Every class must have a state machine defined to describe its lifecycle and behavior.

## What Went Wrong

A class has been defined but does not have an associated `state_machine.json` file. State machines are required to define how instances of the class behave over time.

## Context

State machines describe:
- The **states** an object can be in
- The **events** that cause state changes
- The **transitions** between states
- Optional **guards** (conditions) and **actions** for transitions

```
your_model/
└── domains/
    └── orders/
        └── subdomains/
            └── checkout/
                └── classes/
                    └── order/
                        ├── class.json
                        └── state_machine.json    <-- Required file
```

## How to Fix

### Step 1: Create the State Machine File

Create a `state_machine.json` file in the class directory.

### Step 2: Define States

Identify the different states your class can be in:

```json
{
    "states": {
        "draft": {
            "name": "Draft",
            "details": "Order is being prepared but not yet submitted"
        },
        "pending": {
            "name": "Pending",
            "details": "Order has been submitted, awaiting confirmation"
        },
        "confirmed": {
            "name": "Confirmed",
            "details": "Order has been confirmed and is being processed"
        },
        "completed": {
            "name": "Completed",
            "details": "Order has been fulfilled"
        }
    }
}
```

### Step 3: Define Events

Identify what triggers state changes:

```json
{
    "events": {
        "submit": {
            "name": "Submit",
            "details": "Customer submits the order"
        },
        "confirm": {
            "name": "Confirm",
            "details": "System confirms the order"
        },
        "complete": {
            "name": "Complete",
            "details": "Order fulfillment is finished"
        }
    }
}
```

### Step 4: Define Transitions

Connect states with events:

```json
{
    "transitions": [
        {
            "from_state_key": null,
            "to_state_key": "draft",
            "event_key": "create",
            "uml_comment": "Initial creation"
        },
        {
            "from_state_key": "draft",
            "to_state_key": "pending",
            "event_key": "submit"
        },
        {
            "from_state_key": "pending",
            "to_state_key": "confirmed",
            "event_key": "confirm"
        },
        {
            "from_state_key": "confirmed",
            "to_state_key": "completed",
            "event_key": "complete"
        }
    ]
}
```

## Complete State Machine Example

```json
{
    "states": {
        "draft": {
            "name": "Draft",
            "details": "Order is being prepared"
        },
        "pending": {
            "name": "Pending",
            "details": "Awaiting confirmation"
        },
        "confirmed": {
            "name": "Confirmed",
            "details": "Order confirmed, processing"
        },
        "shipped": {
            "name": "Shipped",
            "details": "Order has been shipped"
        },
        "delivered": {
            "name": "Delivered",
            "details": "Order delivered to customer"
        },
        "cancelled": {
            "name": "Cancelled",
            "details": "Order was cancelled"
        }
    },
    "events": {
        "create": {
            "name": "Create",
            "details": "Create a new order"
        },
        "submit": {
            "name": "Submit",
            "details": "Submit order for processing"
        },
        "confirm": {
            "name": "Confirm",
            "details": "Confirm the order"
        },
        "ship": {
            "name": "Ship",
            "details": "Mark order as shipped"
        },
        "deliver": {
            "name": "Deliver",
            "details": "Mark order as delivered"
        },
        "cancel": {
            "name": "Cancel",
            "details": "Cancel the order"
        }
    },
    "guards": {
        "has_items": {
            "name": "Has Items",
            "details": "Order must have at least one line item"
        },
        "payment_valid": {
            "name": "Payment Valid",
            "details": "Payment information has been validated"
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
            "to_state_key": "pending",
            "event_key": "submit",
            "guard_key": "has_items"
        },
        {
            "from_state_key": "pending",
            "to_state_key": "confirmed",
            "event_key": "confirm",
            "guard_key": "payment_valid"
        },
        {
            "from_state_key": "confirmed",
            "to_state_key": "shipped",
            "event_key": "ship"
        },
        {
            "from_state_key": "shipped",
            "to_state_key": "delivered",
            "event_key": "deliver"
        },
        {
            "from_state_key": "draft",
            "to_state_key": "cancelled",
            "event_key": "cancel"
        },
        {
            "from_state_key": "pending",
            "to_state_key": "cancelled",
            "event_key": "cancel"
        }
    ]
}
```

## State Machine Concepts

### Initial Transitions
Use `from_state_key: null` for the transition that creates new instances:
```json
{
    "from_state_key": null,
    "to_state_key": "draft",
    "event_key": "create"
}
```

### Final Transitions
Use `to_state_key: null` for transitions that terminate the lifecycle:
```json
{
    "from_state_key": "completed",
    "to_state_key": null,
    "event_key": "archive"
}
```

### Guards (Conditions)
Guards are conditions that must be true for a transition to occur:
```json
{
    "from_state_key": "pending",
    "to_state_key": "confirmed",
    "event_key": "confirm",
    "guard_key": "payment_valid"
}
```

## Related Errors

- **E7001**: State machine JSON is invalid
- **E7002**: State machine schema violation
- **E11024**: State machine has no transitions
