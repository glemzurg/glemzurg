# Class Generalization Superclass Required (E10005)

The class generalization JSON file has a `superclass_key` field that is missing, empty, or contains only whitespace.

## What Went Wrong

Every class generalization must specify a superclass (parent class) that the subclasses inherit from. The `superclass_key` field must contain a valid, non-empty class key.

## File Location

Class generalization files are located in the `generalizations/` directory:

```
your_model/
├── model.json
└── generalizations/
    └── payment_types.gen.json    <-- This file has invalid superclass_key
```

## How to Fix

Provide a valid class key for the `superclass_key` field:

```json
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card", "billing.bank_transfer"]
}
```

## Invalid Examples

```json
// WRONG: Missing superclass_key
{
    "name": "Payment Types",
    "subclass_keys": ["billing.credit_card"]
}

// WRONG: Empty superclass_key
{
    "name": "Payment Types",
    "superclass_key": "",
    "subclass_keys": ["billing.credit_card"]
}

// WRONG: Whitespace-only superclass_key
{
    "name": "Payment Types",
    "superclass_key": "   ",
    "subclass_keys": ["billing.credit_card"]
}
```

## Valid Examples

```json
// Single-level domain
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"]
}

// Nested subdomain
{
    "name": "Account Types",
    "superclass_key": "users.accounts.account",
    "subclass_keys": ["users.accounts.admin", "users.accounts.standard"]
}
```

## Understanding Class Keys

The `superclass_key` must reference an existing class file. The key format is:

```
{domain}.{class_name}
```

Or for nested subdomains:
```
{domain}.{subdomain}.{class_name}
```

### Examples

| Class File Path | Class Key |
|-----------------|-----------|
| `billing/payment.class.json` | `billing.payment` |
| `users/account.class.json` | `users.account` |
| `billing/methods/card.class.json` | `billing.methods.card` |

## What is a Superclass?

In a class generalization (inheritance) relationship:

- **Superclass** (parent): The general category that defines common attributes
- **Subclasses** (children): Specialized types that inherit from and extend the superclass

```
        Payment (superclass)
           ↑
    ┌──────┴──────┐
    │             │
CreditCard   BankTransfer (subclasses)
```

The superclass typically:
- Defines shared attributes (e.g., `amount`, `date`)
- Represents the abstract concept
- May or may not be instantiable on its own

## Troubleshooting Checklist

1. **Check the class exists**: Verify the superclass file exists at the expected path
2. **Check the key format**: Use dots to separate domain/subdomain/class
3. **Check for typos**: Class keys are case-sensitive
4. **Don't include file extension**: Use `billing.payment`, not `billing.payment.class.json`

### Verifying the Superclass Exists

```bash
# If superclass_key is "billing.payment", check:
ls billing/payment.class.json

# If superclass_key is "billing.methods.card", check:
ls billing/methods/card.class.json
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
- **E10006**: Subclass keys is missing
- **E10007**: A subclass key is empty
- **E10008**: Superclass not found (reference validation, separate from parsing)
