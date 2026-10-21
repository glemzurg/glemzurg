# Class Actor Not Found (E5007)

The class references an `actor_key` that does not match any defined actor.

## How to Fix

Ensure the `actor_key` in the class JSON matches an actor filename in `actors/`:

```json
{
    "name": "Order",
    "actor_key": "customer"
}
```

The actor file `actors/customer.actor.json` must exist.

## Related Errors

- **E11001**: Tree-level actor reference validation
