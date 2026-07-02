# Class Generalization Key Does Not Match Name (E10014)

The class generalization filename does not match the expected key derived from the `name` field.

## What Went Wrong

Class generalization filenames must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The filename (without `.cgen.json`) must exactly match the result.

## Examples

| Name Field | Expected Filename |
|---|---|
| `"Payment Type"` | `payment_type.cgen.json` |
| `"Medium"` | `medium.cgen.json` |

## How to Fix

### Option 1: Rename the File

If the `name` field is correct, rename the file to match:

```
class_generalizations/payment_type.cgen.json
```

### Option 2: Change the Name

If the filename is correct, update the `name` field to match the current filename.

## Important

Do not change both the filename and the name at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E10010**: Duplicate class generalization key
- **E10001**: Class generalization name is required
