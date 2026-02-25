# Use Case Generalization Superclass Required (E13005)

The use case generalization JSON file has a `superclass_key` field that is missing, empty, or contains only whitespace.

## What Went Wrong

Every use case generalization must specify a superclass (parent use case) that the child use cases inherit from. The `superclass_key` field must contain a valid, non-empty use case key.

## File Location

Use case generalization files are located in the `use_case_generalizations/` directory:

```
your_model/
├── model.json
└── use_case_generalizations/
    └── order_types.uc_gen.json    <-- This file has invalid superclass_key
```

## How to Fix

Provide a valid use case key for the `superclass_key` field:

```json
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order", "process_in_store_order"]
}
```

## Invalid Examples

```json
// WRONG: Missing superclass_key
{
    "name": "Order Types",
    "subclass_keys": ["process_online_order"]
}

// WRONG: Empty superclass_key
{
    "name": "Order Types",
    "superclass_key": "",
    "subclass_keys": ["process_online_order"]
}

// WRONG: Whitespace-only superclass_key
{
    "name": "Order Types",
    "superclass_key": "   ",
    "subclass_keys": ["process_online_order"]
}
```

## Valid Examples

```json
// Simple use case key
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"]
}

// Another valid example
{
    "name": "Account Management",
    "superclass_key": "manage_account",
    "subclass_keys": ["manage_admin_account", "manage_standard_account"]
}
```

## Understanding Use Case Keys

The `superclass_key` must reference an existing use case file. The key matches the filename without the extension:

```
use_cases/
├── process_order.uc.json            <-- Key: process_order
├── process_online_order.uc.json     <-- Key: process_online_order
└── process_in_store_order.uc.json   <-- Key: process_in_store_order
```

### Examples

| Use Case File Path | Use Case Key |
|---------------------|--------------|
| `use_cases/create_order.uc.json` | `create_order` |
| `use_cases/process_payment.uc.json` | `process_payment` |
| `use_cases/manage_account.uc.json` | `manage_account` |

## What is a Superclass Use Case?

In a use case generalization (inheritance) relationship:

- **Superclass** (parent): The general use case that defines common behavior
- **Subclasses** (children): Specialized use cases that inherit from and extend the parent

```
      process_order (superclass)
           ^
    +------+------+
    |             |
process_online  process_in_store (subclasses)
```

The superclass use case typically:
- Defines shared steps and behavior
- Represents the abstract use case concept
- May or may not be invocable on its own

## Troubleshooting Checklist

1. **Check the use case exists**: Verify the superclass use case file exists at the expected path
2. **Check the key format**: Use the filename without the `.uc.json` extension
3. **Check for typos**: Use case keys are case-sensitive
4. **Don't include file extension**: Use `process_order`, not `process_order.uc.json`

### Verifying the Superclass Exists

```bash
# If superclass_key is "process_order", check:
ls use_cases/process_order.uc.json

# If superclass_key is "manage_account", check:
ls use_cases/manage_account.uc.json
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

- **E13001**: Use case generalization name is missing
- **E13004**: Schema violation (general)
- **E13006**: Subclass keys is missing
- **E13007**: A subclass key is empty
- **E13008**: Superclass not found (reference validation, separate from parsing)
