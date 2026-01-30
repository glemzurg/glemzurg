# Generalization Subclass Key Empty (E10007)

A subclass key in the `subclass_keys` array is empty or contains only whitespace.

## What Went Wrong

The `subclass_keys` array contains an entry that is either an empty string (`""`) or contains only whitespace. Every subclass key must be a valid, non-empty class reference.

## File Location

Generalization files are located in the `generalizations/` directory:

```
your_model/
├── model.json
└── generalizations/
    └── payment_types.gen.json    <-- This file has an empty subclass key
```

## How to Fix

Ensure every entry in `subclass_keys` is a valid, non-empty class key:

```json
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card", "billing.bank_transfer"]
}
```

## Invalid Examples

```json
// WRONG: Empty string in array
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card", ""]
}

// WRONG: Whitespace-only entry
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card", "   "]
}

// WRONG: Multiple empty entries
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["", "billing.credit_card", ""]
}
```

## Valid Examples

```json
// All entries are valid class keys
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": [
        "billing.credit_card",
        "billing.bank_transfer",
        "billing.paypal"
    ]
}
```

## Common Causes

### 1. Incomplete Editing

You may have started adding a subclass but not finished:

```json
{
    "subclass_keys": ["billing.credit_card", ""]  // Forgot to fill in second entry
}
```

### 2. Copy-Paste Errors

Extra commas can create empty entries in some editors:

```json
{
    "subclass_keys": ["billing.credit_card", , "billing.bank_transfer"]
}
```

### 3. Placeholder Text Not Replaced

Template values not replaced with real class keys:

```json
{
    "subclass_keys": ["billing.credit_card", "TODO"]  // "TODO" should be replaced
}
```

## Understanding the Error Field

The error message includes the array index (0-based) of the problematic entry:

| Error Field | Meaning |
|-------------|---------|
| `subclass_keys[0]` | First entry is empty |
| `subclass_keys[1]` | Second entry is empty |
| `subclass_keys[2]` | Third entry is empty |

## Troubleshooting Checklist

1. **Check for empty strings**: Look for `""` in the array
2. **Check for whitespace-only**: Look for `"   "` entries
3. **Check for extra commas**: Can create implicit empty entries
4. **Remove placeholder text**: Replace "TODO" or similar with real keys
5. **Verify array formatting**: Each entry should be a quoted string

### Checking with Command Line

```bash
# View the subclass_keys array
cat payment_types.gen.json | jq '.subclass_keys'

# Check for empty strings
cat payment_types.gen.json | jq '.subclass_keys | map(select(. == "" or (. | test("^\\s*$"))))'
```

## Understanding Class Keys

Each subclass key must reference an existing class file:

```
subclass_keys: ["billing.credit_card", "billing.bank_transfer"]
                     │                        │
                     ▼                        ▼
        billing/credit_card.class.json   billing/bank_transfer.class.json
```

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `superclass_key` | string | **Yes** | `minLength: 1` |
| `subclass_keys` | string[] | **Yes** | `minItems: 1`, **each** `minLength: 1` |
| `details` | string | No | None |
| `is_complete` | boolean | No | Default: false |
| `is_static` | boolean | No | Default: false |
| `uml_comment` | string | No | None |

## Related Errors

- **E10004**: Schema violation (general)
- **E10005**: Superclass key is empty
- **E10006**: Subclass keys array is missing or empty
- **E10009**: A subclass not found (reference validation, separate from parsing)
