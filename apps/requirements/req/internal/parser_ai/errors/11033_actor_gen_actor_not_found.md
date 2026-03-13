# ErrTreeActorGenActorNotFound (11033)

Description

- An actor generalization (`*.agen.json`) references an actor key (superclass or subclass) that does not exist in the model's `actors/` directory.

Cause

- The `superclass_key` or one of the `subclass_keys` refers to a missing actor (no `actors/<key>.actor.json` file was read into the model).

Why this matters

- Actor generalizations declare type hierarchies for actors. Missing actor references mean the generalization is invalid and tools or AI cannot reason correctly about actor relationships.

How to fix

- Create the missing actor JSON file under `actors/` using the referenced key, or correct the key in the generalization file.
- Ensure keys are the directory/file-derived keys (snake_case, matching filenames).

Example

- Invalid: `superclass_key: customer` when no `actors/customer.actor.json` exists.
- Fix: add `actors/customer.actor.json` or change the key to the existing actor.

Important: Resolve each issue individually

Do not attempt to fix multiple actor or generalization errors in a single bulk operation. Each reference is distinct — a bulk change will often fix some parts correctly while breaking others, creating more errors to fix later. Address each error one at a time, verifying correctness before moving to the next.
