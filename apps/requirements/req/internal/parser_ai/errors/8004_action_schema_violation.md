# Action Schema Violation (E8004)

The action JSON file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your action file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing (`name`)
- A field has the wrong type
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string)

## File Location

Action files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── order.class.json
    └── order.actions.json    <-- This file violates the schema
```

## Schema Requirements

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | Display name for the action |

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `details` | string | None | Extended description |
| `requires` | string[] | None | Preconditions |
| `guarantees` | string[] | None | Postconditions |

## Common Schema Violations

### 1. Missing Required Name

```json
// WRONG: Missing 'name'
{
    "details": "Sends an email"
}

// CORRECT
{
    "name": "Send Email",
    "details": "Sends an email"
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
    "name": "Send Email"
}
```

### 3. Wrong Type for Fields

```json
// WRONG: requires should be an array, not a string
{
    "name": "Send Email",
    "requires": "Customer must exist"
}

// CORRECT
{
    "name": "Send Email",
    "requires": ["Customer must exist"]
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'type' is not in the schema
{
    "name": "Send Email",
    "type": "notification"
}

// CORRECT
{
    "name": "Send Email"
}
```

## Valid Examples

### Minimal Valid File

```json
{
    "name": "Send Confirmation Email"
}
```

### Action with Details

```json
{
    "name": "Send Confirmation Email",
    "details": "Sends an email to the customer confirming their order has been received"
}
```

### Complete Action

```json
{
    "name": "Process Payment",
    "details": "Charges the customer's payment method and records the transaction",
    "requires": [
        "Order total must be greater than zero",
        "Customer payment method must be valid"
    ],
    "guarantees": [
        "Payment has been charged",
        "Transaction record has been created"
    ]
}
```

## Understanding Requires and Guarantees

- **requires**: Preconditions that must be true before the action executes
- **guarantees**: Postconditions that will be true after the action completes

These are design-by-contract concepts that help document the action's behavior.

## Related Errors

- **E8001**: Name field is missing
- **E8002**: Name field is empty
- **E8003**: JSON syntax is invalid
