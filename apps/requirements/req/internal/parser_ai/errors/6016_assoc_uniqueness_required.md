# Association Uniqueness Invalid (E6016)

The association JSON file has an invalid `uniqueness` object.

## What Went Wrong

The `uniqueness` mapping must list at least one attribute SubKey on the from class, the to class, or both. An empty `uniqueness` object or a mapping with empty `from_attributes` and `to_attributes` arrays triggers this error.

## How to Fix

Provide a valid uniqueness tuple. List only `to_attributes` when uniqueness is implied per from-class instance:

```json
{
    "name": "Configures Customers For",
    "from_class_key": "partner",
    "from_multiplicity": "1",
    "to_class_key": "jurisdiction",
    "to_multiplicity": "any",
    "uniqueness": {
        "to_attributes": ["jurisdiction_code"]
    }
}
```

List only `from_attributes` when uniqueness is implied per to-class instance. List both sides when the tuple spans endpoint attributes.

## Related Errors

- **E21120**: model conversion rejected a uniqueness tuple
- **E6012**: multiplicity format is invalid