# Use Case Generalization Subclass Key Empty (E13007)

A subclass key in the `subclass_keys` array is empty or contains only whitespace.

## What Went Wrong

The `subclass_keys` array contains an entry that is either an empty string (`""`) or contains only whitespace. Every subclass key must be a valid, non-empty use case reference.

## File Location

Use case generalization files are located in the `use_case_generalizations/` directory:

```
your_model/
├── model.json
└── use_case_generalizations/
    └── order_types.uc_gen.json    <-- This file has an empty subclass key
```

## How to Fix

Ensure every entry in `subclass_keys` is a valid, non-empty use case key:

```json
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order", "process_in_store_order"]
}
```

## Invalid Examples

```json
// WRONG: Empty string in array
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order", ""]
}

// WRONG: Whitespace-only entry
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order", "   "]
}

// WRONG: Multiple empty entries
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["", "process_online_order", ""]
}
```

## Valid Examples

```json
// All entries are valid use case keys
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": [
        "process_online_order",
        "process_in_store_order",
        "process_phone_order"
    ]
}
```

## Common Causes

### 1. Incomplete Editing

You may have started adding a subclass but not finished:

```json
{
    "subclass_keys": ["process_online_order", ""]  // Forgot to fill in second entry
}
```

### 2. Copy-Paste Errors

Extra commas can create empty entries in some editors:

```json
{
    "subclass_keys": ["process_online_order", , "process_in_store_order"]
}
```

### 3. Placeholder Text Not Replaced

Template values not replaced with real use case keys:

```json
{
    "subclass_keys": ["process_online_order", "TODO"]  // "TODO" should be replaced
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
cat order_types.uc_gen.json | jq '.subclass_keys'

# Check for empty strings
cat order_types.uc_gen.json | jq '.subclass_keys | map(select(. == "" or (. | test("^\\s*$"))))'
```

## Understanding Use Case Keys

Each subclass key must reference an existing use case file:

```
subclass_keys: ["process_online_order", "process_in_store_order"]
                     |                        |
                     v                        v
    use_cases/process_online_order.uc.json   use_cases/process_in_store_order.uc.json
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

- **E13004**: Schema violation (general)
- **E13005**: Superclass key is empty
- **E13006**: Subclass keys array is missing or empty
- **E13009**: A subclass not found (reference validation, separate from parsing)
