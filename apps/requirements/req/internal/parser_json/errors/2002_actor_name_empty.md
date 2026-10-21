# Actor Name Empty (E2002)

The actor JSON file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The parser found a `name` field in your actor file, but its value is either an empty string (`""`) or contains only whitespace characters (spaces, tabs, newlines). The actor name must contain at least one visible character.

## File Location

Actor files are located in the `actors/` directory at the model root:

```
your_model/
├── model.json
├── actors/
│   └── customer.actor.json    <-- This file has an empty "name" value
├── domains/
└── ...
```

## How to Fix

Provide a meaningful, non-empty name for your actor:

```json
{
    "name": "Customer",
    "type": "human"
}
```

## Invalid Examples

These values will all trigger this error:

```json
{"name": "", "type": "human"}           // Empty string
{"name": "   ", "type": "human"}        // Spaces only
{"name": "\t", "type": "human"}         // Tab only
{"name": "\n", "type": "human"}         // Newline only
```

## Valid Examples

The name must contain at least one non-whitespace character:

```json
{"name": "A", "type": "human"}                    // Single character is valid
{"name": "Customer", "type": "human"}             // Typical name
{"name": "Inventory System", "type": "system"}   // Multi-word name
```

## Choosing Good Actor Names

### Human Actors

Name human actors by their role in the system:

| Good Name | Description |
|-----------|-------------|
| `"Customer"` | End user who purchases products |
| `"Administrator"` | User who manages the system |
| `"Warehouse Worker"` | Staff who handles physical inventory |
| `"Support Agent"` | Staff who handles customer inquiries |

### System Actors

Name system actors by what they do:

| Good Name | Description |
|-----------|-------------|
| `"Payment Gateway"` | External payment processing service |
| `"Inventory System"` | External stock management system |
| `"Email Service"` | Service that sends notifications |
| `"Shipping Provider"` | External logistics integration |

### Device Actors

Name device actors by their function:

| Good Name | Description |
|-----------|-------------|
| `"Barcode Scanner"` | Hardware for scanning products |
| `"POS Terminal"` | Point-of-sale hardware |
| `"IoT Sensor"` | Environmental monitoring device |

## Troubleshooting Checklist

1. **Check for invisible characters**: Copy-paste issues can introduce zero-width spaces
2. **Check your text editor**: Enable "show whitespace" to verify content
3. **Regenerate the file**: If unsure, delete and recreate the file manually

### Checking with Command Line

```bash
# View the raw bytes to see hidden characters
cat actors/customer.actor.json | xxd | head -20

# Pretty-print and inspect with jq
cat actors/customer.actor.json | jq '.name | length'  # Should be > 0
```

## Related Errors

- **E2001**: Actor name field is missing entirely
- **E2003**: Actor type field is missing
- **E2005**: Invalid JSON syntax in actor file
- **E2006**: Actor file violates schema rules
