# Generalization Invalid JSON (E10003)

The generalization JSON file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your generalization file but encountered a JSON syntax error. The file contents are not valid JSON.

## File Location

Generalization files are located in the `generalizations/` directory:

```
your_model/
├── model.json
└── generalizations/
    └── payment_types.gen.json    <-- This file contains invalid JSON syntax
```

## How to Fix

Ensure your generalization file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card", "billing.bank_transfer"]
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "Payment Types"
    "superclass_key": "billing.payment"
}

// CORRECT: Comma separates properties
{
    "name": "Payment Types",
    "superclass_key": "billing.payment"
}
```

### 2. Missing Commas in Arrays

```json
// WRONG: Missing comma in array
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card" "billing.bank_transfer"]
}

// CORRECT: Commas between array elements
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card", "billing.bank_transfer"]
}
```

### 3. Trailing Commas

```json
// WRONG: Trailing comma after last property
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"],
}

// WRONG: Trailing comma in array
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card",]
}

// CORRECT: No trailing commas
{
    "name": "Payment Types",
    "superclass_key": "billing.payment",
    "subclass_keys": ["billing.credit_card"]
}
```

### 4. Single Quotes Instead of Double Quotes

```json
// WRONG: Single quotes
{
    'name': 'Payment Types'
}

// CORRECT: Double quotes
{
    "name": "Payment Types"
}
```

### 5. Boolean Values Must Be Lowercase

```json
// WRONG: True/False with capitals
{
    "name": "Payment Types",
    "is_complete": True,
    "is_static": False
}

// CORRECT: Lowercase true/false
{
    "name": "Payment Types",
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
jq . generalizations/payment_types.gen.json

# Validate with Python
python3 -m json.tool generalizations/payment_types.gen.json
```

## Valid Generalization Template

```json
{
    "name": "Your Generalization Name",
    "superclass_key": "domain.superclass",
    "subclass_keys": ["domain.subclass1", "domain.subclass2"]
}
```

## Related Errors

- **E10001**: Name field is missing (JSON is valid but incomplete)
- **E10002**: Name is empty (JSON is valid but value is wrong)
- **E10004**: JSON is valid but doesn't match the schema
