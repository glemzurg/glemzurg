# Query Filename Does Not Match Name (E9006)

The query filename does not match the expected filename derived from the `name` field in the JSON content.

## What Went Wrong

Query filenames must be derived from the `name` field by:
1. Converting to lowercase
2. Replacing spaces with underscores
3. Replacing hyphens with underscores

The filename (without `.json` extension) must exactly match the result.

## Examples

| Name Field | Expected Filename |
|---|---|
| `"Get Subtotal"` | `get_subtotal.json` |
| `"Query Needs FRS Re-Evaluation"` | `query_needs_frs_re_evaluation.json` |
| `"Check Self-Limitation"` | `check_self_limitation.json` |

## How to Fix

### Option 1: Rename the File

If the `name` field is correct, rename the file to match:

```
# Before
queries/get_total.json          (name: "Get Subtotal")

# After
queries/get_subtotal.json
```

### Option 2: Change the Name

If the filename is correct, update the `name` field in the JSON to match:

```json
// File: queries/get_total.json
{
    "name": "Get Total"
}
```

## Important

Do not change both the filename and the name at the same time. Pick the one that is correct and adjust the other.

## Related Errors

- **E11026**: Key has invalid format (must be lowercase snake_case)
- **E9001**: Query name is required
- **E9004**: Query JSON does not match schema
