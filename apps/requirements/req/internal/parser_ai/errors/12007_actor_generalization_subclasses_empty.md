# Actor Generalization Subclass Key Empty (E12007)

A subclass key in the `subclass_keys` array is empty or contains only whitespace.

## What Went Wrong

The `subclass_keys` array contains an entry that is either an empty string (`""`) or contains only whitespace. Every subclass key must be a valid, non-empty actor reference.

## File Location

Actor generalization files are located in the `actor_generalizations/` directory:

```
your_model/
├── model.json
└── actor_generalizations/
    └── user_types.actor_gen.json    <-- This file has an empty subclass key
```

## How to Fix

Ensure every entry in `subclass_keys` is a valid, non-empty actor key:

```json
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer", "guest"]
}
```

## Invalid Examples

```json
// WRONG: Empty string in array
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer", ""]
}

// WRONG: Whitespace-only entry
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer", "   "]
}

// WRONG: Multiple empty entries
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["", "premium_customer", ""]
}
```

## Valid Examples

```json
// All entries are valid actor keys
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": [
        "premium_customer",
        "guest",
        "corporate_customer"
    ]
}
```

## Common Causes

### 1. Incomplete Editing

You may have started adding a subclass but not finished:

```json
{
    "subclass_keys": ["premium_customer", ""]  // Forgot to fill in second entry
}
```

### 2. Copy-Paste Errors

Extra commas can create empty entries in some editors:

```json
{
    "subclass_keys": ["premium_customer", , "guest"]
}
```

### 3. Placeholder Text Not Replaced

Template values not replaced with real actor keys:

```json
{
    "subclass_keys": ["premium_customer", "TODO"]  // "TODO" should be replaced
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
cat user_types.actor_gen.json | jq '.subclass_keys'

# Check for empty strings
cat user_types.actor_gen.json | jq '.subclass_keys | map(select(. == "" or (. | test("^\\s*$"))))'
```

## Understanding Actor Keys

Each subclass key must reference an existing actor file:

```
subclass_keys: ["premium_customer", "guest"]
                     |                  |
                     v                  v
     actors/premium_customer.actor.json   actors/guest.actor.json
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

- **E12004**: Schema violation (general)
- **E12005**: Superclass key is empty
- **E12006**: Subclass keys array is missing or empty
- **E12009**: A subclass not found (reference validation, separate from parsing)
