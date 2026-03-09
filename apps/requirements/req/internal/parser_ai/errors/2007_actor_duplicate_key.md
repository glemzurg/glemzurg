# Actor Duplicate Key (E2007)

Two actor files resolve to the same key. Actor keys are derived from filenames.

## How to Fix

Each actor file must have a unique filename. Rename or remove the duplicate:

```
actors/
├── customer.actor.json       <- key: customer
└── customer.actor.json       <- DUPLICATE: same key
```

## Related Errors

- **E2001**: Actor name required
- **E2005**: Invalid JSON in actor file
