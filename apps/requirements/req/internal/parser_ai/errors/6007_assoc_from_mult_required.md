# Association From Multiplicity Required (E6007)

The association JSON file has a `from_multiplicity` field that is empty or contains only whitespace.

## What Went Wrong

The `from_multiplicity` field exists but is either an empty string (`""`) or contains only whitespace characters. Every association must specify the multiplicity at the source end.

## File Location

Association files are located in the `associations/` directory at the model root:

```
your_model/
├── model.json
├── associations/
│   └── order_has_items.assoc.json    <-- This file has an empty from_multiplicity
└── order_management/
    └── order.class.json
```

## How to Fix

Provide a valid multiplicity value for the `from_multiplicity` field:

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
// WRONG: Empty from_multiplicity
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}

// WRONG: Whitespace only
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "   ",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
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

// Many-to-many: Students enrolled in courses
{
    "name": "Student Enrolled In Course",
    "from_class_key": "academic.student",
    "from_multiplicity": "*",
    "to_class_key": "academic.course",
    "to_multiplicity": "*"
}
```

## Related Errors

- **E6008**: to_multiplicity is empty
- **E6012**: Multiplicity format is invalid
- **E6004**: Schema violation (general)
