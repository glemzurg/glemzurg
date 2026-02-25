# Use Case Generalization Name Required (E13001)

The use case generalization JSON file is missing the required `name` field.

## What Went Wrong

The parser found a use case generalization file but it does not contain a `name` property. Every use case generalization must have a name that identifies the inheritance relationship.

## File Location

Use case generalization files are located in the `use_case_generalizations/` directory at the model root:

```
your_model/
├── model.json
├── actors/
├── use_cases/
│   └── create_order.uc.json
└── use_case_generalizations/
    └── order_types.uc_gen.json    <-- This file is missing the "name" field
```

## How to Fix

Add a `name` field to your use case generalization file:

```json
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order", "process_in_store_order"]
}
```

## Complete Schema

The use case generalization file accepts these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Human-readable name for the use case generalization |
| `superclass_key` | string | **Yes** | Key of the parent use case |
| `subclass_keys` | string[] | **Yes** | Keys of the child use cases (at least one) |
| `details` | string | No | Extended description |
| `is_complete` | boolean | No | Whether all subclasses are listed (default: false) |
| `is_static` | boolean | No | Whether instances can change type (default: false) |
| `uml_comment` | string | No | Comment for UML diagram annotations |

## Why This Field is Required

The use case generalization name is used throughout the system:

- **Generated documentation**: The name appears in inheritance diagrams
- **UML diagrams**: The use case generalization triangle is labeled with this name
- **Error messages**: Errors reference the use case generalization name for context

## Troubleshooting Checklist

1. **Check the file exists**: Ensure the file is in `use_case_generalizations/` directory
2. **Check JSON syntax**: The file must be valid JSON
3. **Check field name spelling**: The field must be exactly `"name"` (lowercase)
4. **Check the value exists**: Ensure the name has a value, not just the key

## Common Mistakes

```json
// WRONG: Missing name entirely
{
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"]
}

// WRONG: Typo in field name
{
    "Name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"]
}
```

## Valid Examples

```json
// Minimal valid use case generalization
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"]
}

// Full use case generalization with all fields
{
    "name": "Payment Processing",
    "details": "Different ways a payment can be processed depending on the method.",
    "superclass_key": "process_payment",
    "subclass_keys": ["process_credit_card_payment", "process_bank_transfer", "process_paypal_payment"],
    "is_complete": true,
    "is_static": true,
    "uml_comment": "Discriminator: payment_method"
}
```

## Related Errors

- **E13002**: Use case generalization name is present but empty
- **E13003**: Invalid JSON syntax
- **E13004**: Schema violation
- **E13005**: Superclass key is missing or empty
- **E13006**: Subclass keys is missing
- **E13007**: A subclass key is empty
