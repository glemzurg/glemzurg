# Association Filename Name Mismatch (E6015)

The name component of the association filename does not match the expected value derived from the `name` field in the JSON content.

## What Went Wrong

Association filenames have the format `{from}--{to}--{name}.assoc.json`. The `{name}` component must be derived from the `name` field in the JSON by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

## Examples

| Name Field | Expected Name Component |
|---|---|
| `"Order Lines"` | `order_lines` |
| `"Wallet Has Bets"` | `wallet_has_bets` |
| `"User Sessions"` | `user_sessions` |

### Mismatch Example

If the file is `wallet--bet--wallet_bets.assoc.json` but the JSON contains `"name": "Wallet Has Bets"`, the expected filename would be `wallet--bet--wallet_has_bets.assoc.json`.

## How to Fix

### Option 1: Rename the File

If the `name` field is correct, rename the file so the name component matches:

```
# Before
wallet--bet--wallet_bets.assoc.json    (name: "Wallet Has Bets")

# After
wallet--bet--wallet_has_bets.assoc.json
```

### Option 2: Change the Name

If the filename is correct, update the `name` field in the JSON to match:

```json
{
    "name": "Wallet Bets",
    "from_class_key": "wallet",
    "to_class_key": "bet"
}
```

## Important

Do not change both the filename and the name at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E6013**: Association filename format invalid
- **E11027**: Association filename has invalid format (wrong number of parts)
- **E11028**: Association filename has invalid component (not snake_case)
