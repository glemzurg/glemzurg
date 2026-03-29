# State Key Does Not Match Name (E7030)

The state machine map key for a state does not match the expected key derived from the `name` field.

## What Went Wrong

State map keys must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The map key must exactly match the result.

## Examples

| Name Field | Expected Key |
|---|---|
| `"Pending Approval"` | `pending_approval` |
| `"In-Progress"` | `in_progress` |
| `"New"` | `new` |

## How to Fix

### Option 1: Rename the Key

If the `name` field is correct, rename the map key to match:

```json
{
    "states": {
        "pending_approval": {
            "name": "Pending Approval"
        }
    }
}
```

### Option 2: Change the Name

If the key is correct, update the `name` field to match:

```json
{
    "states": {
        "pending": {
            "name": "Pending"
        }
    }
}
```

## Important

Do not change both the key and the name at the same time. Pick the one that is correct and adjust the other. Also update any transitions that reference this state via `from_state_key` or `to_state_key`.

## Related Errors

- **E7027**: Duplicate state name
- **E7003**: State name is required
