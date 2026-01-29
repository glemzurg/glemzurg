# Domain Name Empty (E3002)

The `domain.json` file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The parser found a `name` field in your `domain.json`, but its value is either an empty string (`""`) or contains only whitespace characters. The domain name must contain at least one visible character.

## File Location

Domain files are located in directories named after the domain:

```
your_model/
├── model.json
└── order_management/
    ├── domain.json             <-- This file has an empty "name" value
    └── ... (classes, etc.)
```

## How to Fix

Provide a meaningful, non-empty name for your domain:

```json
{
    "name": "Order Management",
    "details": "Optional description"
}
```

## Invalid Examples

These values will all trigger this error:

```json
{"name": ""}              // Empty string
{"name": "   "}           // Spaces only
{"name": "\t"}            // Tab only
{"name": "\n"}            // Newline only
{"name": "  \t\n  "}      // Mixed whitespace only
```

## Valid Examples

The name must contain at least one non-whitespace character:

```json
{"name": "A"}                              // Single character is valid
{"name": "Order Management"}               // Typical name
{"name": "User Authentication"}            // Descriptive name
{"name": "  Trimmed Name  "}               // Leading/trailing spaces OK if content exists
```

## Choosing a Good Domain Name

A good domain name should:

1. **Be descriptive**: Clearly indicate what area of functionality the domain covers
2. **Be concise**: Avoid overly long names (aim for 2-4 words)
3. **Use title case**: Capitalize major words (e.g., "Order Management")
4. **Reflect business concepts**: Use terminology stakeholders understand

### Good Domain Name Examples

| Name | Use Case |
|------|----------|
| `"Order Management"` | Handling customer orders |
| `"User Authentication"` | Login and identity management |
| `"Inventory Control"` | Stock and warehouse management |
| `"Payment Processing"` | Financial transactions |
| `"Shipping & Fulfillment"` | Delivery logistics |

### Names to Avoid

| Bad Name | Why It's Bad |
|----------|--------------|
| `"Domain"` | Too generic, doesn't describe purpose |
| `"Stuff"` | Meaningless, unclear scope |
| `"Module1"` | Technical naming, not business-focused |
| `"Misc"` | Suggests poor organization |

## Troubleshooting Checklist

1. **Check for invisible characters**: Copy-paste issues can introduce hidden characters
2. **Check your text editor**: Enable "show whitespace" to verify content
3. **Regenerate the file**: If unsure, delete and recreate the file manually
4. **Validate with a JSON tool**: Use `jq` or an online validator to inspect the value

### Checking with Command Line

```bash
# View the raw bytes to see hidden characters
cat domain.json | xxd | head -20

# Pretty-print and inspect with jq
cat domain.json | jq '.name | length'  # Should be > 0
```

## Complete Schema

The `domain.json` file accepts these fields:

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Must have `minLength: 1` (at least one character) |
| `details` | string | No | No length constraints |
| `realized` | boolean | No | Default: false |
| `uml_comment` | string | No | No constraints |

## Related Errors

- **E3001**: Domain name field is missing entirely
- **E3003**: Invalid JSON syntax in domain.json
- **E3004**: domain.json violates other schema rules
