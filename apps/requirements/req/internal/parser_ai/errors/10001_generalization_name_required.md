# Generalization Name Required (E10001)

The generalization JSON file is missing the required `name` field.

## What Went Wrong

The parser found a generalization file but it does not contain a `name` property. Every generalization must have a name that identifies the inheritance relationship.

## File Location

Generalization files are located in the `generalizations/` directory at the model root:

```
your_model/
├── model.json
├── actors/
├── order_management/
│   └── ... (classes)
└── generalizations/
    └── payment_types.gen.json    <-- This file is missing the "name" field
```

## How to Fix

Add a `name` field to your generalization file:

```json
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card", "billing.bank_transfer"]
}
```

## Complete Schema

The generalization file accepts these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Human-readable name for the generalization |
| `superclass_key` | string | **Yes** | Key of the parent class |
| `subclass_keys` | string[] | **Yes** | Keys of the child classes (at least one) |
| `details` | string | No | Extended description |
| `is_complete` | boolean | No | Whether all subclasses are listed (default: false) |
| `is_static` | boolean | No | Whether instances can change type (default: false) |
| `uml_comment` | string | No | Comment for UML diagram annotations |

## Why This Field is Required

The generalization name is used throughout the system:

- **Generated documentation**: The name appears in inheritance diagrams
- **UML diagrams**: The generalization triangle is labeled with this name
- **Error messages**: Errors reference the generalization name for context

## Troubleshooting Checklist

1. **Check the file exists**: Ensure the file is in `generalizations/` directory
2. **Check JSON syntax**: The file must be valid JSON
3. **Check field name spelling**: The field must be exactly `"name"` (lowercase)
4. **Check the value exists**: Ensure the name has a value, not just the key

## Common Mistakes

```json
// WRONG: Missing name entirely
{
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"]
}

// WRONG: Typo in field name
{
    "Name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"]
}
```

## Valid Examples

```json
// Minimal valid generalization
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"]
}

// Full generalization with all fields
{
    "name": "Media Format",
    "details": "Different formats a book can be published in.",
    "superclass_key": "catalog.media",
    "subclass_keys": ["catalog.book", "catalog.ebook", "catalog.audiobook"],
    "is_complete": true,
    "is_static": true,
    "uml_comment": "Discriminator: format_type"
}
```

## Related Errors

- **E10002**: Generalization name is present but empty
- **E10003**: Invalid JSON syntax
- **E10004**: Schema violation
- **E10005**: Superclass key is missing or empty
- **E10006**: Subclass keys is missing
- **E10007**: A subclass key is empty
