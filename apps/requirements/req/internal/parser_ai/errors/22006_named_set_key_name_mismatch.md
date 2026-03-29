# Named Set Key Does Not Match Name (E22006)

The named set filename does not match the expected key derived from the `name` field.

## What Went Wrong

Named set filenames must be derived from the `name` field by:
1. The name must start with `_` (underscore)
2. Strip the leading `_` from the name
3. Convert to lowercase
4. Replace spaces with underscores
5. Replace hyphens with underscores

The filename (without `.nset.json`) must exactly match the result (without the `_` prefix).

## Examples

| Name Field | Expected Filename |
|---|---|
| `"_OrderStatuses"` | `orderstatuses.nset.json` |
| `"_Payment Types"` | `payment_types.nset.json` |
| `"_Is-Valid-Set"` | `is_valid_set.nset.json` |

## How to Fix

### Option 1: Rename the File

If the `name` field is correct, rename the file to match:

```
named_sets/payment_types.nset.json
```

### Option 2: Change the Name

If the filename is correct, update the `name` field to match the current filename. Remember the name must start with `_`.

## Important

Do not change both the filename and the name at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E22005**: Named set name must start with underscore
- **E22001**: Named set name is required
