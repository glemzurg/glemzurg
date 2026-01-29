# Association Multiplicity Invalid (E11016)

An association has an invalid multiplicity format.

## What Went Wrong

An association file has a `from_multiplicity` or `to_multiplicity` field with a value that doesn't match the valid multiplicity format.

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
| `"0..n"` | Zero to n | `"0..3"` means 0 to 3 |

## Rules

1. Single number means exactly that count
2. `*` represents unbounded (no upper limit)
3. Upper bound must be >= lower bound (unless unbounded)
4. Numbers must be non-negative integers

## How to Fix

Use a valid multiplicity format:

```json
{
    "name": "Order Lines",
    "from_class_key": "order",
    "from_multiplicity": "1",
    "to_class_key": "line_item",
    "to_multiplicity": "1..*"
}
```

## Common Mistakes

```json
// WRONG: Using text
{
    "from_multiplicity": "one"
}

// WRONG: Invalid range (upper < lower)
{
    "from_multiplicity": "5..3"
}

// WRONG: Negative numbers
{
    "from_multiplicity": "-1"
}

// WRONG: Missing format
{
    "from_multiplicity": ""
}

// WRONG: Using wrong separator
{
    "from_multiplicity": "1-*"
}
// Should be:
{
    "from_multiplicity": "1..*"
}
```

## Valid Examples

```json
// Exactly one (required)
"from_multiplicity": "1"

// Zero or one (optional)
"from_multiplicity": "0..1"

// Zero or more
"to_multiplicity": "*"

// One or more (required, unbounded)
"to_multiplicity": "1..*"

// Specific range
"to_multiplicity": "2..5"

// Exactly three
"to_multiplicity": "3"
```

## Related Errors

- **E6007**: from_multiplicity is required
- **E6008**: to_multiplicity is required
