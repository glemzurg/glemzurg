# Class Key Does Not Match Name (E5012)

The class directory name does not match the expected key derived from the `name` field in `class.json`.

## What Went Wrong

Class directory names must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The directory name must exactly match the result.

## Examples

| Name Field | Expected Directory |
|---|---|
| `"Book Order"` | `book_order/` |
| `"Line-Item"` | `line_item/` |

## How to Fix

### Option 1: Rename the Directory

If the `name` field is correct, rename the directory to match:

```
classes/book_order/
```

### Option 2: Change the Name

If the directory name is correct, update the `name` field in `class.json` to match.

## Important

Do not change both the directory name and the name field at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E5005**: Duplicate class key
- **E5001**: Class name is required
