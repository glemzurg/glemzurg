# Logic Association Class Not Allowed (E14011)

`endpoint_selector` is only valid on `state_change` action guarantees that reify an association class.

## How to Fix

Use `"type": "state_change"` with `endpoint_selector` and a creation `specification`, or remove `endpoint_selector`.
