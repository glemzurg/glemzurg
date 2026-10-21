# Logic Type Required (E14008)

The `type` field is missing from a logic specification object.

## What Went Wrong

Every logic object must have a `type` field that declares its kind. The type determines how the logic is interpreted and validated.

## Valid Types by Context

| Context | Valid Types |
|---|---|
| Action requires | `assessment`, `let` |
| Action guarantees | `state_change`, `let` |
| Action safety_rules | `safety_rule`, `let` |
| Query requires | `assessment`, `let` |
| Query guarantees | `query`, `let` |
| Invariants (class, attribute, model) | `assessment`, `let` |
| Guard logic | `assessment` |
| Global function logic | `value` |
| Derivation policy | `value` |

## How to Fix

Add the appropriate `type` field based on the context where the logic appears:

```json
{
    "type": "assessment",
    "description": "User must be active",
    "notation": "tla_plus",
    "specification": "self.is_active = TRUE"
}
```

### Requires (preconditions)

Use `"type": "assessment"` for boolean checks, or `"type": "let"` for local variable definitions:

```json
{
    "type": "assessment",
    "description": "Order total must be positive",
    "notation": "tla_plus",
    "specification": "self.total > 0"
}
```

### Guarantees (postconditions in actions)

Use `"type": "state_change"` for attribute assignments, or `"type": "let"` for local variables:

```json
{
    "type": "state_change",
    "description": "Set order status to confirmed",
    "target": "status",
    "notation": "tla_plus",
    "specification": "\"CONFIRMED\""
}
```

### Guarantees (postconditions in queries)

Use `"type": "query"` for return values, or `"type": "let"` for local variables:

```json
{
    "type": "query",
    "description": "Return the order total",
    "target": "result",
    "notation": "tla_plus",
    "specification": "self.total"
}
```

## Related Errors

- **E14004**: Logic JSON does not match schema
- **E14005**: Logic target is required for state_change/query/let types
- **E14006**: Logic target must be empty for assessment/safety_rule/value types
