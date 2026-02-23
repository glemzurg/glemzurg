# Actor Generalization Name Required (E12001)

The actor generalization JSON file is missing the required `name` field.

## What Went Wrong

The parser found an actor generalization file but it does not contain a `name` property. Every actor generalization must have a name that identifies the inheritance relationship.

## File Location

Actor generalization files are located in the `actor_generalizations/` directory at the model root:

```
your_model/
├── model.json
├── actors/
│   ├── admin.actor.json
│   ├── customer.actor.json
│   └── guest.actor.json
└── actor_generalizations/
    └── user_types.actor_gen.json    <-- This file is missing the "name" field
```

## How to Fix

Add a `name` field to your actor generalization file:

```json
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer", "guest"]
}
```

## Complete Schema

The actor generalization file accepts these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Human-readable name for the actor generalization |
| `superclass_key` | string | **Yes** | Key of the parent actor |
| `subclass_keys` | string[] | **Yes** | Keys of the child actors (at least one) |
| `details` | string | No | Extended description |
| `is_complete` | boolean | No | Whether all subclasses are listed (default: false) |
| `is_static` | boolean | No | Whether instances can change type (default: false) |
| `uml_comment` | string | No | Comment for UML diagram annotations |

## Why This Field is Required

The actor generalization name is used throughout the system:

- **Generated documentation**: The name appears in inheritance diagrams
- **UML diagrams**: The actor generalization triangle is labeled with this name
- **Error messages**: Errors reference the actor generalization name for context

## Troubleshooting Checklist

1. **Check the file exists**: Ensure the file is in `actor_generalizations/` directory
2. **Check JSON syntax**: The file must be valid JSON
3. **Check field name spelling**: The field must be exactly `"name"` (lowercase)
4. **Check the value exists**: Ensure the name has a value, not just the key

## Common Mistakes

```json
// WRONG: Missing name entirely
{
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}

// WRONG: Typo in field name
{
    "Name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}
```

## Valid Examples

```json
// Minimal valid actor generalization
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}

// Full actor generalization with all fields
{
    "name": "Staff Roles",
    "details": "Different roles a staff member can have in the system.",
    "superclass_key": "staff",
    "subclass_keys": ["admin", "manager", "support_agent"],
    "is_complete": true,
    "is_static": true,
    "uml_comment": "Discriminator: role_type"
}
```

## Related Errors

- **E12002**: Actor generalization name is present but empty
- **E12003**: Invalid JSON syntax
- **E12004**: Schema violation
- **E12005**: Superclass key is missing or empty
- **E12006**: Subclass keys is missing
- **E12007**: A subclass key is empty
