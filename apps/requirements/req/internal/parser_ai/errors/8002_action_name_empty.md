# Action Name Empty (E8002)

The action JSON file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The `name` field exists but is either an empty string (`""`) or contains only whitespace characters. Every action must have a meaningful name.

## File Location

Action files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.actions.json    <-- This file has an empty name
```

## How to Fix

Provide a non-empty, meaningful name for the action:

```json
{
    "name": "Send Confirmation Email"
}
```

## Invalid Examples

```json
// WRONG: Empty string
{
    "name": ""
}

// WRONG: Whitespace only
{
    "name": "   "
}

// WRONG: Tab characters only
{
    "name": "\t\t"
}
```

## Valid Examples

```json
// Simple name
{
    "name": "Send Email"
}

// Multi-word name
{
    "name": "Reserve Inventory Items"
}

// Full action
{
    "name": "Process Payment",
    "details": "Charges the customer's payment method",
    "requires": ["Order total > 0"],
    "guarantees": ["Payment charged"]
}
```

## Action Naming Guidelines

Action names should:
- Use verb phrases that describe the action
- Be specific and meaningful
- Use title case for multi-word names
- Describe what the action does, not how

| Good Names | Avoid |
|------------|-------|
| `Send Confirmation Email` | `Email` |
| `Reserve Inventory` | `RI` |
| `Calculate Order Total` | `Calc` |
| `Notify Customer` | `Notify` |

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `details` | string | No | None |
| `requires` | string[] | No | Preconditions |
| `guarantees` | string[] | No | Postconditions |

## Related Errors

- **E8001**: Name field is missing entirely
- **E8003**: Invalid JSON syntax
- **E8004**: Schema violation (general)
