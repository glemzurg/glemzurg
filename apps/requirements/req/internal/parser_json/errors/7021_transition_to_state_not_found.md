# Transition To-State Not Found (E7021)

A transition references a `to_state_key` that does not exist in the states map.

## How to Fix

Ensure `to_state_key` matches a key in the `states` map, or use `null` for a final transition.
