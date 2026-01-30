# Class Attribute Name Empty (E5008)

An attribute within the class has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The class has an `attributes` map, and one of the attributes has a `name` field that is either an empty string (`""`) or contains only whitespace characters. Every attribute must have a meaningful name.

## File Location

Class files are located within domain directories:

```
your_model/
├── model.json
└── order_management/           <-- Domain directory
    ├── domain.json
    └── order.class.json        <-- An attribute in this file has an empty name
```

## How to Fix

Provide a non-empty, meaningful name for each attribute:

```json
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date"
        }
    }
}
```

## Understanding the Error Field

The error message includes the path to the problematic attribute:

| Error Field | Meaning |
|-------------|---------|
| `attributes.order_date.name` | The `order_date` attribute's name is empty |
| `attributes.total_amount.name` | The `total_amount` attribute's name is empty |
| `attributes.customer_id.name` | The `customer_id` attribute's name is empty |

## Invalid Examples

```json
// WRONG: Empty attribute name
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": ""
        }
    }
}

// WRONG: Whitespace-only attribute name
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "   "
        }
    }
}

// WRONG: Multiple attributes with empty names
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date"
        },
        "total": {
            "name": ""
        }
    }
}
```

## Valid Examples

```json
// Simple attribute
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date"
        }
    }
}

// Multiple attributes
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date",
            "data_type_rules": "ISO 8601 date"
        },
        "total_amount": {
            "name": "Total Amount",
            "data_type_rules": "Positive decimal"
        }
    }
}
```

## Attribute vs Attribute Key

Don't confuse the attribute **key** with the attribute **name**:

```json
{
    "name": "Order",
    "attributes": {
        "order_date": {          // <-- This is the attribute KEY (snake_case)
            "name": "Order Date"  // <-- This is the attribute NAME (display)
        }
    }
}
```

- **Attribute key**: The JSON object key, used for references (e.g., in indexes)
- **Attribute name**: The display name shown in diagrams and documentation

## Attribute Naming Guidelines

Attribute names should:
- Be descriptive and meaningful
- Use title case for multi-word names
- Match the business terminology

| Key | Good Name | Avoid |
|-----|-----------|-------|
| `order_date` | `Order Date` | `OD`, `Date1` |
| `total_amount` | `Total Amount` | `Tot`, `TA` |
| `customer_id` | `Customer ID` | `CID`, `Cust` |

## Complete Attribute Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `data_type_rules` | string | No | Type and validation rules |
| `details` | string | No | Extended description |
| `derivation_policy` | string | No | How value is computed |
| `nullable` | boolean | No | Can be null? |
| `uml_comment` | string | No | UML diagram annotation |

## Related Errors

- **E5001**: Class name is missing
- **E5002**: Class name is empty
- **E5004**: Schema violation (general)
- **E5009**: Index attribute key is invalid
