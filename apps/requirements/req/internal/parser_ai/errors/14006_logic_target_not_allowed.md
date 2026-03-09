# Logic Target Not Allowed (E14006)

A logic specification has a `target` field but its type does not support targets.

## How to Fix

Only `state_change` and `safety_rule` types support `target`. Remove `target` for `assessment` and `query` types.
