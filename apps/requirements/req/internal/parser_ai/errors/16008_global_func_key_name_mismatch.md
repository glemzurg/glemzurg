# Global Function Key Does Not Match Name (E16008)

The global function filename does not match the expected key derived from the `name` field.

## What Went Wrong

Global function filenames must be derived from the `name` field by:
1. The name must start with `_` (underscore)
2. Strip the leading `_` from the name
3. Convert to lowercase
4. Replace spaces with underscores
5. Replace hyphens with underscores
6. Prepend `_` back to form the filename key

The filename (without `.json`) must exactly match the result.

## Examples

| Name Field | Expected Filename |
|---|---|
| `"_Max"` | `_max.json` |
| `"_Set Of Values"` | `_set_of_values.json` |
| `"_Is-Valid"` | `_is_valid.json` |

## How to Fix

### Option 1: Rename the File

If the `name` field is correct, rename the file to match:

```
global_functions/_max.json
```

### Option 2: Change the Name

If the filename is correct, update the `name` field to match the current filename.

## Important

Do not change both the filename and the name at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E16005**: Global function name must start with underscore
- **E16001**: Global function name is required
