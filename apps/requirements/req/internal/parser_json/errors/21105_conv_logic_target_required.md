# Logic Target Required (E21105)

A logic item requires a `target` field but it is empty.

## What Went Wrong

Certain logic types require a `target` that names the attribute or variable being defined:

- `state_change` guarantees must target an attribute
- `query` guarantees must target a result variable
- `let` items must target a variable name
- `safety_rule` items must target an attribute

## How to Fix

Add a `target` value that names the attribute or variable this logic item defines or modifies.

## Related Errors

- **E21106**: Logic target must be empty for this type
- **E21107**: Logic target must not start with underscore
