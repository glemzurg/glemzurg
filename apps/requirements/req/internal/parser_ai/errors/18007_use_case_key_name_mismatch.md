# Use Case Key Does Not Match Name (E18007)

The use case directory name does not match the expected key derived from the `name` field in `use_case.json`.

## What Went Wrong

Use case directory names must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The directory name must exactly match the result.

## Examples

| Name Field | Expected Directory |
|---|---|
| `"Place Order"` | `place_order/` |
| `"Self-Checkout"` | `self_checkout/` |

## How to Fix

### Option 1: Rename the Directory

If the `name` field is correct, rename the directory to match:

```
use_cases/place_order/
```

### Option 2: Change the Name

If the directory name is correct, update the `name` field in `use_case.json` to match.

## Important

Do not change both the directory name and the name field at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E18001**: Use case name is required
