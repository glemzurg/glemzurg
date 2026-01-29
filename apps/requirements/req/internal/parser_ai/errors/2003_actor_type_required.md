# Actor Type Required (E2003)

The actor JSON file is missing the required `type` field.

## What Went Wrong

The parser found an actor file but it does not contain a `type` property. Every actor must have a type that categorizes what kind of entity it represents.

## File Location

Actor files are located in the `actors/` directory at the model root:

```
your_model/
├── model.json
├── actors/
│   └── customer.actor.json    <-- This file is missing the "type" field
├── domains/
└── ...
```

## How to Fix

Add a `type` field to your actor JSON file:

```json
{
    "name": "Customer",
    "type": "human"
}
```

## Actor Types

The `type` field categorizes the actor. Common values include:

| Type | Description | Examples |
|------|-------------|----------|
| `"human"` | A person who interacts with the system | Customer, Administrator, Support Agent |
| `"system"` | An external software system | Payment Gateway, Email Service, CRM System |
| `"device"` | Hardware that interacts with the system | Barcode Scanner, IoT Sensor, POS Terminal |

## Complete Schema

The actor JSON file accepts these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Human-readable name for the actor |
| `type` | string | **Yes** | The category of actor: "human", "system", or "device" |
| `details` | string | No | Extended description of the actor's role |
| `uml_comment` | string | No | Comment to display in UML diagrams |

## Why Type is Required

The actor type is used for:

1. **UML Diagrams**: Different actor types are rendered with different icons
2. **Documentation**: Reports group actors by type
3. **Analysis**: Helps identify system boundaries and integration points
4. **Code Generation**: May influence how interactions are implemented

## Troubleshooting Checklist

1. **Check field name spelling**: The field must be exactly `"type"` (lowercase, in quotes)
2. **Check the value exists**: Ensure the type has a value, not just the key
3. **Check JSON syntax**: Make sure there's a comma after the previous field

## Common Mistakes

```json
// WRONG: Missing type entirely
{
    "name": "Customer"
}

// WRONG: Typo in field name
{
    "name": "Customer",
    "Type": "human"
}

// WRONG: Using 'kind' instead of 'type'
{
    "name": "Customer",
    "kind": "human"
}

// WRONG: Using 'category' instead of 'type'
{
    "name": "Customer",
    "category": "human"
}
```

## Valid Examples

```json
// Human actor
{
    "name": "Customer",
    "type": "human"
}

// System actor
{
    "name": "Payment Gateway",
    "type": "system",
    "details": "Third-party payment processor for credit card transactions"
}

// Device actor
{
    "name": "Barcode Scanner",
    "type": "device",
    "details": "Handheld scanner for reading product barcodes"
}
```

## Related Errors

- **E2001**: Actor name field is missing
- **E2004**: Actor type is present but invalid (empty or whitespace)
- **E2005**: Invalid JSON syntax in actor file
- **E2006**: Actor file violates schema rules
