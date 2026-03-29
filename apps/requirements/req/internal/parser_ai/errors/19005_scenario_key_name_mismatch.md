# Scenario Key Does Not Match Name (E19005)

The scenario filename does not match the expected key derived from the `name` field.

## What Went Wrong

Scenario filenames must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The filename (without `.scenario.json`) must exactly match the result.

## Examples

| Name Field | Expected Filename |
|---|---|
| `"Happy Path"` | `happy_path.scenario.json` |
| `"Out-Of-Stock"` | `out_of_stock.scenario.json` |

## How to Fix

### Option 1: Rename the File

If the `name` field is correct, rename the file to match:

```
scenarios/happy_path.scenario.json
```

### Option 2: Change the Name

If the filename is correct, update the `name` field to match the current filename.

## Important

Do not change both the filename and the name at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E19001**: Scenario name is required
