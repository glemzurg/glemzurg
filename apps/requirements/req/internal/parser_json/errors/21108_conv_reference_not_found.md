# Cross-Reference Not Found (E21108)

An entity references another entity that does not exist in the model.

## What Went Wrong

During model validation, a reference to another entity (class, actor, generalization, state, event, guard, action, domain, use case, or scenario) could not be resolved. The referenced key does not match any defined entity.

## How to Fix

Check the error message for the specific reference that failed. Common causes:

- Misspelled key name
- Referenced entity was not created
- Referenced entity is in a different scope than expected

Verify that the referenced entity exists and that the key matches exactly.

## Related Errors

- **E11002**: Association from_class_key not found (tree-level)
- **E11003**: Association to_class_key not found (tree-level)
- **E11008**: Transition references non-existent state (tree-level)
