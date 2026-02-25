# Logic Invalid JSON (E14003)

A logic object contains invalid JSON syntax and cannot be parsed.

## What Went Wrong

The parser attempted to read a logic object but encountered a JSON syntax error. The content is not valid JSON.

## Where Logic Objects Appear

Logic objects are embedded sub-objects within action, query, state machine, class, and model JSON files. They are used for:

- **Requires**: Preconditions on actions and queries
- **Guarantees**: Postconditions on actions and queries
- **Safety rules**: Constraints on actions
- **Invariants**: Model-level invariants
- **Guards**: Conditions on state machine transitions
- **Derivation policies**: Rules for derived attributes

## How to Fix

Ensure your logic object contains valid JSON. A minimal valid logic object looks like:

```json
{
    "description": "Order total must be positive"
}
```

## Common JSON Syntax Errors

### 1. Missing Commas Between Properties

```json
// WRONG: Missing comma after "description"
{
    "description": "Order total must be positive"
    "notation": "tla_plus"
}

// CORRECT: Comma separates properties
{
    "description": "Order total must be positive",
    "notation": "tla_plus"
}
```

### 2. Trailing Commas

```json
// WRONG: Trailing comma after last property
{
    "description": "Order total must be positive",
    "notation": "tla_plus",
}

// CORRECT: No trailing commas
{
    "description": "Order total must be positive",
    "notation": "tla_plus"
}
```

### 3. Single Quotes Instead of Double Quotes

```json
// WRONG: Single quotes
{
    'description': 'Order total must be positive'
}

// CORRECT: Double quotes
{
    "description": "Order total must be positive"
}
```

### 4. Unescaped Special Characters in Strings

```json
// WRONG: Unescaped quotes inside specification
{
    "description": "Check the "total" field",
    "specification": "total > 0"
}

// CORRECT: Escaped inner quotes
{
    "description": "Check the \"total\" field",
    "specification": "total > 0"
}
```

### 5. Missing Braces

```json
// WRONG: Missing closing brace
{
    "description": "Order total must be positive"

// CORRECT: Matching braces
{
    "description": "Order total must be positive"
}
```

## Troubleshooting Checklist

1. **Use a JSON validator**: Online tools like [JSONLint](https://jsonlint.com/) can pinpoint errors
2. **Check encoding**: Ensure UTF-8 encoding without BOM
3. **Look for invisible characters**: Copy-paste can introduce hidden characters
4. **Check the parent file**: The logic object is embedded in another JSON file, so the entire file must be valid

### Command Line Validation

```bash
# Validate the parent JSON file with jq
jq . your_file.json

# Validate with Python
python3 -m json.tool your_file.json
```

## Valid Logic Object Template

```json
{
    "description": "Your logic description here"
}
```

Or with all optional fields:

```json
{
    "description": "User must be authenticated",
    "notation": "tla_plus",
    "specification": "user.authenticated = TRUE"
}
```

## Related Errors

- **E14001**: Description field is missing (JSON is valid but incomplete)
- **E14002**: Description is empty (JSON is valid but value is wrong)
- **E14004**: JSON is valid but doesn't match the schema
