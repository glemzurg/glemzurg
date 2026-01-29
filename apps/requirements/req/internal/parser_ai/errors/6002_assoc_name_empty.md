# Association Name Empty (E6002)

The association JSON file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The `name` field exists but is either an empty string (`""`) or contains only whitespace characters. Every association must have a meaningful name.

## File Location

Association files are located in the `associations/` directory at the model root:

```
your_model/
├── model.json
├── associations/
│   └── order_has_items.assoc.json    <-- This file has an empty name
└── order_management/
    └── order.class.json
```

## How to Fix

Provide a non-empty, meaningful name for the association:

```json
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

## Invalid Examples

```json
// WRONG: Empty string
{
    "name": "",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}

// WRONG: Whitespace only
{
    "name": "   ",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

## Valid Examples

```json
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

## Association Naming Guidelines

Association names should:
- Use verb phrases that describe the relationship
- Be specific and meaningful
- Use title case for multi-word names

| Good Names | Avoid |
|------------|-------|
| `Order Contains Items` | `OrderItem` |
| `Customer Places Order` | `CO` |
| `Employee Works In Department` | `Assoc` |

## Related Errors

- **E6001**: Name field is missing entirely
- **E6003**: Invalid JSON syntax
- **E6004**: Schema violation (general)
