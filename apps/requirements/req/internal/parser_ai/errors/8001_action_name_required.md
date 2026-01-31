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
| `details` | string | No | Human-readable summary only |
| `requires` | string[] | No | Preconditions (logic goes here) |
| `guarantees` | string[] | No | Postconditions (logic goes here) |

## Important: Where Logic Belongs

**The `details` field is for human-readable summaries only — NOT for describing logic.**

| DO put in `details` | DON'T put in `details` |
|---------------------|------------------------|
| Brief description | Preconditions |
| What it accomplishes | Postconditions |
| High-level purpose | Business rules |
| | Validation logic |
| | "If X then Y" statements |

**Wrong** — logic stuffed into details:
```json
{
    "name": "Process Payment",
    "details": "If order has items and total > 0, charges the payment method. Must have valid payment method. Creates transaction record on success."
}
```

**Correct** — logic in structured fields:
```json
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

## Related Errors

- **E8002**: Name is empty or whitespace-only
- **E8003**: Invalid JSON syntax
- **E8004**: Schema violation (general)
