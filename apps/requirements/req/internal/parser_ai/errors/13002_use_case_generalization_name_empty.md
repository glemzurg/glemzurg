# Use Case Generalization Name Empty (E13002)

The use case generalization JSON file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The parser found a `name` field in your use case generalization file, but its value is either an empty string (`""`) or contains only whitespace characters. The use case generalization name must contain at least one visible character.

## File Location

Use case generalization files are located in the `use_case_generalizations/` directory:

```
your_model/
├── model.json
└── use_case_generalizations/
    └── order_types.uc_gen.json    <-- This file has an empty "name" value
```

## How to Fix

Provide a meaningful, non-empty name for your use case generalization:

```json
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order", "process_in_store_order"]
}
```

## Invalid Examples

These values will all trigger this error:

```json
{"name": ""}              // Empty string
{"name": "   "}           // Spaces only
{"name": "\t"}            // Tab only
```

## Valid Examples

The name must contain at least one non-whitespace character:

```json
{"name": "Order Types"}           // Typical name
{"name": "Account Management"}    // Descriptive name
{"name": "Payment Processing"}    // Classification name
```

## Choosing a Good Use Case Generalization Name

A good use case generalization name should:

1. **Describe the classification**: What distinguishes the child use cases?
2. **Be concise**: Aim for 2-4 words
3. **Use noun phrases**: e.g., "Order Types", "Payment Processing"
4. **Reflect the discriminator**: What attribute differentiates the child use cases?

### Good Name Examples

| Name | Superclass | Subclasses |
|------|------------|------------|
| `"Order Types"` | process_order | process_online_order, process_in_store_order |
| `"Account Management"` | manage_account | manage_admin_account, manage_standard_account |
| `"Payment Processing"` | process_payment | process_credit_card_payment, process_bank_transfer |
| `"Report Generation"` | generate_report | generate_sales_report, generate_inventory_report |

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `superclass_key` | string | **Yes** | `minLength: 1` |
| `subclass_keys` | string[] | **Yes** | `minItems: 1`, each item `minLength: 1` |
| `details` | string | No | None |
| `is_complete` | boolean | No | Default: false |
| `is_static` | boolean | No | Default: false |
| `uml_comment` | string | No | None |

## Related Errors

- **E13001**: Use case generalization name field is missing entirely
- **E13003**: Invalid JSON syntax
- **E13004**: Schema violation
