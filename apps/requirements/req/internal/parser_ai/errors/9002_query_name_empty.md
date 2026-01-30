# Query Name Empty (E9002)

The query JSON file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The `name` field exists but is either an empty string (`""`) or contains only whitespace characters. Every query must have a meaningful name.

## File Location

Query files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.queries.json    <-- This file has an empty name
```

## How to Fix

Provide a non-empty, meaningful name for the query:

```json
{
    "name": "Get Order Total"
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
    "name": "Get Order Total"
}

// Multi-word name
{
    "name": "Find Orders By Customer"
}

// Full query
{
    "name": "Calculate Shipping Cost",
    "details": "Calculates shipping based on weight and destination",
    "requires": ["Order must have items"],
    "guarantees": ["Returns cost in order currency"]
}
```

## Query Naming Guidelines

Query names should:
- Use verb phrases that describe what information is returned
- Be specific and meaningful
- Use title case for multi-word names
- Describe what the query returns, not how it works

| Good Names | Avoid |
|------------|-------|
| `Get Order Total` | `Total` |
| `Find Pending Orders` | `PO` |
| `Calculate Shipping Cost` | `Calc` |
| `List Customer Orders` | `Orders` |

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `details` | string | No | None |
| `requires` | string[] | No | Preconditions |
| `guarantees` | string[] | No | Postconditions |

## Related Errors

- **E9001**: Name field is missing entirely
- **E9003**: Invalid JSON syntax
- **E9004**: Schema violation (general)
