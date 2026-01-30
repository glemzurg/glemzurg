# Class Has No Attributes (E11022)

Every class must have at least one attribute to describe its data properties.

## What Went Wrong

A class has been defined but its `attributes` map is empty. Classes need attributes to describe what data they hold.

## Context

Attributes are the data properties of a class. They define what information a class stores and the rules for that data.

```json
{
    "name": "Order",
    "details": "Represents a customer order",
    "attributes": {
        "order_number": {              <-- At least one attribute required
            "name": "Order Number",
            "data_type_rules": "string: unique",
            "details": "Unique identifier for the order"
        }
    }
}
```

## How to Fix

### Step 1: Identify Class Data

Think about what information this class needs to store:
- **Identifiers** - ID, code, number
- **Core properties** - Name, title, description
- **Status/State** - Status, phase, stage
- **Timestamps** - Created at, updated at, deleted at
- **Quantities** - Count, amount, total
- **References** - Foreign keys to other entities

### Step 2: Add Attributes

Add attributes to your class's `attributes` map:

```json
{
    "name": "Order",
    "details": "Represents a customer order",
    "attributes": {
        "id": {
            "name": "ID",
            "data_type_rules": "uuid: primary key, auto-generated",
            "details": "Unique identifier"
        },
        "order_number": {
            "name": "Order Number",
            "data_type_rules": "string: unique, format ORD-XXXXXXXX",
            "details": "Human-readable order identifier"
        },
        "status": {
            "name": "Status",
            "data_type_rules": "enum: draft, pending, confirmed, shipped, delivered, cancelled",
            "details": "Current state of the order"
        },
        "total_amount": {
            "name": "Total Amount",
            "data_type_rules": "decimal: precision 2, min 0",
            "details": "Total cost including tax and shipping"
        },
        "created_at": {
            "name": "Created At",
            "data_type_rules": "datetime: auto-set on create",
            "details": "When the order was first created"
        }
    }
}
```

## Attribute Components

### Required Fields

- **name**: Human-readable display name
- **data_type_rules**: Type and validation rules

### Optional Fields

- **details**: Explanation of the attribute's purpose
- **derivation_policy**: How the value is computed
- **nullable**: Whether null values are allowed (default: false)

## Data Type Rules Examples

### Strings
```
"string"
"string: max 255"
"string: min 1, max 100"
"string: pattern ^[A-Z]{3}-[0-9]{6}$"
"string: unique"
```

### Numbers
```
"integer"
"integer: min 0"
"integer: min 1, max 100"
"decimal: precision 2"
"decimal: min 0.00, max 999999.99"
```

### Enumerations
```
"enum: pending, active, completed, cancelled"
"enum: low, medium, high, critical"
```

### Dates and Times
```
"date"
"datetime"
"datetime: auto-set on create"
"datetime: auto-set on update"
```

### Boolean
```
"boolean"
"boolean: default false"
```

### References
```
"reference: User"
"reference: Category, optional"
```

## Common Attribute Patterns

### Identity Attributes
```json
"id": {
    "name": "ID",
    "data_type_rules": "uuid: primary key",
    "details": "Unique identifier"
}
```

### Audit Attributes
```json
"created_at": {
    "name": "Created At",
    "data_type_rules": "datetime: auto-set on create"
},
"updated_at": {
    "name": "Updated At",
    "data_type_rules": "datetime: auto-set on update"
}
```

### Soft Delete
```json
"deleted_at": {
    "name": "Deleted At",
    "data_type_rules": "datetime",
    "nullable": true,
    "details": "Set when record is soft-deleted"
}
```

## Related Errors

- **E5001**: Class name is required
- **E5008**: Attribute name cannot be empty
- **E11023**: Class has no state machine
