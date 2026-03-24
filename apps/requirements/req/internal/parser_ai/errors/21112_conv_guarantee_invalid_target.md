# Guarantee Targets Non-Existent Attribute (E21112)

A guarantee's `target` references an attribute that does not exist on the class.

## What Went Wrong

Action and query guarantees with type `state_change` or `query` must target an attribute that is defined in the class's `attributes` map. The target name does not match any attribute key.

## How to Fix

Check that the guarantee's `target` matches an attribute key defined in `class.json`. The target must use the attribute's key (the map key in `attributes`), not its display name.

## Related Errors

- **E21104**: Duplicate guarantee target
- **E21105**: Logic target required
