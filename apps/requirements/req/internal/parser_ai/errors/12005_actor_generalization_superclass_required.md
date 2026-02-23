# Actor Generalization Superclass Required (E12005)

The actor generalization JSON file has a `superclass_key` field that is missing, empty, or contains only whitespace.

## What Went Wrong

Every actor generalization must specify a superclass (parent actor) that the subclasses inherit from. The `superclass_key` field must contain a valid, non-empty actor key.

## File Location

Actor generalization files are located in the `actor_generalizations/` directory:

```
your_model/
├── model.json
└── actor_generalizations/
    └── user_types.actor_gen.json    <-- This file has invalid superclass_key
```

## How to Fix

Provide a valid actor key for the `superclass_key` field:

```json
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer", "guest"]
}
```

## Invalid Examples

```json
// WRONG: Missing superclass_key
{
    "name": "User Types",
    "subclass_keys": ["premium_customer"]
}

// WRONG: Empty superclass_key
{
    "name": "User Types",
    "superclass_key": "",
    "subclass_keys": ["premium_customer"]
}

// WRONG: Whitespace-only superclass_key
{
    "name": "User Types",
    "superclass_key": "   ",
    "subclass_keys": ["premium_customer"]
}
```

## Valid Examples

```json
// Simple actor key
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}

// Another example
{
    "name": "Staff Hierarchy",
    "superclass_key": "staff",
    "subclass_keys": ["admin", "manager"]
}
```

## Understanding Actor Keys

The `superclass_key` must reference an existing actor file. Actor keys match the filename (without the extension):

```
actors/                               <-- Actors directory
├── admin.actor.json                  <-- Key: admin
├── customer.actor.json               <-- Key: customer
├── staff.actor.json                  <-- Key: staff
└── guest.actor.json                  <-- Key: guest
```

### Examples

| Actor File Path | Actor Key |
|-----------------|-----------|
| `actors/customer.actor.json` | `customer` |
| `actors/admin.actor.json` | `admin` |
| `actors/staff.actor.json` | `staff` |

## What is a Superclass?

In an actor generalization (inheritance) relationship:

- **Superclass** (parent): The general actor that defines common capabilities
- **Subclasses** (children): Specialized actors that inherit from and extend the superclass

```
        customer (superclass)
           ^
    +------+------+
    |             |
premium_customer  guest (subclasses)
```

The superclass typically:
- Defines shared capabilities (e.g., browse, purchase)
- Represents the abstract actor concept
- May or may not be instantiable on its own

## Troubleshooting Checklist

1. **Check the actor exists**: Verify the superclass actor file exists at the expected path
2. **Check the key format**: The key is the filename without the `.actor.json` extension
3. **Check for typos**: Actor keys are case-sensitive
4. **Don't include file extension**: Use `customer`, not `customer.actor.json`

### Verifying the Superclass Exists

```bash
# If superclass_key is "customer", check:
ls actors/customer.actor.json

# If superclass_key is "staff", check:
ls actors/staff.actor.json
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
- **E12006**: Subclass keys is missing
- **E12007**: A subclass key is empty
- **E12008**: Superclass not found (reference validation, separate from parsing)
