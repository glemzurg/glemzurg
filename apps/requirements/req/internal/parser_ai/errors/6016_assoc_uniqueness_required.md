# Association Uniqueness Invalid (E6016)

The association JSON file has an invalid `uniqueness` constraint.

## What Went Wrong

`uniqueness` is an array of constraint objects. Each entry must list at least one attribute SubKey on the from class, the to class, or both. An empty constraint object or a mapping with empty `from_attributes` and `to_attributes` arrays triggers this error.

## How to Fix

Provide one or more uniqueness tuples. List only `to_attributes` when uniqueness is implied per from-class instance:

```json
{
    "name": "Has Phases",
    "from_class_key": "family",
    "from_multiplicity": "1",
    "to_class_key": "phase",
    "to_multiplicity": "any",
    "uniqueness": [
        {"to_attributes": ["name"]},
        {"to_attributes": ["num"]}
    ]
}
```

List only `from_attributes` when uniqueness is implied per to-class instance. List both sides when the tuple spans endpoint attributes.

## Related Errors

- **E21120**: model conversion rejected a uniqueness tuple
- **E6012**: multiplicity format is invalid
