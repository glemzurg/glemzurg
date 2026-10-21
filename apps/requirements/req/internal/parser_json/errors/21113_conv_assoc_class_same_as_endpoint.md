# Association Class Same as Endpoint (E21113)

An association's `association_class_key` is the same as its `from_class_key` or `to_class_key`.

## What Went Wrong

An association class is a separate class that holds attributes about the relationship itself. It cannot be the same class as either endpoint of the association.

## How to Fix

The `association_class_key` must reference a distinct class. Create a separate class for the association's attributes, or remove the `association_class_key` if the relationship does not need its own attributes.

## Related Errors

- **E6011**: Association class not found
