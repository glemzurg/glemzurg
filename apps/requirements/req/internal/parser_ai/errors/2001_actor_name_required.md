# Actor Name Required (E2001)

The actor JSON file is missing the required `name` field.

## What Went Wrong

The parser found an actor file but it does not contain a `name` property. Every actor must have a name that identifies it in diagrams and documentation.

## File Location

Actor files are located in the `actors/` directory at the model root. The filename (without extension) becomes the actor's key.

```
your_model/
├── model.json
├── actors/
│   └── customer.actor.json    <-- This file is missing the "name" field
├── domains/
└── ...
```

## How to Fix

Add a `name` field to your actor JSON file:

```json
{
    "name": "Customer",
    "type": "human",
    "details": "Optional description of the actor"
}
```

## Complete Schema

The actor JSON file accepts these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Human-readable name for the actor (e.g., "Customer", "Inventory System") |
| `type` | string | **Yes** | The type of actor: "human", "system", or "device" |
| `details` | string | No | Extended description of the actor's role and responsibilities |
| `uml_comment` | string | No | Comment to display in UML diagrams |

## How Actor Keys Work

The actor key is derived from the filename:

| File Path | Actor Key |
|-----------|-----------|
| `actors/customer.actor.json` | `customer` |
| `actors/inventory_system.actor.json` | `inventory_system` |
| `actors/payment_gateway.actor.json` | `payment_gateway` |

Classes reference actors using the `actor_key` field, which must match the actor's filename (without extension).

## Connecting Actors to Classes

When a class has an `actor_key`, it references an actor file:

```
actors/customer.actor.json         <-- Actor file defines "customer"
    │
    └── order_management/order.class.json
            {
                "name": "Order",
                "actor_key": "customer"    <-- References the actor
            }
```

## Troubleshooting Checklist

1. **Check the file exists**: Ensure the actor file is in the `actors/` directory
2. **Check JSON syntax**: The file must be valid JSON (see E2005 for JSON syntax help)
3. **Check field name spelling**: The field must be exactly `"name"` (lowercase, in quotes)
4. **Check the value exists**: Ensure the name has a value, not just the key

## Common Mistakes

```json
// WRONG: Missing name entirely
{
    "type": "human"
}

// WRONG: Typo in field name
{
    "Name": "Customer",
    "type": "human"
}

// WRONG: Using 'title' instead of 'name'
{
    "title": "Customer",
    "type": "human"
}
```

## Valid Examples

```json
// Minimal valid actor
{
    "name": "Customer",
    "type": "human"
}

// Full actor with all fields
{
    "name": "Inventory System",
    "type": "system",
    "details": "External system that manages product stock levels and triggers reorder alerts when inventory drops below threshold.",
    "uml_comment": "Async integration via message queue"
}
```

## Related Errors

- **E2002**: Actor name is present but empty or whitespace
- **E2003**: Actor type field is missing
- **E2005**: Invalid JSON syntax in actor file
- **E2006**: Actor file violates schema rules
