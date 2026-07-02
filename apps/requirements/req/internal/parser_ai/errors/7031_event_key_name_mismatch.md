# Event Key Does Not Match Name (E7031)

The state machine map key for an event does not match the expected key derived from the `name` field.

## What Went Wrong

Event map keys must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The map key must exactly match the result.

## Examples

| Name Field | Expected Key |
|---|---|
| `"Submit Order"` | `submit_order` |
| `"Re-Enable"` | `re_enable` |
| `"Cancel"` | `cancel` |

## How to Fix

### Option 1: Rename the Key

If the `name` field is correct, rename the map key to match:

```json
{
    "events": {
        "submit_order": {
            "name": "Submit Order"
        }
    }
}
```

### Option 2: Change the Name

If the key is correct, update the `name` field to match:

```json
{
    "events": {
        "submit": {
            "name": "Submit"
        }
    }
}
```

## Important

Do not change both the key and the name at the same time. Pick the one that is correct and adjust the other. Also update any transitions that reference this event via `event_key`.

## Related Errors

- **E7028**: Duplicate event name
- **E7009**: Event name is required
