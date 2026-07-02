# Domain Key Does Not Match Name (E3007)

The domain directory name does not match the expected key derived from the `name` field in `domain.json`.

## What Went Wrong

Domain directory names must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The directory name must exactly match the result.

## Examples

| Name Field | Expected Directory |
|---|---|
| `"Order Fulfillment"` | `order_fulfillment/` |
| `"User-Management"` | `user_management/` |

## How to Fix

### Option 1: Rename the Directory

If the `name` field is correct, rename the directory to match:

```
domains/order_fulfillment/
```

### Option 2: Change the Name

If the directory name is correct, update the `name` field in `domain.json` to match.

## Important

Do not change both the directory name and the name field at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E3005**: Duplicate domain key
- **E3001**: Domain name is required
