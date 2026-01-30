# Model Schema Violation (E1004)

The `model.json` file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your `model.json` file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing
- A field has the wrong type (e.g., number instead of string)
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string when `minLength: 1` is required)

## File Location

The `model.json` file must exist at the **root** of your model directory:

```
your_model/
├── model.json          <-- This file violates the schema
├── actors/
├── domains/
├── associations/
└── generalizations/
```

## Schema Requirements

The `model.json` file must conform to this schema:

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `minLength: 1` | The model's display name |

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `details` | string | none | Extended description of the model |

### Rules

- **No additional properties**: Only `name` and `details` are allowed
- **name must be non-empty**: At least one character required
- **All values must be strings**: Numbers, booleans, arrays, and objects are not valid

## Common Schema Violations

### 1. Missing Required `name` Field

```json
// WRONG: Missing 'name'
{
    "details": "My model description"
}

// CORRECT: Include 'name'
{
    "name": "My Model",
    "details": "My model description"
}
```

### 2. Empty `name` Field

```json
// WRONG: Empty string violates minLength: 1
{
    "name": ""
}

// CORRECT: Non-empty name
{
    "name": "My Model"
}
```

### 3. Wrong Type for `name`

```json
// WRONG: Number instead of string
{
    "name": 123
}

// WRONG: Boolean instead of string
{
    "name": true
}

// WRONG: Null instead of string
{
    "name": null
}

// WRONG: Array instead of string
{
    "name": ["My", "Model"]
}

// CORRECT: String value
{
    "name": "My Model"
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'version' is not in the schema
{
    "name": "My Model",
    "version": "1.0.0"
}

// WRONG: 'title' is not in the schema (use 'name')
{
    "name": "My Model",
    "title": "Also My Model"
}

// WRONG: 'description' is not in the schema (use 'details')
{
    "name": "My Model",
    "description": "This should be 'details'"
}

// CORRECT: Only allowed fields
{
    "name": "My Model",
    "details": "This is the correct field for description"
}
```

### 5. Wrong Type for `details`

```json
// WRONG: Object instead of string
{
    "name": "My Model",
    "details": {
        "description": "Some text",
        "author": "Someone"
    }
}

// CORRECT: String value
{
    "name": "My Model",
    "details": "Some text written by Someone"
}
```

## How to Read Schema Error Messages

Schema validation errors typically include:

1. **Path**: Where in the JSON the error occurred (e.g., `/name`, `/details`)
2. **Violation type**: What rule was broken (e.g., `required`, `type`, `minLength`, `additionalProperties`)
3. **Expected vs actual**: What was expected and what was found

### Example Error Messages

| Error Message | Meaning | Fix |
|--------------|---------|-----|
| `missing properties: 'name'` | The `name` field is required but missing | Add `"name": "..."` |
| `expected string, but got number` | Field value is wrong type | Use a string in quotes |
| `string is too short: minLength=1` | Empty string not allowed | Provide at least one character |
| `additionalProperties 'foo' not allowed` | Unknown field present | Remove the unknown field |

## Troubleshooting Checklist

1. **Check required fields**: Ensure `name` is present
2. **Check field types**: All values must be strings (in double quotes)
3. **Check for extra fields**: Remove any fields not listed in the schema
4. **Check for typos**: `Name` is not the same as `name` (case-sensitive)
5. **Check common alternatives**: Use `name` not `title`, use `details` not `description`

### Field Name Reference

| You Might Write | Correct Field Name |
|-----------------|-------------------|
| `title` | `name` |
| `description` | `details` |
| `desc` | `details` |
| `info` | `details` |
| `modelName` | `name` |

## Valid model.json Examples

### Minimal Valid File

```json
{
    "name": "My Model"
}
```

### Complete Valid File

```json
{
    "name": "Order Management System",
    "details": "Handles customer orders, inventory tracking, and fulfillment for the e-commerce platform. Includes integration points for payment processing and shipping providers."
}
```

## Understanding the Model File's Role

The `model.json` file is the entry point for your entire model. It defines:

1. **Identity**: The name uniquely identifies this model
2. **Documentation**: The details provide context for readers

### How model.json Connects to Other Files

```
model.json                    <-- You are here (defines model name)
    │
    ├── actors/               <-- Actor files reference this model
    │   └── *.actor.json
    │
    ├── {domain}/             <-- Domain directories
    │   ├── domain.json       <-- Domains belong to this model
    │   └── *.class.json      <-- Classes belong to domains
    │
    ├── associations/         <-- Associations connect classes
    │   └── *.assoc.json
    │
    └── generalizations/      <-- Generalizations define inheritance
        └── *.gen.json
```

The model name from `model.json` appears in:
- Generated documentation titles
- UML diagram package names
- Error messages for context
- Export file names

## Related Errors

- **E1001**: Name field is missing entirely (specific case of schema violation)
- **E1002**: Name field is empty (specific case of schema violation)
- **E1003**: JSON syntax is invalid (must fix this before schema can be checked)
