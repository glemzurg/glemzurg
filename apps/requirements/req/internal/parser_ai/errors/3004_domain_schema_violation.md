# Domain Schema Violation (E3004)

The `domain.json` file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your `domain.json` file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing (`name`)
- A field has the wrong type (e.g., string instead of boolean for `realized`)
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string when `minLength: 1` is required)

## File Location

Domain files are located in directories named after the domain:

```
your_model/
├── model.json
└── order_management/
    ├── domain.json             <-- This file violates the schema
    └── ... (classes, etc.)
```

## Schema Requirements

The `domain.json` file must conform to this schema:

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | The domain's display name |

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `details` | string | none | Extended description of the domain |
| `realized` | boolean | none | Whether the domain is realized (default: false) |
| `uml_comment` | string | none | Comment for UML diagram annotations |

### Rules

- **No additional properties**: Only the fields listed above are allowed
- **name must be non-empty**: At least one character required
- **realized must be boolean**: Use `true` or `false`, not strings

## Common Schema Violations

### 1. Missing Required `name` Field

```json
// WRONG: Missing 'name'
{
    "details": "Handles customer orders"
}

// CORRECT: Include 'name'
{
    "name": "Order Management",
    "details": "Handles customer orders"
}
```

### 2. Empty `name` Field

```json
// WRONG: Empty string violates minLength: 1
{
    "name": ""
}

// CORRECT: Non-empty name
{
    "name": "Order Management"
}
```

### 3. Wrong Type for `realized`

```json
// WRONG: String instead of boolean
{
    "name": "Order Management",
    "realized": "yes"
}

// WRONG: Number instead of boolean
{
    "name": "Order Management",
    "realized": 1
}

// CORRECT: Boolean value
{
    "name": "Order Management",
    "realized": true
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'version' is not in the schema
{
    "name": "Order Management",
    "version": "1.0"
}

// WRONG: 'description' is not in the schema (use 'details')
{
    "name": "Order Management",
    "description": "This should be 'details'"
}

// CORRECT: Only allowed fields
{
    "name": "Order Management",
    "details": "This is the correct field for description"
}
```

### 5. Wrong Type for `name`

```json
// WRONG: Number instead of string
{
    "name": 123
}

// WRONG: Array instead of string
{
    "name": ["Order", "Management"]
}

// CORRECT: String value
{
    "name": "Order Management"
}
```

## How to Read Schema Error Messages

Schema validation errors typically include:

1. **Path**: Where in the JSON the error occurred (e.g., `/name`, `/realized`)
2. **Violation type**: What rule was broken
3. **Expected vs actual**: What was expected and what was found

### Example Error Messages

| Error Message | Meaning | Fix |
|--------------|---------|-----|
| `missing properties: 'name'` | The `name` field is required | Add `"name": "..."` |
| `expected string, but got number` | Field value is wrong type | Use a string in quotes |
| `expected boolean, but got string` | `realized` should be true/false | Use `true` or `false` without quotes |
| `string is too short: minLength=1` | Empty string not allowed | Provide at least one character |
| `additionalProperties 'foo' not allowed` | Unknown field | Remove the unknown field |

## Troubleshooting Checklist

1. **Check required fields**: Ensure `name` is present
2. **Check field types**: `name`/`details`/`uml_comment` are strings, `realized` is boolean
3. **Check for extra fields**: Remove any fields not listed in the schema
4. **Check for typos**: Field names are case-sensitive

### Field Name Reference

| You Might Write | Correct Field Name |
|-----------------|-------------------|
| `title` | `name` |
| `description` | `details` |
| `implemented` | `realized` |
| `comment` | `uml_comment` |

## Valid domain.json Examples

### Minimal Valid File

```json
{
    "name": "Order Management"
}
```

### Complete Valid File

```json
{
    "name": "Order Management",
    "details": "Handles the complete lifecycle of customer orders from placement through fulfillment and delivery. Includes order creation, modification, cancellation, and status tracking.",
    "realized": false,
    "uml_comment": "Core business domain"
}
```

### Realized Domain

```json
{
    "name": "Payment Gateway Integration",
    "details": "Wraps the third-party Stripe payment API.",
    "realized": true
}
```

## How Domains Connect to Other Files

```
order_management/                    <-- Domain directory (key: order_management)
├── domain.json                      <-- You are here (defines domain metadata)
│
├── order.class.json                 <-- Classes belong to this domain
├── order_line.class.json            │   Full key: order_management.order
│                                    │   Full key: order_management.order_line
│
└── order.state_machine.json         <-- State machines reference domain classes

associations/
└── customer_order.assoc.json        <-- Associations reference domain classes
    {                                    using keys like:
        "from_class_key": "order_management.order",
        "to_class_key": "user_management.customer"
    }
```

## Related Errors

- **E3001**: Name field is missing entirely (specific case of schema violation)
- **E3002**: Name field is empty (specific case of schema violation)
- **E3003**: JSON syntax is invalid (must fix this before schema can be checked)
