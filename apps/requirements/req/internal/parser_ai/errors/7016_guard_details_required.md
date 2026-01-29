# Guard Details Required (E7016)

A guard in the state machine has a `details` field that is empty or contains only whitespace.

## What Went Wrong

The guard has a `details` field that is either an empty string (`""`) or contains only whitespace characters. Every guard must have a meaningful description of the condition it checks.

## File Location

State machine files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.state_machine.json    <-- A guard in this file has empty details
```

## How to Fix

Provide a clear description of the guard condition:

```json
{
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "Order passes all validation rules: has at least one item, total is positive, and shipping address is complete"
        }
    }
}
```

## Understanding the Error Field

The error message includes the path to the problematic guard:

| Error Field | Meaning |
|-------------|---------|
| `guards.is_valid.details` | The `is_valid` guard's details is empty |
| `guards.has_budget.details` | The `has_budget` guard's details is empty |

## Invalid Examples

```json
// WRONG: Empty guard details
{
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": ""
        }
    }
}

// WRONG: Whitespace-only guard details
{
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "   "
        }
    }
}
```

## Valid Examples

```json
// Simple condition
{
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "Order passes all validation rules"
        }
    }
}

// Detailed condition
{
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "Order validation: (1) has at least one line item, (2) total amount is positive, (3) shipping address has all required fields, (4) payment method is set"
        }
    }
}

// Condition with thresholds
{
    "guards": {
        "under_limit": {
            "name": "Under Approval Limit",
            "details": "Order total is less than $1,000, which is the threshold for automatic approval without manager review"
        }
    }
}
```

## Why Details Are Required

Guard details are required because:
1. Guards are boolean conditions that need to be implemented
2. The name alone is often not specific enough
3. Developers need to know exactly what conditions to check
4. Business rules should be documented at the source

## What to Include in Details

| Include | Example |
|---------|---------|
| The exact condition | "total > 0 AND items.count > 0" |
| Threshold values | "amount is less than $1,000" |
| Business context | "required for regulatory compliance" |
| Edge cases | "null values are treated as invalid" |

## Related Errors

- **E7002**: Schema violation (general)
- **E7014**: Guard name is missing
- **E7015**: Guard name is empty
