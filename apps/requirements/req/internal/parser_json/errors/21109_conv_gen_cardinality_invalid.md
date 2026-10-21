# Generalization Cardinality Invalid (E21109)

A generalization (actor, class, or use case) does not have the correct number of superclasses or subclasses referencing it.

## What Went Wrong

Every generalization must have exactly one entity marked as its superclass and at least one entity marked as a subclass. The error message indicates which generalization has the wrong count.

## How to Fix

Check the actor, class, or use case entities that reference this generalization:

- Exactly **one** entity must set `superclass_of_key` to this generalization's key
- At least **one** entity must set `subclass_of_key` to this generalization's key

If a generalization has no superclass or subclass references, either add the missing references or remove the generalization.

## Related Errors

- **E10005**: Class generalization superclass_key required
- **E10006**: Class generalization subclass_keys required
- **E12005**: Actor generalization superclass_key required
