# Use Case Generalization Key Does Not Match Name (E13008)

The use case generalization filename does not match the expected key derived from the `name` field.

## What Went Wrong

Use case generalization filenames must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The filename (without `.ucgen.json`) must exactly match the result.

## Examples

| Name Field | Expected Filename |
|---|---|
| `"Order Management"` | `order_management.ucgen.json` |
| `"Self-Service"` | `self_service.ucgen.json` |

## How to Fix

### Option 1: Rename the File

If the `name` field is correct, rename the file to match:

```
use_case_generalizations/order_management.ucgen.json
```

### Option 2: Change the Name

If the filename is correct, update the `name` field to match the current filename.

## Important

Do not change both the filename and the name at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E13001**: Use case generalization name is required
