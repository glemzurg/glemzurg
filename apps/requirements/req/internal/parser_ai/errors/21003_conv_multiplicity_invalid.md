# Conversion Multiplicity Invalid (E21003)

A multiplicity value could not be parsed during model conversion.

## What Went Wrong

An association has a `from_multiplicity` or `to_multiplicity` value that could not be converted to a valid multiplicity during model conversion. This is a defensive check - multiplicity format should already be validated before conversion.

## Valid Multiplicity Formats

| Format | Meaning | Example Use |
|--------|---------|-------------|
| `"1"` | Exactly one | Required 1-to-1 |
| `"0..1"` | Zero or one | Optional relationship |
| `"*"` | Zero or more | Optional many |
| `"1..*"` | One or more | Required many |
| `"n"` | Exactly n | `"3"` means exactly 3 |
| `"n..m"` | Range from n to m | `"2..5"` means 2 to 5 |
| `"n..*"` | n or more | `"3..*"` means 3 or more |

## Rules

1. Single number means exactly that count
2. `*` represents unbounded (zero or more)
3. Upper bound must be >= lower bound
4. Numbers must be non-negative integers

## How to Fix

Use a valid multiplicity format in your association file:

```json
{
    "name": "Order Lines",
    "from_class_key": "order",
    "from_multiplicity": "1",
    "to_class_key": "line_item",
    "to_multiplicity": "1..*"
}
```

## Related Errors

- **E11016**: Association multiplicity invalid (caught during tree validation)
- **E6007**: from_multiplicity is required
- **E6008**: to_multiplicity is required
