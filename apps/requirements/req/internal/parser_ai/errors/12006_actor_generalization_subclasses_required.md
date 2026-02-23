# Actor Generalization Subclasses Required (E12006)

The actor generalization JSON file is missing the `subclass_keys` field or it is empty.

## What Went Wrong

Every actor generalization must specify at least one subclass (child actor). The `subclass_keys` field must be an array containing at least one valid actor key.

## File Location

Actor generalization files are located in the `actor_generalizations/` directory:

```
your_model/
├── model.json
└── actor_generalizations/
    └── user_types.actor_gen.json    <-- This file has missing or empty subclass_keys
```

## How to Fix

Provide an array with at least one actor key for the `subclass_keys` field:

```json
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer", "guest"]
}
```

## Invalid Examples

```json
// WRONG: Missing subclass_keys
{
    "name": "User Types",
    "superclass_key": "customer"
}

// WRONG: Empty subclass_keys array
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": []
}

// WRONG: subclass_keys is not an array
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": "premium_customer"
}
```

## Valid Examples

```json
// Single subclass (minimum)
{
    "name": "Premium User",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}

// Multiple subclasses
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": [
        "premium_customer",
        "guest",
        "corporate_customer"
    ]
}
```

## Why At Least One Subclass?

An actor generalization defines an inheritance relationship. Without subclasses, there is no inheritance:

```
        customer (superclass)
           ^
    +------+------+
    |             |
premium_customer  guest (subclasses - at least one required!)
```

If an actor has no subclasses, it doesn't need an actor generalization -- it's just a regular actor.

## Understanding Actor Keys

Each entry in `subclass_keys` must reference an existing actor file:

| Actor File Path | Actor Key |
|-----------------|-----------|
| `actors/premium_customer.actor.json` | `premium_customer` |
| `actors/guest.actor.json` | `guest` |
| `actors/corporate_customer.actor.json` | `corporate_customer` |

## Troubleshooting Checklist

1. **Check the array is not empty**: Must have at least one element
2. **Check it's an array**: Use `[]` brackets, not a plain string
3. **Check each key is valid**: Each element should be a non-empty string
4. **Verify the actors exist**: Each key should correspond to an actor file

### Verifying Subclasses Exist

```bash
# For subclass_keys: ["premium_customer", "guest"]
ls actors/premium_customer.actor.json
ls actors/guest.actor.json
```

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `superclass_key` | string | **Yes** | `minLength: 1` |
| `subclass_keys` | string[] | **Yes** | `minItems: 1`, each `minLength: 1` |
| `details` | string | No | None |
| `is_complete` | boolean | No | Default: false |
| `is_static` | boolean | No | Default: false |
| `uml_comment` | string | No | None |

## Related Errors

- **E12001**: Actor generalization name is missing
- **E12004**: Schema violation (general)
- **E12005**: Superclass key is missing or empty
- **E12007**: A subclass key entry is empty
- **E12009**: A subclass not found (reference validation, separate from parsing)
