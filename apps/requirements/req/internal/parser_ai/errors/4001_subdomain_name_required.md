# Subdomain Name Required (E4001)

The `subdomain.json` file is missing the required `name` field.

## What Went Wrong

The parser found a `subdomain.json` file but it does not contain a `name` property. Every subdomain must have a name that identifies it throughout the system.

## File Location

Subdomain files are located within domain directories, providing finer-grained organization:

```
your_model/
├── model.json
└── order_management/                 <-- Domain directory
    ├── domain.json
    └── fulfillment/                  <-- Subdomain directory
        ├── subdomain.json            <-- This file is missing the "name" field
        └── ... (classes, etc.)
```

## How to Fix

Add a `name` field to your `subdomain.json` file:

```json
{
    "name": "Order Fulfillment",
    "details": "Optional description of what this subdomain covers"
}
```

## Complete Schema

The `subdomain.json` file accepts these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Human-readable name for the subdomain (e.g., "Order Fulfillment") |
| `details` | string | No | Extended description of the subdomain's purpose and scope |
| `uml_comment` | string | No | Comment for UML diagram annotations |

## Why This Field is Required

The subdomain name is used throughout the system:

- **Generated documentation**: The name appears in subdomain documentation
- **UML diagrams**: Package diagrams show the subdomain name as a nested package
- **Error messages**: Errors reference the subdomain name for context
- **Navigation**: Helps readers understand the model's organizational hierarchy

## Subdomain Key vs Subdomain Name

The subdomain has two identifiers:

1. **Subdomain Key**: Derived from the directory name (e.g., `fulfillment`)
   - Used in code and cross-references
   - Must be lowercase with underscores
   - Combined with domain to form class keys like `order_management.fulfillment.shipment`

2. **Subdomain Name**: The `name` field in `subdomain.json` (e.g., "Order Fulfillment")
   - Used for display purposes
   - Can contain spaces and mixed case
   - Shown in diagrams and documentation

## Troubleshooting Checklist

1. **Check the file exists**: Ensure `subdomain.json` is in the subdomain directory
2. **Check JSON syntax**: The file must be valid JSON (see E4003 for JSON syntax help)
3. **Check field name spelling**: The field must be exactly `"name"` (lowercase, in quotes)
4. **Check the value exists**: Ensure the name has a value, not just the key

## Common Mistakes

```json
// WRONG: Missing name entirely
{
    "details": "Handles order fulfillment"
}

// WRONG: Typo in field name
{
    "Name": "Order Fulfillment"
}

// WRONG: Using 'title' instead of 'name'
{
    "title": "Order Fulfillment"
}
```

## Valid Examples

```json
// Minimal valid subdomain.json
{
    "name": "Order Fulfillment"
}

// Full subdomain.json with all fields
{
    "name": "Order Fulfillment",
    "details": "Handles the picking, packing, and shipping of customer orders. Coordinates with warehouse systems and shipping carriers.",
    "uml_comment": "Integrates with WMS"
}
```

## How Subdomains Connect to Other Files

```
order_management/                         <-- Domain (key: order_management)
├── domain.json
└── fulfillment/                          <-- Subdomain (key: fulfillment)
    ├── subdomain.json                    <-- You are here
    │
    ├── shipment.class.json               <-- Classes in this subdomain
    ├── package.class.json                │   Key: order_management.fulfillment.shipment
    │                                     │   Key: order_management.fulfillment.package
    │
    └── shipment.state_machine.json       <-- State machines for classes

associations/
└── *.assoc.json                          <-- May reference subdomain classes
                                              using keys like:
                                              "order_management.fulfillment.shipment"
```

## Related Errors

- **E4002**: Subdomain name is present but empty
- **E4003**: Invalid JSON syntax in subdomain.json
- **E4004**: subdomain.json violates schema rules
