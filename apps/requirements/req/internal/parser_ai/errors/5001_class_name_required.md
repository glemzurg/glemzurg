# Class Name Required (E5001)

The class JSON file is missing the required `name` field.

## What Went Wrong

Every class must have a `name` field that identifies it. The parser found a class file without this required field.

## File Location

Class files are located within domain directories:

```
your_model/
├── model.json
└── order_management/           <-- Domain directory
    ├── domain.json
    └── order.class.json        <-- This file is missing the name field
```

## How to Fix

Add a `name` field with a descriptive display name for the class:

```json
{
    "name": "Order",
    "details": "Represents a customer order"
}
```

## Invalid Examples

```json
// WRONG: Missing name field entirely
{
    "details": "Some description"
}

// WRONG: name is null
{
    "name": null,
    "details": "Some description"
}
```

## Valid Examples

```json
// Minimal valid class
{
    "name": "Order"
}

// Class with details
{
    "name": "Order",
    "details": "Represents a customer order in the system"
}

// Full class
{
    "name": "Order",
    "details": "Represents a customer order",
    "actor_key": "customer",
    "attributes": {
        "order_date": {
            "name": "Order Date"
        }
    }
}
```

## Understanding Class Keys

The class key is derived from the file path, not the name field:

```
order_management/order.class.json  -->  Class key: order_management.order
billing/invoice.class.json         -->  Class key: billing.invoice
```

The `name` field is the human-readable display name shown in diagrams.

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `details` | string | No | None |
| `actor_key` | string | No | Must reference existing actor |
| `uml_comment` | string | No | None |
| `attributes` | object | No | Map of attribute keys to definitions |
| `indexes` | array | No | Array of attribute key arrays |

## Related Errors

- **E5002**: Name is empty or whitespace-only
- **E5003**: Invalid JSON syntax
- **E5004**: Schema violation (general)
