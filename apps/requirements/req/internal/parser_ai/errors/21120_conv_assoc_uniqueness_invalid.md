# Association Uniqueness Conversion Failed (E21120)

After conversion, model validation rejected an association `uniqueness` object.

## What Went Wrong

The uniqueness tuple failed core validation. Common causes: unknown attribute SubKey on an endpoint class, or an empty tuple with neither `from_attributes` nor `to_attributes`.

## How to Fix

Ensure each listed SubKey exists on the matching endpoint class:

```json
"uniqueness": {
    "to_attributes": ["jurisdiction_code"]
}
```

Re-run tree validation after fixing the association JSON.

## Related Errors

- **E6016**: parse-time uniqueness validation
- **E11016**: tree validation reports invalid multiplicity format