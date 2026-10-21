# Logic Description Empty (E14002)

A logic object has a `description` field that is empty or contains only whitespace.

## What Went Wrong

The parser found a `description` field in a logic object, but its value is either an empty string (`""`) or contains only whitespace characters. The description must contain at least one visible character.

## Where Logic Objects Appear

Logic objects are embedded sub-objects within action, query, state machine, class, and model JSON files. They are used for:

- **Requires**: Preconditions on actions and queries
- **Guarantees**: Postconditions on actions and queries
- **Safety rules**: Constraints on actions
- **Invariants**: Model-level invariants
- **Guards**: Conditions on state machine transitions
- **Derivation policies**: Rules for derived attributes

## How to Fix

Provide a meaningful, non-empty description for your logic object:

```json
{
    "description": "Order total must be positive"
}
```

## Invalid Examples

These values will all trigger this error:

```json
{"description": ""}              // Empty string
{"description": "   "}           // Spaces only
{"description": "\t"}            // Tab only
```

## Valid Examples

The description must contain at least one non-whitespace character:

```json
// Minimal valid logic object
{"description": "Order total must be positive"}

// Logic object with formal specification
{"description": "User must be authenticated", "notation": "tla_plus", "specification": "user.authenticated = TRUE"}
```

## Choosing a Good Description

A good logic description should:

1. **State the constraint clearly**: What must be true?
2. **Be concise**: Aim for a single sentence
3. **Use domain language**: Reference entities and attributes from your model
4. **Be unambiguous**: Another reader should understand the intent

### Good Description Examples

| Context | Description |
|---------|-------------|
| Action requires | `"Customer must have a valid shipping address"` |
| Action guarantees | `"Order status is set to confirmed"` |
| Safety rule | `"Order total must not exceed credit limit"` |
| Model invariant | `"Every order must belong to exactly one customer"` |
| Guard | `"Payment has been received"` |

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `description` | string | **Yes** | `minLength: 1` |
| `notation` | string | No | None |
| `specification` | string | No | None |

## Related Errors

- **E14001**: Logic description field is missing entirely
- **E14003**: Invalid JSON syntax in a logic object
- **E14004**: Logic object schema violation
