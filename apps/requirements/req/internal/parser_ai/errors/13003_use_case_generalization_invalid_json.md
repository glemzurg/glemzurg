# Use Case Generalization Invalid JSON (E13003)

The use case generalization JSON file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your use case generalization file but encountered a JSON syntax error. The file contents are not valid JSON.

## File Location

Use case generalization files are located in the `use_case_generalizations/` directory:

```
your_model/
├── model.json
└── use_case_generalizations/
    └── order_types.uc_gen.json    <-- This file contains invalid JSON syntax
```

## How to Fix

Ensure your use case generalization file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order", "process_in_store_order"]
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "Order Types"
    "superclass_key": "process_order"
}

// CORRECT: Comma separates properties
{
    "name": "Order Types",
    "superclass_key": "process_order"
}
```

### 2. Missing Commas in Arrays

```json
// WRONG: Missing comma in array
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order" "process_in_store_order"]
}

// CORRECT: Commas between array elements
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order", "process_in_store_order"]
}
```

### 3. Trailing Commas

```json
// WRONG: Trailing comma after last property
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"],
}

// WRONG: Trailing comma in array
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order",]
}

// CORRECT: No trailing commas
{
    "name": "Order Types",
    "superclass_key": "process_order",
    "subclass_keys": ["process_online_order"]
}
```

### 4. Single Quotes Instead of Double Quotes

```json
// WRONG: Single quotes
{
    'name': 'Order Types'
}

// CORRECT: Double quotes
{
    "name": "Order Types"
}
```

### 5. Boolean Values Must Be Lowercase

```json
// WRONG: True/False with capitals
{
    "name": "Order Types",
    "is_complete": True,
    "is_static": False
}

// CORRECT: Lowercase true/false
{
    "name": "Order Types",
    "is_complete": true,
    "is_static": false
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like [JSONLint](https://jsonlint.com/) can pinpoint errors
2. **Check encoding**: Ensure UTF-8 encoding without BOM
3. **Look for invisible characters**: Copy-paste can introduce hidden characters
4. **Verify array syntax**: Square brackets `[]` with comma-separated strings

### Command Line Validation

```bash
# Validate JSON with jq
jq . use_case_generalizations/order_types.uc_gen.json

# Validate with Python
python3 -m json.tool use_case_generalizations/order_types.uc_gen.json
```

## Valid Use Case Generalization Template

```json
{
    "name": "Your Use Case Generalization Name",
    "superclass_key": "parent_use_case",
    "subclass_keys": ["child_use_case_1", "child_use_case_2"]
}
```

## Related Errors

- **E13001**: Name field is missing (JSON is valid but incomplete)
- **E13002**: Name is empty (JSON is valid but value is wrong)
- **E13004**: JSON is valid but doesn't match the schema
