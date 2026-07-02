# Attribute Key Does Not Match Name (E5013)

The attribute map key in `class.json` does not match the expected key derived from the `name` field.

## What Went Wrong

Attribute map keys must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The map key must exactly match the result.

## Examples

| Name Field | Expected Key |
|---|---|
| `"Order Total"` | `order_total` |
| `"Is-Active"` | `is_active` |
| `"Status"` | `status` |

## How to Fix

### Option 1: Rename the Key

If the `name` field is correct, rename the map key to match:

```json
{
    "attributes": {
        "order_total": {
            "name": "Order Total",
            "data_type_rules": "[0..unconstrained] at 0.01 dollars"
        }
    }
}
```

### Option 2: Change the Name

If the key is correct, update the `name` field to match:

```json
{
    "attributes": {
        "status": {
            "name": "Status",
            "data_type_rules": "enum of active, inactive"
        }
    }
}
```

## Important

Do not change both the key and the name at the same time. Pick the one that is correct and adjust the other. Also update any indexes or logic that reference this attribute key.

## Related Errors

- **E5008**: Attribute name is empty
