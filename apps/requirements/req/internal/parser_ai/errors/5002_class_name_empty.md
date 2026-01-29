# Class Name Empty (E5002)

The class JSON file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The `name` field exists but is either an empty string (`""`) or contains only whitespace characters. Every class must have a meaningful name.

## File Location

Class files are located within domain directories:

```
your_model/
├── model.json
└── order_management/           <-- Domain directory
    ├── domain.json
    └── order.class.json        <-- This file has an empty name
```

## How to Fix

Provide a non-empty, meaningful name for the class:

```json
{
    "name": "Order"
}
```

## Invalid Examples

```json
// WRONG: Empty string
{
    "name": ""
}

// WRONG: Whitespace only
{
    "name": "   "
}

// WRONG: Tab characters only
{
    "name": "\t\t"
}
```

## Valid Examples

```json
// Simple name
{
    "name": "Order"
}

// Multi-word name
{
    "name": "Customer Account"
}

// Name with special characters
{
    "name": "Order (Legacy)"
}
```

## Naming Guidelines

Class names should:
- Be descriptive and meaningful
- Use title case for multi-word names
- Represent a single concept or entity
- Be unique within their domain context

| Good Names | Avoid |
|------------|-------|
| `Order` | `O` |
| `Customer Account` | `Cust Acct` |
| `Payment Transaction` | `PayTxn` |
| `Shipping Address` | `Addr1` |

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

- **E5001**: Name field is missing entirely
- **E5003**: Invalid JSON syntax
- **E5004**: Schema violation (general)
