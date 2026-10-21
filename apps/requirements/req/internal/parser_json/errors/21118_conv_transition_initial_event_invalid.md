# Initial Transition Event Must Be _new (E21118)

After conversion, model validation found an initial transition that does not use the `_new` event.

## What Went Wrong

A class transition leaves the initial pseudo-state (`FromStateKey` is nil) but the bound event name is not `_new`.

## How to Fix

In `state_machine.json`, declare `events._new` and set the creation transition's `event_key` to `"_new"`.

See **E11037** for the tree-level check with the same rule.