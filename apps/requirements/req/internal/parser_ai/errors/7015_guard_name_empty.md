# Guard Name Empty (E7015)

A guard in the state machine has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The guard has a `name` field that is either an empty string (`""`) or contains only whitespace characters. Every guard must have a meaningful name.

## File Location

State machine files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.state_machine.json    <-- A guard in this file has an empty name
```

## How to Fix

Provide a non-empty, meaningful name for each guard:

```json
{
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "Order passes all validation rules"
        }
    }
}
```

## Understanding the Error Field

The error message includes the path to the problematic guard:

| Error Field | Meaning |
|-------------|---------|
| `guards.is_valid.name` | The `is_valid` guard's name is empty |
| `guards.has_budget.name` | The `has_budget` guard's name is empty |

## Invalid Examples

```json
// WRONG: Empty guard name
{
    "guards": {
        "is_valid": {
            "name": "",
            "details": "Validation check"
        }
    }
}

// WRONG: Whitespace-only guard name
{
    "guards": {
        "is_valid": {
            "name": "   ",
            "details": "Validation check"
        }
    }
}
```

## Valid Examples

```json
// Simple guard
{
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "Order passes all validation rules"
        }
    }
}

// Multiple guards
{
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "Order has items, valid addresses, and positive total"
        },
        "has_budget": {
            "name": "Has Budget",
            "details": "Department budget can cover the order total"
        },
        "under_limit": {
            "name": "Under Approval Limit",
            "details": "Order total is under $1000 auto-approval threshold"
        }
    }
}
```

## Guard Naming Guidelines

Guard names should:
- Describe the condition being checked
- Start with "Is", "Has", "Can", etc. when appropriate
- Be displayed in square brackets on transitions: `[Is Valid]`

| Key | Good Name | Avoid |
|-----|-----------|-------|
| `is_valid` | `Is Valid` | `Valid`, `V` |
| `has_inventory` | `Has Inventory` | `Inv`, `Check1` |
| `can_proceed` | `Can Proceed` | `OK`, `Yes` |

## Related Errors

- **E7002**: Schema violation (general)
- **E7014**: Guard name is missing entirely
- **E7016**: Guard details is missing or empty
