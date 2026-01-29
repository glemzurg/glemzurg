# Class Invalid JSON (E5003)

The class JSON file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your class file but encountered a JSON syntax error. The file contents are not valid JSON.

## File Location

Class files are located within domain directories:

```
your_model/
├── model.json
└── order_management/           <-- Domain directory
    ├── domain.json
    └── order.class.json        <-- This file contains invalid JSON
```

## How to Fix

Ensure your class file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "Order"
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "Order"
    "details": "A customer order"
}

// CORRECT: Comma separates properties
{
    "name": "Order",
    "details": "A customer order"
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma after last property
{
    "name": "Order",
    "details": "A customer order",
}

// CORRECT: No trailing comma
{
    "name": "Order",
    "details": "A customer order"
}
```

### 3. Single Quotes Instead of Double Quotes

```json
// WRONG: Single quotes
{
    'name': 'Order'
}

// CORRECT: Double quotes
{
    "name": "Order"
}
```

### 4. Unquoted Keys

```json
// WRONG: Unquoted key
{
    name: "Order"
}

// CORRECT: Quoted key
{
    "name": "Order"
}
```

### 5. Boolean Values Must Be Lowercase

```json
// WRONG: True/False with capitals
{
    "name": "Order",
    "attributes": {
        "is_active": {
            "name": "Is Active",
            "nullable": True
        }
    }
}

// CORRECT: Lowercase true/false
{
    "name": "Order",
    "attributes": {
        "is_active": {
            "name": "Is Active",
            "nullable": true
        }
    }
}
```

### 6. Missing Closing Braces in Nested Objects

```json
// WRONG: Missing closing brace for attribute
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date"
    }
}

// CORRECT: All braces closed
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date"
        }
    }
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like [JSONLint](https://jsonlint.com/) can pinpoint errors
2. **Check encoding**: Ensure UTF-8 encoding without BOM
3. **Look for invisible characters**: Copy-paste can introduce hidden characters
4. **Verify nested structure**: Attributes contain nested objects that all need proper braces

### Command Line Validation

```bash
# Validate JSON with jq
jq . order_management/order.class.json

# Validate with Python
python3 -m json.tool order_management/order.class.json
```

## Valid Class Template

```json
{
    "name": "Your Class Name",
    "details": "Description of the class",
    "attributes": {
        "attribute_key": {
            "name": "Attribute Display Name",
            "data_type_rules": "Type and constraints"
        }
    }
}
```

## Related Errors

- **E5001**: Name field is missing (JSON is valid but incomplete)
- **E5002**: Name is empty (JSON is valid but value is wrong)
- **E5004**: JSON is valid but doesn't match the schema
