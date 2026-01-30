# Actor Schema Violation (E2006)

The actor JSON file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your actor file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing (`name` or `type`)
- A field has the wrong type (e.g., number instead of string)
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string when `minLength: 1` is required)

## File Location

Actor files are located in the `actors/` directory at the model root:

```
your_model/
├── model.json
├── actors/
│   └── customer.actor.json    <-- This file violates the schema
├── domains/
└── ...
```

## Schema Requirements

The actor JSON file must conform to this schema:

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | The actor's display name |
| `type` | string | `minLength: 1` | The actor category (human, system, device) |

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `details` | string | none | Extended description of the actor |
| `uml_comment` | string | none | Comment for UML diagram annotations |

### Rules

- **No additional properties**: Only the fields listed above are allowed
- **Both name and type must be non-empty**: At least one character required
- **All values must be strings**: Numbers, booleans, arrays, and objects are not valid

## Common Schema Violations

### 1. Missing Required Fields

```json
// WRONG: Missing 'name'
{
    "type": "human"
}

// WRONG: Missing 'type'
{
    "name": "Customer"
}

// WRONG: Missing both required fields
{
    "details": "Some description"
}

// CORRECT: Both required fields present
{
    "name": "Customer",
    "type": "human"
}
```

### 2. Empty Required Fields

```json
// WRONG: Empty name violates minLength: 1
{
    "name": "",
    "type": "human"
}

// WRONG: Empty type violates minLength: 1
{
    "name": "Customer",
    "type": ""
}

// CORRECT: Non-empty values
{
    "name": "Customer",
    "type": "human"
}
```

### 3. Wrong Types

```json
// WRONG: Number instead of string
{
    "name": 123,
    "type": "human"
}

// WRONG: Boolean instead of string
{
    "name": "Customer",
    "type": true
}

// WRONG: Array instead of string
{
    "name": ["Customer"],
    "type": "human"
}

// CORRECT: String values
{
    "name": "Customer",
    "type": "human"
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'role' is not in the schema
{
    "name": "Customer",
    "type": "human",
    "role": "buyer"
}

// WRONG: 'description' is not in the schema (use 'details')
{
    "name": "Customer",
    "type": "human",
    "description": "A customer"
}

// CORRECT: Only allowed fields
{
    "name": "Customer",
    "type": "human",
    "details": "A customer who purchases products"
}
```

## How to Read Schema Error Messages

Schema validation errors typically include:

1. **Path**: Where in the JSON the error occurred (e.g., `/name`, `/type`)
2. **Violation type**: What rule was broken
3. **Expected vs actual**: What was expected and what was found

### Example Error Messages

| Error Message | Meaning | Fix |
|--------------|---------|-----|
| `missing properties: 'name'` | The `name` field is required | Add `"name": "..."` |
| `missing properties: 'type'` | The `type` field is required | Add `"type": "..."` |
| `expected string, but got number` | Field value is wrong type | Use a string in quotes |
| `string is too short: minLength=1` | Empty string not allowed | Provide at least one character |
| `additionalProperties 'foo' not allowed` | Unknown field | Remove the unknown field |

## Troubleshooting Checklist

1. **Check required fields**: Ensure both `name` and `type` are present
2. **Check field types**: All values must be strings (in double quotes)
3. **Check for extra fields**: Remove any fields not listed in the schema
4. **Check for typos**: Field names are case-sensitive

### Field Name Reference

| You Might Write | Correct Field Name |
|-----------------|-------------------|
| `title` | `name` |
| `description` | `details` |
| `kind` | `type` |
| `category` | `type` |
| `comment` | `uml_comment` |

## How Actors Connect to Other Files

Actors are referenced by classes via the `actor_key` field:

```
actors/customer.actor.json         <-- Actor file with name="Customer", type="human"
    │
    └── Referenced by classes:
        order_management/order.class.json
        {
            "name": "Order",
            "actor_key": "customer"   <-- Key matches filename (without extension)
        }
```

The actor key is derived from the filename:
- `actors/customer.actor.json` → actor key is `customer`
- `actors/payment_gateway.actor.json` → actor key is `payment_gateway`

## Valid Actor Examples

### Minimal Valid File

```json
{
    "name": "Customer",
    "type": "human"
}
```

### Complete Valid File

```json
{
    "name": "Payment Gateway",
    "type": "system",
    "details": "Third-party payment processing service that handles credit card transactions, fraud detection, and payment authorization. Integrates via REST API.",
    "uml_comment": "Requires PCI DSS compliance"
}
```

## Related Errors

- **E2001**: Name field is missing entirely (specific case of schema violation)
- **E2002**: Name field is empty (specific case of schema violation)
- **E2003**: Type field is missing entirely (specific case of schema violation)
- **E2004**: Type field is empty (specific case of schema violation)
- **E2005**: JSON syntax is invalid (must fix this before schema can be checked)
