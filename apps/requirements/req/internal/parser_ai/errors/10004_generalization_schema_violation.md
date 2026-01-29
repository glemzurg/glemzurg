# Generalization Schema Violation (E10004)

The generalization JSON file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your generalization file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing (`name`, `superclass_key`, or `subclass_keys`)
- A field has the wrong type
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string, empty array)

## File Location

Generalization files are located in the `generalizations/` directory:

```
your_model/
├── model.json
└── generalizations/
    └── payment_types.gen.json    <-- This file violates the schema
```

## Schema Requirements

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | Display name for the generalization |
| `superclass_key` | string | `minLength: 1` | Key of the parent class |
| `subclass_keys` | string[] | `minItems: 1`, each `minLength: 1` | Keys of child classes |

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
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"]
}

// WRONG: Missing 'superclass_key'
{
    "name": "Payment Types",
    "subclass_keys": ["billing.credit_card"]
}

// WRONG: Missing 'subclass_keys'
{
    "name": "Payment Types",
    "superclass_key": "billing.payment"
}

// CORRECT: All required fields
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"]
}
```

### 2. Empty Values

```json
// WRONG: Empty name
{
    "name": "",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"]
}

// WRONG: Empty superclass_key
{
    "name": "Payment Types",
    "superclass_key": "",
    "subclass_keys": ["billing.credit_card"]
}

// WRONG: Empty subclass_keys array
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": []
}

// WRONG: Empty string in subclass_keys
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card", ""]
}
```

### 3. Wrong Types

```json
// WRONG: subclass_keys is a string, not an array
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": "billing.credit_card"
}

// WRONG: is_complete is a string, not a boolean
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"],
    "is_complete": "true"
}

// CORRECT: Proper types
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"],
    "is_complete": true
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'type' is not in the schema
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"],
    "type": "inheritance"
}

// CORRECT: Only allowed fields
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"]
}
```

## Understanding Class Keys

Class keys follow the `domain.class_name` pattern:

```
billing/                          <-- Domain directory
├── domain.json
├── payment.class.json            <-- Key: billing.payment
├── credit_card.class.json        <-- Key: billing.credit_card
└── bank_transfer.class.json      <-- Key: billing.bank_transfer
```

For nested subdomains:
```
billing/
└── methods/                      <-- Subdomain
    ├── subdomain.json
    └── card.class.json           <-- Key: billing.methods.card
```

## Valid Examples

### Minimal Valid File

```json
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"]
}
```

### Complete Valid File

```json
{
    "name": "Media Format",
    "details": "Different formats a book can be published in. Once created, a product cannot change its format type.",
    "superclass_key": "catalog.media",
    "subclass_keys": ["catalog.book", "catalog.ebook", "catalog.audiobook"],
    "is_complete": true,
    "is_static": true,
    "uml_comment": "Discriminator: format_type"
}
```

## How Generalizations Connect Classes

```
generalizations/payment_types.gen.json
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",        --> billing/payment.class.json
    "subclass_keys": [
        "billing.credit_card",                  --> billing/credit_card.class.json
        "billing.bank_transfer"                 --> billing/bank_transfer.class.json
    ]
}
```

The generalization creates an inheritance relationship:
- `Payment` is the superclass (parent)
- `CreditCard` and `BankTransfer` are subclasses (children)
- Subclasses inherit attributes from the superclass

## Related Errors

- **E10001**: Name field is missing
- **E10002**: Name field is empty
- **E10003**: JSON syntax is invalid
- **E10005**: Superclass key is missing or empty
- **E10006**: Subclass keys array is missing
- **E10007**: A subclass key entry is empty
