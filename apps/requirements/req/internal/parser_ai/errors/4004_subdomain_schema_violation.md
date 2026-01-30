# Subdomain Schema Violation (E4004)

The `subdomain.json` file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your `subdomain.json` file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing (`name`)
- A field has the wrong type (e.g., number instead of string)
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string when `minLength: 1` is required)

## File Location

Subdomain files are located within domain directories:

```
your_model/
├── model.json
└── order_management/
    ├── domain.json
    └── fulfillment/
        ├── subdomain.json            <-- This file violates the schema
        └── ... (classes, etc.)
```

## Schema Requirements

The `subdomain.json` file must conform to this schema:

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | The subdomain's display name |

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `details` | string | none | Extended description of the subdomain |
| `uml_comment` | string | none | Comment for UML diagram annotations |

### Rules

- **No additional properties**: Only the fields listed above are allowed
- **name must be non-empty**: At least one character required
- **All values must be strings**: Numbers, booleans, arrays, and objects are not valid

## Common Schema Violations

### 1. Missing Required `name` Field

```json
// WRONG: Missing 'name'
{
    "details": "Handles order fulfillment"
}

// CORRECT: Include 'name'
{
    "name": "Order Fulfillment",
    "details": "Handles order fulfillment"
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
    "name": "Order Fulfillment"
}
```

### 3. Wrong Type for `name`

```json
// WRONG: Number instead of string
{
    "name": 123
}

// WRONG: Array instead of string
{
    "name": ["Order", "Fulfillment"]
}

// CORRECT: String value
{
    "name": "Order Fulfillment"
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'version' is not in the schema
{
    "name": "Order Fulfillment",
    "version": "1.0"
}

// WRONG: 'description' is not in the schema (use 'details')
{
    "name": "Order Fulfillment",
    "description": "This should be 'details'"
}

// WRONG: 'realized' is not valid for subdomains (only domains)
{
    "name": "Order Fulfillment",
    "realized": true
}

// CORRECT: Only allowed fields
{
    "name": "Order Fulfillment",
    "details": "This is the correct field for description"
}
```

## How to Read Schema Error Messages

Schema validation errors typically include:

1. **Path**: Where in the JSON the error occurred (e.g., `/name`)
2. **Violation type**: What rule was broken
3. **Expected vs actual**: What was expected and what was found

### Example Error Messages

| Error Message | Meaning | Fix |
|--------------|---------|-----|
| `missing properties: 'name'` | The `name` field is required | Add `"name": "..."` |
| `expected string, but got number` | Field value is wrong type | Use a string in quotes |
| `string is too short: minLength=1` | Empty string not allowed | Provide at least one character |
| `additionalProperties 'foo' not allowed` | Unknown field | Remove the unknown field |

## Troubleshooting Checklist

1. **Check required fields**: Ensure `name` is present
2. **Check field types**: All values must be strings (in double quotes)
3. **Check for extra fields**: Remove any fields not listed in the schema
4. **Check for typos**: Field names are case-sensitive
5. **Don't confuse with domain**: Subdomains don't have `realized` field

### Field Name Reference

| You Might Write | Correct Field Name |
|-----------------|-------------------|
| `title` | `name` |
| `description` | `details` |
| `comment` | `uml_comment` |

### Subdomain vs Domain Fields

| Field | Domain | Subdomain |
|-------|--------|-----------|
| `name` | Yes | Yes |
| `details` | Yes | Yes |
| `realized` | Yes | **No** |
| `uml_comment` | Yes | Yes |

## Valid subdomain.json Examples

### Minimal Valid File

```json
{
    "name": "Order Fulfillment"
}
```

### Complete Valid File

```json
{
    "name": "Order Fulfillment",
    "details": "Handles the picking, packing, and shipping of customer orders. Coordinates with warehouse management systems and shipping carrier APIs.",
    "uml_comment": "Integrates with WMS and carrier APIs"
}
```

## How Subdomains Connect to Other Files

```
order_management/                         <-- Domain (key: order_management)
├── domain.json
└── fulfillment/                          <-- Subdomain (key: fulfillment)
    ├── subdomain.json                    <-- You are here
    │
    ├── shipment.class.json               <-- Classes belong to this subdomain
    ├── package.class.json                │   Full key: order_management.fulfillment.shipment
    │                                     │   Full key: order_management.fulfillment.package
    │
    └── shipment.state_machine.json       <-- State machines reference subdomain classes

associations/
└── fulfillment_inventory.assoc.json      <-- Cross-domain associations use full keys
    {
        "from_class_key": "order_management.fulfillment.shipment",
        "to_class_key": "inventory.warehouse.stock_item"
    }
```

## Related Errors

- **E4001**: Name field is missing entirely (specific case of schema violation)
- **E4002**: Name field is empty (specific case of schema violation)
- **E4003**: JSON syntax is invalid (must fix this before schema can be checked)
