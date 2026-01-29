# Generalization Name Empty (E10002)

The generalization JSON file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The parser found a `name` field in your generalization file, but its value is either an empty string (`""`) or contains only whitespace characters. The generalization name must contain at least one visible character.

## File Location

Generalization files are located in the `generalizations/` directory:

```
your_model/
├── model.json
└── generalizations/
    └── payment_types.gen.json    <-- This file has an empty "name" value
```

## How to Fix

Provide a meaningful, non-empty name for your generalization:

```json
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card", "billing.bank_transfer"]
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
{"name": "Payment Types"}         // Typical name
{"name": "Account Hierarchy"}     // Descriptive name
{"name": "Media Format"}          // Classification name
```

## Choosing a Good Generalization Name

A good generalization name should:

1. **Describe the classification**: What distinguishes the subclasses?
2. **Be concise**: Aim for 2-4 words
3. **Use noun phrases**: e.g., "Payment Types", "Account Hierarchy"
4. **Reflect the discriminator**: What attribute differentiates subclasses?

### Good Name Examples

| Name | Superclass | Subclasses |
|------|------------|------------|
| `"Payment Types"` | Payment | CreditCard, BankTransfer, PayPal |
| `"Account Hierarchy"` | Account | Admin, Standard, Guest |
| `"Media Format"` | Media | Book, Ebook, Audiobook |
| `"Vehicle Categories"` | Vehicle | Car, Truck, Motorcycle |

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

- **E10001**: Generalization name field is missing entirely
- **E10003**: Invalid JSON syntax
- **E10004**: Schema violation
