# Class Schema Violation (E5004)

The class JSON file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your class file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing (`name`)
- A field has the wrong type
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string, empty array)

## File Location

Class files are located within domain directories:

```
your_model/
├── model.json
└── order_management/           <-- Domain directory
    ├── domain.json
    └── order.class.json        <-- This file violates the schema
```

## Schema Requirements

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | Display name for the class |

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `details` | string | None | Extended description |
| `actor_key` | string | None | Reference to owning actor |
| `uml_comment` | string | None | UML diagram annotation |
| `attributes` | object | See below | Map of attribute definitions |
| `indexes` | array | See below | List of index definitions |

### Attribute Schema

Each attribute in the `attributes` map must have:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Display name (`minLength: 1`) |
| `data_type_rules` | string | No | Type and constraints |
| `details` | string | No | Description |
| `derivation_policy` | string | No | How value is computed |
| `nullable` | boolean | No | Can be null? |
| `uml_comment` | string | No | UML annotation |

### Index Schema

Each index is an array of attribute keys with `minItems: 1`.

## Common Schema Violations

### 1. Missing Required Name

```json
// WRONG: Missing 'name'
{
    "details": "A customer order"
}

// CORRECT: Include name
{
    "name": "Order",
    "details": "A customer order"
}
```

### 2. Empty Name

```json
// WRONG: Empty name
{
    "name": ""
}

// CORRECT: Non-empty name
{
    "name": "Order"
}
```

### 3. Wrong Type for Fields

```json
// WRONG: nullable is a string, not boolean
{
    "name": "Order",
    "attributes": {
        "notes": {
            "name": "Notes",
            "nullable": "true"
        }
    }
}

// CORRECT: nullable is boolean
{
    "name": "Order",
    "attributes": {
        "notes": {
            "name": "Notes",
            "nullable": true
        }
    }
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'type' is not in the schema
{
    "name": "Order",
    "type": "aggregate"
}

// CORRECT: Only allowed fields
{
    "name": "Order"
}
```

### 5. Missing Attribute Name

```json
// WRONG: Attribute missing 'name'
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "data_type_rules": "ISO date"
        }
    }
}

// CORRECT: Attribute has name
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date",
            "data_type_rules": "ISO date"
        }
    }
}
```

### 6. Empty Index

```json
// WRONG: Empty index array
{
    "name": "Order",
    "indexes": [[]]
}

// CORRECT: Index has at least one attribute
{
    "name": "Order",
    "indexes": [["order_date"]]
}
```

### 7. Unknown Field in Attribute

```json
// WRONG: 'type' is not allowed in attributes
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date",
            "type": "date"
        }
    }
}

// CORRECT: Use data_type_rules instead
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date",
            "data_type_rules": "ISO 8601 date"
        }
    }
}
```

## Valid Examples

### Minimal Valid File

```json
{
    "name": "Order"
}
```

### Class with Attributes

```json
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date",
            "data_type_rules": "ISO 8601 date"
        },
        "total": {
            "name": "Total Amount",
            "nullable": false
        }
    }
}
```

### Complete Valid File

```json
{
    "name": "Order",
    "details": "Represents a customer order",
    "actor_key": "customer",
    "uml_comment": "<<aggregate root>>",
    "attributes": {
        "order_number": {
            "name": "Order Number",
            "data_type_rules": "Unique alphanumeric",
            "details": "Human-readable identifier",
            "nullable": false,
            "uml_comment": "<<unique>>"
        },
        "total": {
            "name": "Total Amount",
            "derivation_policy": "Sum of line items"
        }
    },
    "indexes": [
        ["order_number"],
        ["order_date", "status"]
    ]
}
```

## Related Errors

- **E5001**: Name field is missing
- **E5002**: Name field is empty
- **E5003**: JSON syntax is invalid
- **E5008**: Attribute name is empty
- **E5009**: Index attribute key is invalid
