# Logic Type Invalid for Context (E21102)

A logic item has a `type` that is not allowed in its context.

## What Went Wrong

Each logic context (invariants, requires, guarantees, safety_rules, guards, global functions) accepts only specific logic types. The logic item's `type` does not match what is allowed.

## Allowed Types by Context

| Context | Allowed Types |
|---------|--------------|
| Invariants (model, class, attribute) | `assessment`, `let` |
| Action requires | `assessment`, `let` |
| Action guarantees | `state_change`, `let` |
| Action safety_rules | `safety_rule`, `let` |
| Query requires | `assessment`, `let` |
| Query guarantees | `query`, `let` |
| Guards | `assessment` |
| Global functions | `value` |
| Derivation policies | `value` |

## How to Fix

Change the `type` field in the logic item to match the context it appears in. For example, if this logic is in an invariant, use `assessment` or `let`.

## Related Errors

- **E21103**: Duplicate let target in logic list
