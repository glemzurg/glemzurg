# Model Name Required (E1001)

The `model.json` file is missing the required `name` field.

## What Went Wrong

The parser found a `model.json` file but it does not contain a `name` property. Every model must have a name that identifies it throughout the system.

## File Location

The `model.json` file must exist at the **root** of your model directory. This is the entry point for the entire model and defines top-level information.

```
your_model/
├── model.json          <-- This file is missing the "name" field
├── actors/
├── domains/
├── associations/
└── generalizations/
```

## How to Fix

Add a `name` field to your `model.json` file:

```json
{
    "name": "Your Model Name",
    "details": "Optional description of what this model represents"
}
```

## Complete Schema

The `model.json` file accepts these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Human-readable name for the model (e.g., "Order Management System") |
| `details` | string | No | Extended description of the model's purpose and scope |

## Why This Field is Required

The model name is the primary identifier used throughout the system:

- **Generated documentation**: The name appears as the title in all generated docs
- **UML diagrams**: Package diagrams show the model name as the root element
- **Error messages**: Errors reference the model name for context
- **Cross-references**: Other tools and systems identify this model by name

## Troubleshooting Checklist

1. **Check the file exists**: Ensure `model.json` is at the root of your model directory
2. **Check JSON syntax**: The file must be valid JSON (see E1003 for JSON syntax help)
3. **Check field name spelling**: The field must be exactly `"name"` (lowercase, in quotes)
4. **Check the value exists**: Ensure the name has a value, not just the key

## Common Mistakes

```json
// WRONG: Missing name entirely
{
    "details": "My model description"
}

// WRONG: Typo in field name
{
    "Name": "My Model"
}

// WRONG: Using 'title' instead of 'name'
{
    "title": "My Model"
}
```

## Valid Examples

```json
// Minimal valid model.json
{
    "name": "Web Store"
}

// Full model.json with all fields
{
    "name": "Order Management System",
    "details": "Handles customer orders, inventory tracking, and fulfillment for the e-commerce platform. Integrates with payment processing and shipping providers."
}
```

## Related Errors

- **E1002**: Model name is present but empty
- **E1003**: Invalid JSON syntax in model.json
- **E1004**: model.json violates schema rules
