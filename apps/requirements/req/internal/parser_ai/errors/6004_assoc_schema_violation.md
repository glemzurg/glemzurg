# Association Schema Violation (E6004)

The association JSON file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your association file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing
- A field has the wrong type
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string)

## File Location

Association files are located in the `associations/` directory at the model root:

```
your_model/
├── model.json
├── associations/
│   └── order_has_items.assoc.json    <-- This file violates the schema
└── order_management/
    └── order.class.json
```

## Schema Requirements

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | Display name for the association |
| `from_class_key` | string | `minLength: 1` | Key referencing the source class |
| `from_multiplicity` | string | `minLength: 1` | Multiplicity at source end |
| `to_class_key` | string | `minLength: 1` | Key referencing the target class |
| `to_multiplicity` | string | `minLength: 1` | Multiplicity at target end |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `details` | string | Extended description |
| `association_class_key` | string or null | Key referencing an association class |
| `uml_comment` | string | UML diagram annotation |

## Common Schema Violations

### 1. Missing Required Fields

```json
// WRONG: Missing from_class_key and to_class_key
{
    "name": "Order Contains Items",
    "from_multiplicity": "1",
    "to_multiplicity": "1..*"
}

// CORRECT
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

### 2. Empty Required Fields

```json
// WRONG: Empty name
{
    "name": "",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}

// CORRECT
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

### 3. Additional Properties Not Allowed

```json
// WRONG: 'type' is not in the schema
{
    "name": "Order Contains Items",
    "type": "composition",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}

// CORRECT
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

## Valid Examples

### Minimal Valid File

```json
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

### Complete Association

```json
{
    "name": "Student Enrolled In Course",
    "details": "Represents student enrollment in courses with grades and dates",
    "from_class_key": "academic.student",
    "from_multiplicity": "*",
    "to_class_key": "academic.course",
    "to_multiplicity": "*",
    "association_class_key": "academic.enrollment",
    "uml_comment": "Many-to-many with enrollment details"
}
```

## Understanding Multiplicity

Standard UML notation:
- `1` - exactly one
- `0..1` - zero or one
- `*` - zero or more
- `1..*` - one or more
- `2..5` - specific range

## Related Errors

- **E6001**: Name field is missing
- **E6002**: Name field is empty
- **E6003**: JSON syntax is invalid
