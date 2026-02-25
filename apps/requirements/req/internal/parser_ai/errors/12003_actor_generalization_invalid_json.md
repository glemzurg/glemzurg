# Actor Generalization Invalid JSON (E12003)

The actor generalization JSON file contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read your actor generalization file but encountered a JSON syntax error. The file contents are not valid JSON.

## File Location

Actor generalization files are located in the `actor_generalizations/` directory:

```
your_model/
├── model.json
└── actor_generalizations/
    └── user_types.actor_gen.json    <-- This file contains invalid JSON syntax
```

## How to Fix

Ensure your actor generalization file contains valid JSON. A minimal valid file looks like:

```json
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer", "guest"]
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "name"
{
    "name": "User Types"
    "superclass_key": "customer"
}

// CORRECT: Comma separates properties
{
    "name": "User Types",
    "superclass_key": "customer"
}
```

### 2. Missing Commas in Arrays

```json
// WRONG: Missing comma in array
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer" "guest"]
}

// CORRECT: Commas between array elements
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer", "guest"]
}
```

### 3. Trailing Commas

```json
// WRONG: Trailing comma after last property
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"],
}

// WRONG: Trailing comma in array
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer",]
}

// CORRECT: No trailing commas
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer"]
}
```

### 4. Single Quotes Instead of Double Quotes

```json
// WRONG: Single quotes
{
    'name': 'User Types'
}

// CORRECT: Double quotes
{
    "name": "User Types"
}
```

### 5. Boolean Values Must Be Lowercase

```json
// WRONG: True/False with capitals
{
    "name": "User Types",
    "is_complete": True,
    "is_static": False
}

// CORRECT: Lowercase true/false
{
    "name": "User Types",
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
jq . actor_generalizations/user_types.actor_gen.json

# Validate with Python
python3 -m json.tool actor_generalizations/user_types.actor_gen.json
```

## Valid Actor Generalization Template

```json
{
    "name": "Your Actor Generalization Name",
    "superclass_key": "parent_actor",
    "subclass_keys": ["child_actor1", "child_actor2"]
}
```

## Related Errors

- **E12001**: Name field is missing (JSON is valid but incomplete)
- **E12002**: Name is empty (JSON is valid but value is wrong)
- **E12004**: JSON is valid but doesn't match the schema
