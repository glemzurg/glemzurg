# Conversion Key Construction Failed (E21001)

An identity key could not be constructed during model conversion.

## What Went Wrong

When converting the parsed input model to the internal model representation, a key could not be created for an entity. This typically means a key value is empty, whitespace-only, or otherwise invalid.

Keys in the model are derived from directory and file names. Each key must be a non-empty, lowercase snake_case string (letters, digits, and underscores, starting with a letter).

## How to Fix

1. Check that the directory or file name for the entity uses valid snake_case format
2. Ensure the name is not empty or whitespace-only
3. Verify the name starts with a lowercase letter and contains only `a-z`, `0-9`, and `_`

## Valid Key Format

Keys must match the pattern: `^[a-z][a-z0-9]*(_[a-z0-9]+)*$`

| Valid | Invalid |
|-------|---------|
| `order` | `Order` (uppercase) |
| `line_item` | `line-item` (hyphen) |
| `user_v2` | `2nd_user` (starts with digit) |
| `account` | `` (empty) |

## Common Causes

- A directory or file was renamed to an invalid format after initial creation
- A key reference in a JSON file points to a non-existent or misspelled entity
- An internal conversion step received an empty or malformed key string

## Related Errors

- **E11026**: Key has invalid format (caught during tree reading)
- **E11027**: Association filename has invalid format
