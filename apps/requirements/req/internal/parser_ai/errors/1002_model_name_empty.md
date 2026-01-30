# Model Name Empty (E1002)

The `model.json` file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The parser found a `name` field in your `model.json`, but its value is either an empty string (`""`) or contains only whitespace characters (spaces, tabs, newlines). The model name must contain at least one visible character.

## File Location

The `model.json` file must exist at the **root** of your model directory:

```
your_model/
├── model.json          <-- This file has an empty "name" value
├── actors/
├── domains/
├── associations/
└── generalizations/
```

## How to Fix

Provide a meaningful, non-empty name for your model:

```json
{
    "name": "Your Model Name",
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
{"name": "Web Store"}                      // Typical name
{"name": "Order Management System"}        // Descriptive name
{"name": "  Trimmed Name  "}               // Leading/trailing spaces OK if content exists
```

## Choosing a Good Model Name

A good model name should:

1. **Be descriptive**: Clearly indicate what the model represents
2. **Be concise**: Avoid overly long names (aim for 2-5 words)
3. **Use title case**: Capitalize major words (e.g., "Order Management System")
4. **Avoid special characters**: Stick to letters, numbers, and spaces

### Good Name Examples

| Name | Use Case |
|------|----------|
| `"E-Commerce Platform"` | Online shopping system |
| `"Hospital Management"` | Healthcare administration |
| `"Fleet Tracker"` | Vehicle tracking system |
| `"Customer Portal"` | Client-facing application |
| `"Inventory Control"` | Warehouse management |

### Names to Avoid

| Bad Name | Why It's Bad |
|----------|--------------|
| `"Model"` | Too generic, doesn't describe purpose |
| `"Test"` | Unclear if this is production or test |
| `"New Model"` | Temporary-sounding, not descriptive |
| `"asdf"` | Meaningless placeholder |

## Troubleshooting Checklist

1. **Check for invisible characters**: Copy-paste issues can introduce zero-width spaces or other invisible characters that look like nothing but aren't truly empty
2. **Check your text editor**: Some editors may display whitespace differently; enable "show whitespace" to verify
3. **Regenerate the file**: If unsure, delete and recreate the `model.json` file manually
4. **Validate with a JSON tool**: Use `jq` or an online JSON validator to inspect the actual value

### Checking with Command Line

```bash
# View the raw bytes to see hidden characters
cat model.json | xxd | head -20

# Pretty-print and inspect with jq
cat model.json | jq '.name | length'  # Should be > 0
```

## Complete Schema

The `model.json` file accepts these fields:

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Must have `minLength: 1` (at least one character) |
| `details` | string | No | No length constraints |

## Related Errors

- **E1001**: Model name field is missing entirely
- **E1003**: Invalid JSON syntax in model.json
- **E1004**: model.json violates other schema rules
