# Actor Type Invalid (E2004)

The actor JSON file has a `type` field that is empty or contains only whitespace.

## What Went Wrong

The parser found a `type` field in your actor file, but its value is either an empty string (`""`) or contains only whitespace characters. The actor type must contain at least one visible character.

## File Location

Actor files are located in the `actors/` directory at the model root:

```
your_model/
├── model.json
├── actors/
│   └── customer.actor.json    <-- This file has an empty "type" value
├── domains/
└── ...
```

## How to Fix

Provide a valid type for your actor:

```json
{
    "name": "Customer",
    "type": "human"
}
```

## Invalid Examples

These values will all trigger this error:

```json
{"name": "Customer", "type": ""}           // Empty string
{"name": "Customer", "type": "   "}        // Spaces only
{"name": "Customer", "type": "\t"}         // Tab only
```

## Valid Actor Types

Common actor types include:

| Type | Description | When to Use |
|------|-------------|-------------|
| `"human"` | A person | Users, operators, administrators, staff |
| `"system"` | External software | APIs, services, databases, other applications |
| `"device"` | Hardware | Sensors, scanners, terminals, IoT devices |

## Choosing the Right Type

### Use `"human"` for:
- End users (customers, visitors)
- Internal users (employees, managers)
- Support staff (agents, administrators)
- Any person who interacts with the system

### Use `"system"` for:
- Payment processors (Stripe, PayPal)
- Email services (SendGrid, Mailchimp)
- Cloud services (AWS, Azure)
- Partner APIs
- Legacy systems
- Databases that act as sources of truth

### Use `"device"` for:
- Input devices (scanners, readers)
- Output devices (printers, displays)
- IoT sensors (temperature, motion)
- Embedded systems
- Point-of-sale terminals

## Custom Types

While `"human"`, `"system"`, and `"device"` are the most common types, you can use other meaningful values if they better describe your domain:

```json
{"name": "Regulatory Body", "type": "organization"}
{"name": "Partner Company", "type": "external_business"}
{"name": "Scheduled Job", "type": "timer"}
```

Just ensure the type is:
- Non-empty
- Meaningful to readers
- Consistent across your model

## Troubleshooting Checklist

1. **Check for invisible characters**: Copy-paste issues can introduce zero-width spaces
2. **Check your text editor**: Enable "show whitespace" to verify content
3. **Check the value**: Ensure you have a meaningful type string

### Checking with Command Line

```bash
# View the raw bytes to see hidden characters
cat actors/customer.actor.json | xxd | head -20

# Check type field length with jq
cat actors/customer.actor.json | jq '.type | length'  # Should be > 0
```

## Related Errors

- **E2001**: Actor name field is missing
- **E2002**: Actor name is empty or whitespace
- **E2003**: Actor type field is missing entirely
- **E2005**: Invalid JSON syntax in actor file
- **E2006**: Actor file violates schema rules
