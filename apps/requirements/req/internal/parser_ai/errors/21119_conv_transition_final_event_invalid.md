# Final Transition Event Must Be _delete (E21119)

After conversion, model validation found a final transition that does not use the `_delete` event.

## What Went Wrong

A class transition reaches the final pseudo-state (`ToStateKey` is nil) but the bound event name is not `_delete`.

## How to Fix

In `state_machine.json`, declare `events._delete` and set the finalization transition's `event_key` to `"_delete"`.

See **E11038** for the tree-level check with the same rule.