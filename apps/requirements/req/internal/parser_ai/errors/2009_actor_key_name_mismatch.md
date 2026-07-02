# Actor Key Does Not Match Name (E2009)

The actor filename does not match the expected key derived from the `name` field.

## What Went Wrong

Actor filenames must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The filename (without `.actor.json`) must exactly match the result.

## Examples

| Name Field | Expected Filename |
|---|---|
| `"Customer"` | `customer.actor.json` |
| `"System Admin"` | `system_admin.actor.json` |
| `"Third-Party"` | `third_party.actor.json` |

## How to Fix

### Option 1: Rename the File

If the `name` field is correct, rename the file to match:

```
actors/system_admin.actor.json
```

with content:
```json
{
    "name": "System Admin"
}
```

### Option 2: Change the Name

If the filename is correct, update the `name` field to match:

```json
{
    "name": "Customer"
}
```

in file `actors/customer.actor.json`.

## Important

Do not change both the filename and the name at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E2007**: Duplicate actor key
- **E2001**: Actor name is required
