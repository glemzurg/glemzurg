# Logic Schema Violation (E14004)

A logic object contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read a logic object as valid JSON, but its structure or content violates the schema rules. This typically means:

- The required `description` field is missing
- A field has the wrong type
- An unknown field is present
- A field value doesn't meet constraints (e.g., empty string)

## Where Logic Objects Appear

Logic objects are embedded sub-objects within action, query, state machine, class, and model JSON files. They are used for:

- **Requires**: Preconditions on actions and queries
- **Guarantees**: Postconditions on actions and queries
- **Safety rules**: Constraints on actions
- **Invariants**: Model-level invariants
- **Guards**: Conditions on state machine transitions
- **Derivation policies**: Rules for derived attributes

## Schema Requirements

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `description` | string | `minLength: 1` | Human-readable explanation of the logic |

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `type` | string | enum: `assessment`, `state_change`, `query`, `safety_rule`, `value`, `let` | The kind of logic specification |
| `target` | string | `minLength: 1` | Target identifier (required for `state_change`, `query`, `let`) |
| `target_type_spec` | string | none | TLA+ type declaration for the target (e.g., `"Int"`, `"STRING"`) |
| `notation` | string | enum: `tla_plus` | Formal notation used |
| `specification` | string | none | Formal specification in the given notation |

## Common Schema Violations

### 1. Missing Required Field

```json
// WRONG: Missing 'description'
{
    "notation": "tla_plus",
    "specification": "x > 0"
}

// CORRECT: Description is present
{
    "description": "Value must be positive",
    "notation": "tla_plus",
    "specification": "x > 0"
}
```

### 2. Empty Description

```json
// WRONG: Empty description
{
    "description": "",
    "notation": "tla_plus",
    "specification": "x > 0"
}

// CORRECT: Non-empty description
{
    "description": "Value must be positive",
    "notation": "tla_plus",
    "specification": "x > 0"
}
```

### 3. Wrong Types

```json
// WRONG: description is a number, not a string
{
    "description": 42
}

// WRONG: notation is an array, not a string
{
    "description": "Value must be positive",
    "notation": ["tla_plus"]
}

// CORRECT: Proper types
{
    "description": "Value must be positive",
    "notation": "tla_plus"
}
```

### 4. Additional Properties Not Allowed

```json
// WRONG: 'name' is not in the schema
{
    "description": "Value must be positive",
    "name": "positive_check"
}

// CORRECT: Only allowed fields
{
    "description": "Value must be positive"
}
```

### 5. Invalid Type Value

```json
// WRONG: 'precondition' is not a valid type
{
    "description": "Value must be positive",
    "type": "precondition"
}

// CORRECT: Valid type values
{
    "description": "Value must be positive",
    "type": "assessment"
}
```

## Troubleshooting Checklist

1. **Check required fields**: The `description` field must be present and non-empty
2. **Check field names**: Only `description`, `type`, `target`, `target_type_spec`, `notation`, and `specification` are allowed
3. **Check field types**: All fields must be strings
4. **Check type values**: `type` must be one of: `assessment`, `state_change`, `query`, `safety_rule`, `value`, `let`
5. **Remove extra fields**: Any field not in the schema will cause a violation
6. **Check nesting**: Ensure the logic object is at the correct level in the parent file

## Primed Values and the Target Field

The `target` field separates the left-hand side of an assignment from the specification, so the specification contains only the right-hand-side value expression.

For example, for the TLA+ expression `self.log' = Append(self.log, amount)`:
- `target` = `"log"` (the attribute being set)
- `specification` = `"Append(self.log, amount)"` (just the RHS)

**Primed values in specifications:**
- **`state_change`** (guarantees): NO primes — the target captures the LHS
- **`query`**: NO primes — the target captures the output name
- **`assessment`** (requires): NO primes — preconditions check current state only
- **`safety_rule`**: YES, primes ARE allowed (e.g., `"inventory' >= 0"`) to express constraints over both prior and new state
- **`value`**, **`let`**: NO primes

## Valid Examples

### Minimal Valid Logic Object

```json
{"description": "Order total must be positive"}
```

### Guarantee (state_change) — Target Separates LHS

```json
{
    "description": "The log is updated with the new amount",
    "target": "log",
    "notation": "tla_plus",
    "specification": "Append(self.log, amount)"
}
```

### Safety Rule — Primed Values Allowed

```json
{
    "description": "Inventory cannot go negative",
    "notation": "tla_plus",
    "specification": "inventory' >= 0"
}
```

### Let Binding — Local Variable Definition

```json
{
    "type": "let",
    "description": "Compute the threshold value",
    "target": "threshold",
    "target_type_spec": "Int",
    "notation": "tla_plus",
    "specification": "10"
}
```

### Assessment (Require) — No Primes

```json
{
    "description": "User must be authenticated",
    "notation": "tla_plus",
    "specification": "user.authenticated = TRUE"
}
```

## Related Errors

- **E14001**: Description field is missing
- **E14002**: Description field is empty
- **E14003**: JSON syntax is invalid
