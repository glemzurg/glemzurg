# Class Actor Not Found (E11001)

A class references an actor that does not exist in the model.

## What Went Wrong

A `class.json` file contains an `actor_key` field that references an actor name, but no actor with that key exists in the model's `actors/` directory.

## Context

Classes can optionally specify an `actor_key` to indicate which actor primarily interacts with that class. This creates a cross-reference from the class to an actor defined elsewhere in the model.

```
your_model/
├── model.json
├── actors/
│   ├── customer.actor.json    <-- Actor files must exist here
│   └── admin.actor.json
└── domains/
    └── orders/
        └── subdomains/
            └── default/
                └── classes/
                    └── book_order/
                        └── class.json    <-- This file references a missing actor
```

## How to Fix

### Option 1: Create the Missing Actor

Add an actor file with the referenced key:

```
actors/{actor_key}.actor.json
```

With content like:

```json
{
    "name": "Customer",
    "type": "person",
    "details": "A user who places orders"
}
```

### Option 2: Fix the Reference

Update the `class.json` to reference an existing actor:

```json
{
    "name": "Book Order",
    "actor_key": "customer",
    "details": "..."
}
```

### Option 3: Remove the Reference

If no actor is appropriate, remove the `actor_key` field entirely:

```json
{
    "name": "Book Order",
    "details": "..."
}
```

## Troubleshooting Checklist

1. **Check spelling**: Ensure the actor key in `class.json` matches the actor filename exactly
2. **Check file extension**: Actor files must end with `.actor.json`
3. **Check actor location**: Actors must be in the model's `actors/` directory
4. **Check case sensitivity**: Filenames and keys are case-sensitive

## Common Mistakes

```json
// WRONG: Typo in actor_key
{
    "name": "Book Order",
    "actor_key": "custmer"
}

// WRONG: Using actor name instead of key
{
    "name": "Book Order",
    "actor_key": "Customer"
}
```

## Related Errors

- **E2001**: Actor name is required
- **E2006**: Actor schema violation
