# Association Name Required (E6001)

The association JSON file is missing the required `name` field.

## What Went Wrong

Every association must have a `name` field that identifies it. The parser found an association file without this required field.

## File Location

Association files are located in the `associations/` directory at the model root:

```
your_model/
├── model.json
├── associations/
│   └── order_has_items.assoc.json    <-- This file is missing the name field
└── order_management/
    └── order.class.json
```

## How to Fix

Add a `name` field with a descriptive verb phrase for the association:

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
// WRONG: Missing name field entirely
{
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}

// WRONG: name is null
{
    "name": null,
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

## Valid Examples

```json
// Minimal valid association
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
- Describe the connection between the two classes

| Good Names | Avoid |
|------------|-------|
| `Order Contains Items` | `OrderItem`, `Assoc1` |
| `Customer Places Order` | `CustomerOrder` |
| `Employee Works In Department` | `EmpDept` |

## Complete Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Display name for the association |
| `from_class_key` | string | **Yes** | Key referencing the source class |
| `from_multiplicity` | string | **Yes** | Multiplicity at source end |
| `to_class_key` | string | **Yes** | Key referencing the target class |
| `to_multiplicity` | string | **Yes** | Multiplicity at target end |
| `details` | string | No | Extended description |
| `association_class_key` | string | No | Key referencing an association class |
| `uml_comment` | string | No | UML diagram annotation |

## Related Errors

- **E6002**: Name is empty or whitespace-only
- **E6003**: Invalid JSON syntax
- **E6004**: Schema violation (general)
