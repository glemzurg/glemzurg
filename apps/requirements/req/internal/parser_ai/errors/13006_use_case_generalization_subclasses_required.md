# Use Case Generalization Subclasses Required (E13006)

The use case generalization JSON file is missing the `subclass_keys` field or it is empty.

## What Went Wrong

Every use case generalization must specify at least one subclass (child use case). The `subclass_keys` field must be an array containing at least one valid use case key.

## File Location

Use case generalization files are located in the `use_case_generalizations/` directory:

```
your_model/
├── model.json
└── use_case_generalizations/
    └── order_types.uc_gen.json    <-- This file has missing or empty subclass_keys
```

## How to Fix

Provide an array with at least one use case key for the `subclass_keys` field:

```json
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order", "process_in_store_order"]
}
```

## Invalid Examples

```json
// WRONG: Missing subclass_keys
{
    "name": "Order Types",
    "superclass_key": "process_order"
}

// WRONG: Empty subclass_keys array
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": []
}

// WRONG: subclass_keys is not an array
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": "process_online_order"
}
```

## Valid Examples

```json
// Single subclass (minimum)
{
    "name": "Premium Order Processing",
    "superclass_key": "process_order",
    "subclass_keys": ["process_premium_order"]
}

// Multiple subclasses
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": [
        "process_online_order",
        "process_in_store_order",
        "process_phone_order"
    ]
}
```

## Why At Least One Subclass?

A use case generalization defines an inheritance relationship. Without child use cases, there is no inheritance:

```
      process_order (superclass)
           ^
    +------+------+
    |             |
process_online  process_in_store (subclasses - at least one required!)
```

If a use case has no child use cases, it doesn't need a generalization -- it's just a regular use case.

## Understanding Use Case Keys

Each entry in `subclass_keys` must reference an existing use case file:

| Use Case File Path | Use Case Key |
|---------------------|--------------|
| `use_cases/process_online_order.uc.json` | `process_online_order` |
| `use_cases/process_in_store_order.uc.json` | `process_in_store_order` |
| `use_cases/process_phone_order.uc.json` | `process_phone_order` |

## Troubleshooting Checklist

1. **Check the array is not empty**: Must have at least one element
2. **Check it's an array**: Use `[]` brackets, not a plain string
3. **Check each key is valid**: Each element should be a non-empty string
4. **Verify the use cases exist**: Each key should correspond to a use case file

### Verifying Subclasses Exist

```bash
# For subclass_keys: ["process_online_order", "process_in_store_order"]
ls use_cases/process_online_order.uc.json
ls use_cases/process_in_store_order.uc.json
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
- **E13005**: Superclass key is missing or empty
- **E13007**: A subclass key entry is empty
- **E13009**: A subclass not found (reference validation, separate from parsing)
