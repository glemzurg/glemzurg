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
            "data_type_rules": "unconstrained",
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
        "order_number": {
            "name": "Order Number",
            "data_type_rules": "unconstrained",
            "details": "Human-readable order identifier"
        },
        "status": {
            "name": "Status",
            "data_type_rules": "enum of draft, pending, confirmed, shipped, delivered, cancelled",
            "details": "Current state of the order"
        },
        "total_amount": {
            "name": "Total Amount",
            "data_type_rules": "[0 .. unconstrained] at 0.01 dollars",
            "details": "Total cost including tax and shipping"
        }
    }
}
```

## Attribute Components

### Required Fields

- **name**: Human-readable display name
- **data_type_rules**: Type and validation rules (see **E5011** for full syntax)

### Optional Fields

- **details**: Explanation of the attribute's purpose
- **derivation_policy**: How the value is computed
- **nullable**: Whether null values are allowed (default: false)

## Data Type Rules

For the complete data type syntax including spans, enums, collections, and records, see **E5011** (Attribute Data Type Unparseable).

## Related Errors

- **E5001**: Class name is required
- **E5008**: Attribute name cannot be empty
- **E11023**: Class has no state machine
