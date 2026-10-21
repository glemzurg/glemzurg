# Action Filename Does Not Match Name (E8006)

The action filename does not match the expected filename derived from the `name` field in the JSON content.

## What Went Wrong

Action filenames must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The filename (without `.json` extension) must exactly match the result.

## Examples

| Name Field | Expected Filename |
|---|---|
| `"Calculate Total"` | `calculate_total.json` |
| `"Re-Enable Promo Code"` | `re_enable_promo_code.json` |
| `"Fail On Name-DOB Mismatch"` | `fail_on_name_dob_mismatch.json` |
| `"Assign Role"` | `assign_role.json` |

## How to Fix

### Option 1: Rename the File

If the `name` field is correct, rename the file to match:

```
# Before
actions/create_promo.json          (name: "Create Promo Code Source")

# After
actions/create_promo_code_source.json
```

### Option 2: Change the Name

If the filename is correct, update the `name` field in the JSON to match:

```json
// File: actions/create_promo.json
{
    "name": "Create Promo"
}
```

## Important

Do not change both the filename and the name at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E11026**: Key has invalid format (must be lowercase snake_case)
- **E8001**: Action name is required
- **E8004**: Action JSON does not match schema
