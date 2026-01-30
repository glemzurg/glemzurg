# State Machine Guard Not Found (E11010)

A transition references a guard that does not exist in the state machine.

## What Went Wrong

A `state_machine.json` file has a transition with a `guard_key` that references a guard condition, but no guard with that key exists in the `guards` map.

## How Guards Work

Guards are conditions that must be true for a transition to occur. They are optional - a transition can have `guard_key: null` if no condition is needed.

```json
{
    "guards": {
        "has_items": {
            "name": "hasItems",
            "details": "Order has at least one line item"
        },
        "payment_valid": {
            "name": "paymentValid",
            "details": "Payment has been validated"
        }
    },
    "transitions": [
        {
            "from_state_key": "pending",
            "to_state_key": "confirmed",
            "event_key": "confirm",
            "guard_key": "payment_valid"     // Must exist in guards
        }
    ]
}
```

## How to Fix

### Option 1: Add the Missing Guard

Add the guard to the `guards` map:

```json
{
    "guards": {
        "missing_guard": {
            "name": "missingGuard",
            "details": "Description of the condition"
        }
    }
}
```

### Option 2: Fix the Reference

Update the transition to reference an existing guard:

```json
{
    "transitions": [
        {
            "guard_key": "existing_guard"
        }
    ]
}
```

### Option 3: Remove the Guard

If no guard condition is needed, set it to null:

```json
{
    "transitions": [
        {
            "guard_key": null
        }
    ]
}
```

## Troubleshooting Checklist

1. **Check spelling**: Guard keys are case-sensitive
2. **Check guard exists**: The guard must be defined in the `guards` map
3. **Check null handling**: Use `null` if no guard is needed, not an empty string

## Common Mistakes

```json
// WRONG: Using empty string instead of null
{
    "transitions": [
        {"guard_key": "", ...}
    ]
}
// Should be:
{
    "transitions": [
        {"guard_key": null, ...}
    ]
}
```

## Related Errors

- **E11008**: Transition state not found
- **E11009**: Transition event not found
- **E7014**: Guard name is required
