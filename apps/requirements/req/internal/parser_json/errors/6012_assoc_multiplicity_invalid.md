# Association Multiplicity Invalid (E6012)

The multiplicity value is not in a valid format.

## How to Fix

Valid multiplicity values: `0..1`, `1`, `0..*`, `1..*`, `*`.

```json
{
    "name": "Places",
    "from_class_key": "customer",
    "from_multiplicity": "1",
    "to_class_key": "order",
    "to_multiplicity": "0..*"
}
```
