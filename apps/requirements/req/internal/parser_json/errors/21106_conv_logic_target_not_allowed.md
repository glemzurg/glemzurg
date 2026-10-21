# Logic Target Must Be Empty (E21106)

A logic item has a `target` field set, but this logic type does not accept targets.

## What Went Wrong

The `assessment` logic type does not use a `target`. Assessment logic evaluates a condition without modifying any specific attribute or variable.

## How to Fix

Remove the `target` field from this logic item, or change the `type` to one that accepts targets.

## Related Errors

- **E21105**: Logic target is required
- **E21102**: Logic type invalid for context
