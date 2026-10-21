# Logic Description Required (E14001)

A logic object is missing the required `description` field.

## What Went Wrong

The parser found a logic object that does not contain a `description` property. Every logic object must have a description that explains the intent of the precondition, postcondition, invariant, safety rule, guard, or derivation policy.

## Where Logic Objects Appear

Logic objects are embedded sub-objects within action, query, state machine, class, and model JSON files. They are used for:

- **Requires**: Preconditions on actions and queries
- **Guarantees**: Postconditions on actions and queries
- **Safety rules**: Constraints on actions
- **Invariants**: Model-level invariants
- **Guards**: Conditions on state machine transitions
- **Derivation policies**: Rules for derived attributes

## How to Fix

Add a `description` field to your logic object:

```json
{
    "description": "Order total must be positive"
}
```

Or with optional fields:

```json
{
    "description": "User must be authenticated",
    "notation": "tla_plus",
    "specification": "user.authenticated = TRUE"
}
```

## Complete Schema

The logic object accepts these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | **Yes** | Human-readable explanation of the logic |
| `notation` | string | No | Formal notation used (e.g., `"tla_plus"`) |
| `specification` | string | No | Formal specification in the given notation |

## Troubleshooting Checklist

1. **Check field name spelling**: The field must be exactly `"description"` (lowercase)
2. **Check the value exists**: Ensure the description has a value, not just the key
3. **Check nesting**: The `description` must be inside the logic object, not at a parent level
4. **Check JSON syntax**: The enclosing file must be valid JSON

## Common Mistakes

```json
// WRONG: Missing description entirely
{
    "notation": "tla_plus",
    "specification": "x > 0"
}

// WRONG: Typo in field name
{
    "Description": "Order total must be positive"
}

// WRONG: Empty object
{}
```

## Valid Examples

```json
// Minimal valid logic object
{"description": "Order total must be positive"}

// Logic object with formal specification
{"description": "User must be authenticated", "notation": "tla_plus", "specification": "user.authenticated = TRUE"}

// Logic object with all fields
{
    "description": "Account balance must not go negative",
    "notation": "tla_plus",
    "specification": "account.balance >= 0"
}
```

## Related Errors

- **E14002**: Logic description is present but empty
- **E14003**: Invalid JSON syntax in a logic object
- **E14004**: Logic object schema violation
