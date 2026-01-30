# Association To Class Required (E6006)

The association JSON file has a `to_class_key` field that is empty or contains only whitespace.

## What Went Wrong

The `to_class_key` field exists but is either an empty string (`""`) or contains only whitespace characters. Every association must specify the target class it connects.

## File Location

Association files are located in the `associations/` directory at the model root:

```
your_model/
├── model.json
├── associations/
│   └── order_has_items.assoc.json    <-- This file has an empty to_class_key
└── order_management/
    └── order.class.json
```

## How to Fix

Provide a valid class key for the `to_class_key` field. The class key is constructed from the domain path plus filename:

```json
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

## Class Key Format

The class key follows the pattern `domain.classname`:
- For `order_management/order_item.class.json` → `order_management.order_item`
- For `inventory/product.class.json` → `inventory.product`
- For nested subdomains, include all path segments separated by dots

## Invalid Examples

```json
// WRONG: Empty to_class_key
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "",
    "to_multiplicity": "1..*"
}

// WRONG: Whitespace only
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "   ",
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

## Related Errors

- **E6005**: from_class_key is empty
- **E6010**: to_class_key references a non-existent class
- **E6004**: Schema violation (general)
