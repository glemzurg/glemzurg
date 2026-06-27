# Association Uniqueness Required (E6016)

The association JSON file has no `uniqueness` field, or the field is empty or whitespace only.

## What Went Wrong

Every association must declare how many links may exist between the same from-class and to-class instances. The `uniqueness` field is required and uses the same multiplicity format as `from_multiplicity` and `to_multiplicity`.

## How to Fix

Provide a valid multiplicity value for `uniqueness`. Use `"any"` when there is no per-pair cap:

```json
{
    "name": "Order Contains Items",
    "from_class_key": "order",
    "from_multiplicity": "1",
    "to_class_key": "line_item",
    "to_multiplicity": "1..*",
    "uniqueness": "any"
}
```

Use `"0..1"` when at most one link may exist between the same two instances:

```json
    "uniqueness": "0..1"
```

## Related Errors

- **E6007**: from_multiplicity is required
- **E6008**: to_multiplicity is required
- **E6012**: multiplicity format is invalid
- **E11016**: tree validation reports invalid multiplicity format