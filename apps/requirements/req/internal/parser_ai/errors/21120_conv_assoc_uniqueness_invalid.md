# Association Uniqueness Conversion Failed (E21120)

After conversion, model validation rejected an association `uniqueness` constraint.

## What Went Wrong

A uniqueness tuple failed core validation. Common causes: unknown attribute SubKey on an endpoint class, an empty tuple with neither `from_attributes` nor `to_attributes`, or two identical uniqueness tuples on the same association.

## How to Fix

Ensure each listed SubKey exists on the matching endpoint class, and that each constraint is distinct:

```json
"uniqueness": [
    {"to_attributes": ["jurisdiction_code"]}
]
```

Re-run tree validation after fixing the association JSON.

## Related Errors

- **E6016**: parse-time uniqueness validation
- **E11016**: tree validation reports invalid multiplicity format
