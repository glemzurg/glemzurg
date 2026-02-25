# Actor Generalization Schema Violation (E12004)

The actor generalization JSON file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your actor generalization file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing (`name`, `superclass_key`, or `subclass_keys`)
- A field has the wrong type
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string, empty array)

## File Location

Actor generalization files are located in the `actor_generalizations/` directory:

```
your_model/
├── model.json
└── actor_generalizations/
    └── user_types.actor_gen.json    <-- This file violates the schema
```

## Schema Requirements

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | Display name for the actor generalization |
| `superclass_key` | string | `minLength: 1` | Key of the parent actor |
| `subclass_keys` | string[] | `minItems: 1`, each `minLength: 1` | Keys of child actors |

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `details` | string | none | Extended description |
| `is_complete` | boolean | none | All subclasses listed? (default: false) |
| `is_static` | boolean | none | Instances can't change type? (default: false) |
| `uml_comment` | string | none | UML diagram annotation |

## Common Schema Violations

### 1. Missing Required Fields

```json
// WRONG: Missing 'name'
{
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}

// WRONG: Missing 'superclass_key'
{
    "name": "User Types",
    "subclass_keys": ["premium_customer"]
}

// WRONG: Missing 'subclass_keys'
{
    "name": "User Types",
    "superclass_key": "customer"
}

// CORRECT: All required fields
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}
```

### 2. Empty Values

```json
// WRONG: Empty name
{
    "name": "",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}

// WRONG: Empty superclass_key
{
    "name": "User Types",
    "superclass_key": "",
    "subclass_keys": ["premium_customer"]
}

// WRONG: Empty subclass_keys array
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": []
}

// WRONG: Empty string in subclass_keys
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer", ""]
}
```

### 3. Wrong Types

```json
// WRONG: subclass_keys is a string, not an array
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": "premium_customer"
}

// WRONG: is_complete is a string, not a boolean
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"],
    "is_complete": "true"
}

// CORRECT: Proper types
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"],
    "is_complete": true
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'type' is not in the schema
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"],
    "type": "inheritance"
}

// CORRECT: Only allowed fields
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}
```

## Understanding Actor Keys

Actor keys match the filename of the actor file (without the extension). Actor files are in the `actors/` directory:

```
actors/                               <-- Actors directory
├── admin.actor.json                  <-- Key: admin
├── customer.actor.json               <-- Key: customer
├── premium_customer.actor.json       <-- Key: premium_customer
└── guest.actor.json                  <-- Key: guest
```

## Valid Examples

### Minimal Valid File

```json
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}
```

### Complete Valid File

```json
{
    "name": "Staff Roles",
    "details": "Different roles a staff member can have in the system. Once assigned, a staff member cannot change their role type.",
    "superclass_key": "staff",
    "subclass_keys": ["admin", "manager", "support_agent"],
    "is_complete": true,
    "is_static": true,
    "uml_comment": "Discriminator: role_type"
}
```

## How Actor Generalizations Connect Actors

```
actor_generalizations/user_types.actor_gen.json
{
    "name": "User Types",
    "superclass_key": "customer",            --> actors/customer.actor.json
    "subclass_keys": [
        "premium_customer",                  --> actors/premium_customer.actor.json
        "guest"                              --> actors/guest.actor.json
    ]
}
```

The actor generalization creates an inheritance relationship:
- `customer` is the superclass (parent)
- `premium_customer` and `guest` are subclasses (children)
- Subclasses inherit capabilities from the superclass

## Related Errors

- **E12001**: Name field is missing
- **E12002**: Name field is empty
- **E12003**: JSON syntax is invalid
- **E12005**: Superclass key is missing or empty
- **E12006**: Subclass keys array is missing
- **E12007**: A subclass key entry is empty
