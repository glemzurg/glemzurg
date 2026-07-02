# Actor Generalization Key Does Not Match Name (E2010)

The actor generalization filename does not match the expected key derived from the `name` field.

## What Went Wrong

Actor generalization filenames must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The filename (without `.agen.json`) must exactly match the result.

## Examples

| Name Field | Expected Filename |
|---|---|
| `"User Type"` | `user_type.agen.json` |
| `"External-Actor"` | `external_actor.agen.json` |

## How to Fix

### Option 1: Rename the File

If the `name` field is correct, rename the file to match:

```
actor_generalizations/user_type.agen.json
```

### Option 2: Change the Name

If the filename is correct, update the `name` field to match the current filename.

## Important

Do not change both the filename and the name at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E12001**: Actor generalization name is required
