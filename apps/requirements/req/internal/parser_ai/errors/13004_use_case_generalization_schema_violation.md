# Use Case Generalization Schema Violation (E13004)

The use case generalization JSON file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your use case generalization file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing (`name`, `superclass_key`, or `subclass_keys`)
- A field has the wrong type
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string, empty array)

## File Location

Use case generalization files are located in the `use_case_generalizations/` directory:

```
your_model/
├── model.json
└── use_case_generalizations/
    └── order_types.uc_gen.json    <-- This file violates the schema
```

## Schema Requirements

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | Display name for the use case generalization |
| `superclass_key` | string | `minLength: 1` | Key of the parent use case |
| `subclass_keys` | string[] | `minItems: 1`, each `minLength: 1` | Keys of child use cases |

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
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"]
}

// WRONG: Missing 'superclass_key'
{
    "name": "Order Types",
    "subclass_keys": ["process_online_order"]
}

// WRONG: Missing 'subclass_keys'
{
    "name": "Order Types",
    "superclass_key": "process_order"
}

// CORRECT: All required fields
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"]
}
```

### 2. Empty Values

```json
// WRONG: Empty name
{
    "name": "",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"]
}

// WRONG: Empty superclass_key
{
    "name": "Order Types",
    "superclass_key": "",
    "subclass_keys": ["process_online_order"]
}

// WRONG: Empty subclass_keys array
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": []
}

// WRONG: Empty string in subclass_keys
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order", ""]
}
```

### 3. Wrong Types

```json
// WRONG: subclass_keys is a string, not an array
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": "process_online_order"
}

// WRONG: is_complete is a string, not a boolean
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"],
    "is_complete": "true"
}

// CORRECT: Proper types
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"],
    "is_complete": true
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'type' is not in the schema
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"],
    "type": "inheritance"
}

// CORRECT: Only allowed fields
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"]
}
```

## Understanding Use Case Keys

Use case keys match the filename without the extension from the `use_cases/` directory:

```
use_cases/
├── create_order.uc.json            <-- Key: create_order
├── process_order.uc.json           <-- Key: process_order
├── process_online_order.uc.json    <-- Key: process_online_order
└── process_in_store_order.uc.json  <-- Key: process_in_store_order
```

## Valid Examples

### Minimal Valid File

```json
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"]
}
```

### Complete Valid File

```json
{
    "name": "Payment Processing",
    "details": "Different ways a payment can be processed depending on the payment method chosen by the customer.",
    "superclass_key": "process_payment",
    "subclass_keys": ["process_credit_card_payment", "process_bank_transfer", "process_paypal_payment"],
    "is_complete": true,
    "is_static": true,
    "uml_comment": "Discriminator: payment_method"
}
```

## How Use Case Generalizations Connect Use Cases

```
use_case_generalizations/order_types.uc_gen.json
{
    "name": "Order Types",
    "superclass_key": "process_order",              --> use_cases/process_order.uc.json
    "subclass_keys": [
        "process_online_order",                     --> use_cases/process_online_order.uc.json
        "process_in_store_order"                    --> use_cases/process_in_store_order.uc.json
    ]
}
```

The use case generalization creates an inheritance relationship:
- `process_order` is the superclass (parent use case)
- `process_online_order` and `process_in_store_order` are subclasses (child use cases)
- Child use cases inherit behavior from the parent use case

## Related Errors

- **E13001**: Name field is missing
- **E13002**: Name field is empty
- **E13003**: JSON syntax is invalid
- **E13005**: Superclass key is missing or empty
- **E13006**: Subclass keys array is missing
- **E13007**: A subclass key entry is empty
