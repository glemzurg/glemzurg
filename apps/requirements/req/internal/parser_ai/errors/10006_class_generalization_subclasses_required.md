# Class Generalization Subclasses Required (E10006)

The class generalization JSON file is missing the `subclass_keys` field or it is empty.

## What Went Wrong

Every class generalization must specify at least one subclass (child class). The `subclass_keys` field must be an array containing at least one valid class key.

## File Location

Class generalization files are located in the `generalizations/` directory:

```
your_model/
├── model.json
└── generalizations/
    └── payment_types.gen.json    <-- This file has missing or empty subclass_keys
```

## How to Fix

Provide an array with at least one class key for the `subclass_keys` field:

```json
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card", "billing.bank_transfer"]
}
```

## Invalid Examples

```json
// WRONG: Missing subclass_keys
{
    "name": "Payment Types",
    "superclass_key": "billing.payment"
}

// WRONG: Empty subclass_keys array
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": []
}

// WRONG: subclass_keys is not an array
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": "billing.credit_card"
}
```

## Valid Examples

```json
// Single subclass (minimum)
{
    "name": "Premium Product",
    "superclass_key": "catalog.product",
    "subclass_keys": ["catalog.premium_product"]
}

// Multiple subclasses
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": [
        "billing.credit_card",
        "billing.bank_transfer",
        "billing.paypal"
    ]
}
```

## Why At Least One Subclass?

A class generalization defines an inheritance relationship. Without subclasses, there is no inheritance:

```
        Payment (superclass)
           ↑
    ┌──────┴──────┐
    │             │
CreditCard   BankTransfer (subclasses - at least one required!)
```

If a class has no subclasses, it doesn't need a class generalization — it's just a regular class.

## Understanding Class Keys

Each entry in `subclass_keys` must reference an existing class file:

| Class File Path | Class Key |
|-----------------|-----------|
| `billing/credit_card.class.json` | `billing.credit_card` |
| `billing/bank_transfer.class.json` | `billing.bank_transfer` |
| `billing/methods/paypal.class.json` | `billing.methods.paypal` |

## Troubleshooting Checklist

1. **Check the array is not empty**: Must have at least one element
2. **Check it's an array**: Use `[]` brackets, not a plain string
3. **Check each key is valid**: Each element should be a non-empty string
4. **Verify the classes exist**: Each key should correspond to a class file

### Verifying Subclasses Exist

```bash
# For subclass_keys: ["billing.credit_card", "billing.bank_transfer"]
ls billing/credit_card.class.json
ls billing/bank_transfer.class.json
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

- **E10001**: Class generalization name is missing
- **E10004**: Schema violation (general)
- **E10005**: Superclass key is missing or empty
- **E10007**: A subclass key entry is empty
- **E10009**: A subclass not found (reference validation, separate from parsing)
