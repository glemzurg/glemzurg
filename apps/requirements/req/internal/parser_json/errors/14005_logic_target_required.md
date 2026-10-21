# Logic Target Required (E14005)

A logic specification of type `state_change` or `safety_rule` is missing the required `target` field.

## How to Fix

Add a `target` field specifying the attribute being constrained:

```json
{
    "description": "Set status to active",
    "type": "state_change",
    "target": "status",
    "specification": "status' = \"active\""
}
```
