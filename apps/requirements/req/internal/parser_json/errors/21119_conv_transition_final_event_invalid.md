# Final Transition Event Must Be _destroy (E21119)

After conversion, model validation found a final transition that does not use the `_destroy` event.

## What Went Wrong

A class transition reaches the final pseudo-state (`ToStateKey` is nil) but the bound event name is not `_destroy`.

## How to Fix

In `state_machine.json`, declare `events._destroy` and set the finalization transition's `event_key` to `"_destroy"`.

See **E11038** for the tree-level check with the same rule.