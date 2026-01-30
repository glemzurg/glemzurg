# Action Name Required (E8001)

The action JSON file is missing the required `name` field.

## What Went Wrong

Every action must have a `name` field that identifies it. The parser found an action file without this required field.

## File Location

Action files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── domain.json
    ├── order.class.json
    └── order.actions.json    <-- This file is missing the name field
```

## How to Fix

Add a `name` field with a descriptive verb phrase for the action:

```json
{
    "name": "Send Confirmation Email"
}
```

## Invalid Examples

```json
// WRONG: Missing name field entirely
{
    "details": "Sends an email to the customer"
}

// WRONG: name is null
{
    "name": null,
    "details": "Sends an email"
}
```

## Valid Examples

```json
// Minimal valid action
{
    "name": "Send Confirmation Email"
}

// Action with details
{
    "name": "Send Confirmation Email",
    "details": "Sends an email to the customer confirming their order"
}

// Full action
{
    "name": "Process Payment",
    "details": "Charges the customer's payment method",
    "requires": [
        "Order total must be greater than zero",
        "Payment method must be valid"
    ],
    "guarantees": [
        "Payment has been charged",
        "Transaction record has been created"
    ]
}
```

## Action Naming Guidelines

Action names should:
- Use verb phrases that describe what the action does
- Be specific and meaningful
- Use title case for multi-word names

| Good Names | Avoid |
|------------|-------|
| `Send Confirmation Email` | `Email`, `Action1` |
| `Reserve Inventory` | `RI`, `Reserve` |
| `Calculate Order Total` | `Calc`, `Total` |

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `details` | string | No | None |
| `requires` | string[] | No | Preconditions |
| `guarantees` | string[] | No | Postconditions |

## Related Errors

- **E8002**: Name is empty or whitespace-only
- **E8003**: Invalid JSON syntax
- **E8004**: Schema violation (general)
