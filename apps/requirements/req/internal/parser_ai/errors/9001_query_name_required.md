# Query Name Required (E9001)

The query JSON file is missing the required `name` field.

## What Went Wrong

Every query must have a `name` field that identifies it. The parser found a query file without this required field.

## File Location

Query files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── domain.json
    ├── order.class.json
    └── order.queries.json    <-- This file is missing the name field
```

## How to Fix

Add a `name` field with a descriptive verb phrase for the query:

```json
{
    "name": "Get Order Total"
}
```

## Invalid Examples

```json
// WRONG: Missing name field entirely
{
    "details": "Returns the order total"
}

// WRONG: name is null
{
    "name": null,
    "details": "Returns the order total"
}
```

## Valid Examples

```json
// Minimal valid query
{
    "name": "Get Order Total"
}

// Query with details
{
    "name": "Get Order Total",
    "details": "Calculates the sum of all line items including taxes"
}

// Full query
{
    "name": "Find Pending Orders",
    "details": "Returns all orders that have not been processed",
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

## Query Naming Guidelines

Query names should:
- Use verb phrases that describe what information is returned
- Be specific and meaningful
- Use title case for multi-word names

| Good Names | Avoid |
|------------|-------|
| `Get Order Total` | `Total`, `Query1` |
| `Find Pending Orders` | `PO`, `Find` |
| `Calculate Shipping Cost` | `Calc`, `Shipping` |

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `details` | string | No | Human-readable summary only |
| `requires` | string[] | No | Preconditions (logic goes here) |
| `guarantees` | string[] | No | Postconditions (logic goes here) |

## Important: Where Logic Belongs

**The `details` field is for human-readable summaries only — NOT for describing logic.**

| DO put in `details` | DON'T put in `details` |
|---------------------|------------------------|
| Brief description | Preconditions |
| What it returns | Postconditions |
| High-level purpose | Filtering rules |
| | Business logic |
| | "If X then Y" statements |

**Wrong** — logic stuffed into details:
```json
{
    "name": "Find Pending Orders",
    "details": "Returns orders where status='pending'. User must be authenticated. Results sorted by date descending. Only returns orders user has access to."
}
```

**Correct** — logic in structured fields:
```json
{
    "name": "Find Pending Orders",
    "details": "Returns all orders that have not been processed",
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

## Related Errors

- **E9002**: Name is empty or whitespace-only
- **E9003**: Invalid JSON syntax
- **E9004**: Schema violation (general)
