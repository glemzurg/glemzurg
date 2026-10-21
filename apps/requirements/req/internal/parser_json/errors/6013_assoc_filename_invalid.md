# Association Filename Invalid (E6013)

The association filename does not follow the expected pattern `<from>__<to>__<name>.assoc.json`.

## How to Fix

Association filenames must use double underscores to separate the three components:

```
customer__order__places.assoc.json
```

Each component must be lowercase snake_case.
