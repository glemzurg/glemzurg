# Query Schema Violation (E9004)

The query JSON file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your query file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing (`name`)
- A field has the wrong type
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string)

## File Location

Query files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.queries.json    <-- This file violates the schema
```

## Schema Requirements

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | Display name for the query |

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `details` | string | None | Human-readable summary (NOT for logic) |
| `requires` | string[] | None | Preconditions (logic goes here) |
| `guarantees` | string[] | None | Postconditions (logic goes here) |

**Important**: The `details` field is for human-readable summaries only. Logic (preconditions, postconditions, business rules) must go in `requires` and `guarantees`.

## Common Schema Violations

### 1. Missing Required Name

```json
// WRONG: Missing 'name'
{
    "details": "Returns the order total"
}

// CORRECT
{
    "name": "Get Order Total",
    "details": "Returns the order total"
}
```

### 2. Empty Name

```json
// WRONG: Empty name
{
    "name": ""
}

// CORRECT
{
    "name": "Get Order Total"
}
```

### 3. Wrong Type for Fields

```json
// WRONG: requires should be an array, not a string
{
    "name": "Get Order Total",
    "requires": "Order must exist"
}

// CORRECT
{
    "name": "Get Order Total",
    "requires": ["Order must exist"]
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'type' is not in the schema
{
    "name": "Get Order Total",
    "type": "read"
}

// CORRECT
{
    "name": "Get Order Total"
}
```

## Valid Examples

### Minimal Valid File

```json
{
    "name": "Get Order Total"
}
```

### Query with Details

```json
{
    "name": "Get Order Total",
    "details": "Calculates the sum of all line items including taxes and shipping"
}
```

### Complete Query

```json
{
    "name": "Find Pending Orders",
    "details": "Returns all orders that have not been processed yet",
    "requires": [
        "User must be authenticated",
        "User must have read access to orders"
    ],
    "guarantees": [
        "Returns a list of orders with status 'pending'",
        "Results are sorted by creation date descending"
    ]
}
```

## Understanding Requires and Guarantees

- **requires**: Preconditions that must be true before the query executes
- **guarantees**: Postconditions describing what the query returns

These are design-by-contract concepts that help document the query's behavior.

## Related Errors

- **E9001**: Name field is missing
- **E9002**: Name field is empty
- **E9003**: JSON syntax is invalid
