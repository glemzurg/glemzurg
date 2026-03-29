# Subdomain Key Does Not Match Name (E4007)

The subdomain directory name does not match the expected key derived from the `name` field in `subdomain.json`.

## What Went Wrong

Subdomain directory names must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The directory name must exactly match the result.

## Examples

| Name Field | Expected Directory |
|---|---|
| `"Default"` | `default/` |
| `"Order Processing"` | `order_processing/` |

## How to Fix

### Option 1: Rename the Directory

If the `name` field is correct, rename the directory to match:

```
subdomains/order_processing/
```

### Option 2: Change the Name

If the directory name is correct, update the `name` field in `subdomain.json` to match.

## Important

Do not change both the directory name and the name field at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E4005**: Duplicate subdomain key
- **E4001**: Subdomain name is required
