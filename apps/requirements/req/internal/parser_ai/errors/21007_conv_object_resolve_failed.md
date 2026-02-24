# Conversion Object Resolve Failed (E21007)

Failed to resolve an object reference when converting a scenario step.

## What Went Wrong

A scenario step references an event or query, but the step has no `from_object_key` or `to_object_key` to determine which class the event or query belongs to. Without an object reference, the system cannot resolve the class key needed to create the event or query key.

## How to Fix

Ensure that scenario steps referencing events or queries also include at least one object reference:

```json
{
    "step_type": "leaf",
    "leaf_type": "synchronous",
    "from_object_key": "customer_obj",
    "to_object_key": "order_obj",
    "event_key": "place_order",
    "description": "Customer places an order"
}
```

## Rules

- Steps with `event_key` must have a `to_object_key` or `from_object_key`
- Steps with `query_key` must have a `to_object_key` or `from_object_key`
- The event/query is looked up on the class of the target object

## Related Errors

- **E11035**: Scenario step event not found on referenced class
- **E11036**: Scenario step query not found on referenced class
- **E11034**: Scenario step references object key that doesn't exist
