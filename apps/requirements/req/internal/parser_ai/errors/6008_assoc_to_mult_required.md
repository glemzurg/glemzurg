# Association To Multiplicity Required (E6008)

The association JSON file has a `to_multiplicity` field that is empty or contains only whitespace.

## What Went Wrong

The `to_multiplicity` field exists but is either an empty string (`""`) or contains only whitespace characters. Every association must specify the multiplicity at the target end.

## File Location

Association files are located in the `associations/` directory at the model root:

```
your_model/
├── model.json
├── associations/
│   └── order_has_items.assoc.json    <-- This file has an empty to_multiplicity
└── order_management/
    └── order.class.json
```

## How to Fix

Provide a valid multiplicity value for the `to_multiplicity` field:

```json
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

## Understanding Multiplicity

Multiplicity indicates how many instances of a class can participate in the relationship. Standard UML notation:

| Notation | Meaning |
|----------|---------|
| `1` | Exactly one |
| `0..1` | Zero or one (optional) |
| `*` | Zero or more |
| `1..*` | One or more |
| `2..5` | Specific range (2 to 5) |

## Invalid Examples

```json
// WRONG: Empty to_multiplicity
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": ""
}

// WRONG: Whitespace only
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "   "
}
```

## Valid Examples

```json
// One-to-many: One order has many items
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}

// One-to-one optional: Order has optional shipping
{
    "name": "Order Has Shipping",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "shipping.shipment",
    "to_multiplicity": "0..1"
}
```

## Related Errors

- **E6007**: from_multiplicity is empty
- **E6012**: Multiplicity format is invalid
- **E6004**: Schema violation (general)
