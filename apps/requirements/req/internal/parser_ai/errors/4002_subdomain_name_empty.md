# Subdomain Name Empty (E4002)

The `subdomain.json` file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The parser found a `name` field in your `subdomain.json`, but its value is either an empty string (`""`) or contains only whitespace characters. The subdomain name must contain at least one visible character.

## File Location

Subdomain files are located within domain directories:

```
your_model/
├── model.json
└── order_management/
    ├── domain.json
    └── fulfillment/
        ├── subdomain.json            <-- This file has an empty "name" value
        └── ... (classes, etc.)
```

## How to Fix

Provide a meaningful, non-empty name for your subdomain:

```json
{
    "name": "Order Fulfillment",
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
{"name": "Order Fulfillment"}              // Typical name
{"name": "Payment Validation"}             // Descriptive name
{"name": "  Trimmed Name  "}               // Leading/trailing spaces OK if content exists
```

## Choosing a Good Subdomain Name

A good subdomain name should:

1. **Be descriptive**: Clearly indicate what area of the domain it covers
2. **Be concise**: Avoid overly long names (aim for 2-4 words)
3. **Use title case**: Capitalize major words (e.g., "Order Fulfillment")
4. **Relate to parent domain**: The name should make sense in context of the domain

### Good Subdomain Name Examples

| Domain | Subdomain | Description |
|--------|-----------|-------------|
| Order Management | `"Order Processing"` | Handling new orders |
| Order Management | `"Order Fulfillment"` | Picking, packing, shipping |
| Payment | `"Payment Validation"` | Verifying payment methods |
| Payment | `"Transaction Processing"` | Executing payments |
| User Management | `"Authentication"` | Login and sessions |
| User Management | `"Authorization"` | Permissions and roles |

### Names to Avoid

| Bad Name | Why It's Bad |
|----------|--------------|
| `"Subdomain"` | Too generic, doesn't describe purpose |
| `"Default"` | Suggests laziness, unclear scope |
| `"Misc"` | Suggests poor organization |
| `"Module1"` | Technical naming, not business-focused |

## Troubleshooting Checklist

1. **Check for invisible characters**: Copy-paste issues can introduce hidden characters
2. **Check your text editor**: Enable "show whitespace" to verify content
3. **Regenerate the file**: If unsure, delete and recreate the file manually
4. **Validate with a JSON tool**: Use `jq` or an online validator to inspect the value

### Checking with Command Line

```bash
# View the raw bytes to see hidden characters
cat subdomain.json | xxd | head -20

# Pretty-print and inspect with jq
cat subdomain.json | jq '.name | length'  # Should be > 0
```

## Complete Schema

The `subdomain.json` file accepts these fields:

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Must have `minLength: 1` (at least one character) |
| `details` | string | No | No length constraints |
| `uml_comment` | string | No | No constraints |

## Related Errors

- **E4001**: Subdomain name field is missing entirely
- **E4003**: Invalid JSON syntax in subdomain.json
- **E4004**: subdomain.json violates other schema rules
