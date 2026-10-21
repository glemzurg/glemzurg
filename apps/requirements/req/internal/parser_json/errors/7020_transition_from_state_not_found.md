# Transition From-State Not Found (E7020)

A transition references a `from_state_key` that does not exist in the states map.

## How to Fix

Ensure `from_state_key` matches a key in the `states` map, or use `null` for an initial transition.
