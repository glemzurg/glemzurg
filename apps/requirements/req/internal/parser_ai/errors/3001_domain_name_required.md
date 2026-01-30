# Domain Name Required (E3001)

The `domain.json` file is missing the required `name` field.

## What Went Wrong

The parser found a `domain.json` file but it does not contain a `name` property. Every domain must have a name that identifies it throughout the system.

## File Location

Domain files are located in directories named after the domain. The `domain.json` file defines the domain's metadata:

```
your_model/
├── model.json
├── actors/
└── order_management/           <-- Domain directory (name becomes the domain key)
    ├── domain.json             <-- This file is missing the "name" field
    └── ... (classes, etc.)
```

## How to Fix

Add a `name` field to your `domain.json` file:

```json
{
    "name": "Order Management",
    "details": "Optional description of what this domain covers"
}
```

## Complete Schema

The `domain.json` file accepts these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Human-readable name for the domain (e.g., "Order Management") |
| `details` | string | No | Extended description of the domain's purpose and scope |
| `realized` | boolean | No | Whether this domain is realized/implemented (default: false) |
| `uml_comment` | string | No | Comment for UML diagram annotations |

## Why This Field is Required

The domain name is used throughout the system:

- **Generated documentation**: The name appears in domain documentation
- **UML diagrams**: Package diagrams show the domain name
- **Error messages**: Errors reference the domain name for context
- **Navigation**: Helps readers understand the model structure

## Domain Key vs Domain Name

The domain has two identifiers:

1. **Domain Key**: Derived from the directory name (e.g., `order_management`)
   - Used in code and cross-references
   - Must be lowercase with underscores
   - Appears in class keys like `order_management.order`

2. **Domain Name**: The `name` field in `domain.json` (e.g., "Order Management")
   - Used for display purposes
   - Can contain spaces and mixed case
   - Shown in diagrams and documentation

## Troubleshooting Checklist

1. **Check the file exists**: Ensure `domain.json` is in the domain directory
2. **Check JSON syntax**: The file must be valid JSON (see E3003 for JSON syntax help)
3. **Check field name spelling**: The field must be exactly `"name"` (lowercase, in quotes)
4. **Check the value exists**: Ensure the name has a value, not just the key

## Common Mistakes

```json
// WRONG: Missing name entirely
{
    "details": "Handles customer orders"
}

// WRONG: Typo in field name
{
    "Name": "Order Management"
}

// WRONG: Using 'title' instead of 'name'
{
    "title": "Order Management"
}
```

## Valid Examples

```json
// Minimal valid domain.json
{
    "name": "Order Management"
}

// Full domain.json with all fields
{
    "name": "Order Management",
    "details": "Handles the complete lifecycle of customer orders from placement through fulfillment and delivery.",
    "realized": false,
    "uml_comment": "Core business domain"
}
```

## How Domains Connect to Other Files

```
order_management/                    <-- Domain directory
├── domain.json                      <-- You are here (defines domain name)
│
├── order.class.json                 <-- Classes in this domain
├── order_line.class.json            │   Key: order_management.order
│                                    │   Key: order_management.order_line
│
├── order.state_machine.json         <-- State machines for classes
├── order.actions.json               <-- Actions for classes
└── order.queries.json               <-- Queries for classes

associations/
└── *.assoc.json                     <-- May reference classes in this domain
                                         using keys like "order_management.order"
```

## Related Errors

- **E3002**: Domain name is present but empty
- **E3003**: Invalid JSON syntax in domain.json
- **E3004**: domain.json violates schema rules
