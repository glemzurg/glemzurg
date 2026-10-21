# Duplicate Guarantee Target (E21104)

Two or more guarantee logic items in the same action or query target the same attribute.

## What Went Wrong

Each guarantee in an action or query must target a different attribute. The same attribute was targeted by multiple guarantees.

## How to Fix

Ensure each guarantee has a unique `target` (attribute name). If multiple changes apply to the same attribute, combine them into a single guarantee.

## Related Errors

- **E21103**: Duplicate let target
- **E21112**: Guarantee targets an attribute that doesn't exist
