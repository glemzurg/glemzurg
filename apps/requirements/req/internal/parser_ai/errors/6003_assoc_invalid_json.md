# Association Invalid JSON (E6003)

The association JSON file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your association file but encountered a JSON syntax error. The file contents are not valid JSON.

## File Location

Association files are located in the `associations/` directory at the model root:

```
your_model/
├── model.json
├── associations/
│   └── order_has_items.assoc.json    <-- This file contains invalid JSON
└── order_management/
    └── order.class.json
```

## How to Fix

Ensure your association file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
    "from_multiplicity": "1",
    "to_class_key": "order_management.order_item",
    "to_multiplicity": "1..*"
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "Order Contains Items"
    "from_class_key": "order_management.order"
}

// CORRECT
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order"
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order",
}

// CORRECT
{
    "name": "Order Contains Items",
    "from_class_key": "order_management.order"
}
```

### 3. Single Quotes Instead of Double Quotes

```json
// WRONG
{
    'name': 'Order Contains Items'
}

// CORRECT
{
    "name": "Order Contains Items"
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like [JSONLint](https://jsonlint.com/) can pinpoint errors
2. **Check encoding**: Ensure UTF-8 encoding without BOM
3. **Check all required fields**: name, from_class_key, from_multiplicity, to_class_key, to_multiplicity

### Command Line Validation

```bash
# Validate JSON with jq
jq . order_has_items.assoc.json

# Validate with Python
python3 -m json.tool order_has_items.assoc.json
```

## Related Errors

- **E6001**: Name field is missing (JSON is valid but incomplete)
- **E6002**: Name is empty (JSON is valid but value is wrong)
- **E6004**: JSON is valid but doesn't match the schema
