# Guard Key Does Not Match Name (E7032)

The state machine map key for a guard does not match the expected key derived from the `name` field.

## What Went Wrong

Guard map keys must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The map key must exactly match the result.

## Examples

| Name Field | Expected Key |
|---|---|
| `"Has Items"` | `has_items` |
| `"Is-Active"` | `is_active` |
| `"Authorized"` | `authorized` |

## How to Fix

### Option 1: Rename the Key

If the `name` field is correct, rename the map key to match:

```json
{
    "guards": {
        "has_items": {
            "name": "Has Items"
        }
    }
}
```

### Option 2: Change the Name

If the key is correct, update the `name` field to match:

```json
{
    "guards": {
        "authorized": {
            "name": "Authorized"
        }
    }
}
```

## Important

Do not change both the key and the name at the same time. Pick the one that is correct and adjust the other. Also update any transitions that reference this guard via `guard_key`.

## Related Errors

- **E7029**: Duplicate guard name
- **E7014**: Guard name is required
