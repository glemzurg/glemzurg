# Class Generalization Superclass Is Subclass (E11014)

A class generalization lists its superclass as one of its subclasses.

## What Went Wrong

A class generalization file has the same class key appearing in both `superclass_key` and `subclass_keys`. A class cannot be both the parent and child in the same class generalization.

## Example of the Error

```json
{
    "name": "Medium",
    "superclass_key": "product",
    "subclass_keys": ["book", "product", "ebook"]
}
```

In this example, `product` appears as both the superclass and a subclass, which is invalid.

## How to Fix

Remove the superclass from the subclass list:

```json
{
    "name": "Medium",
    "superclass_key": "product",
    "subclass_keys": ["book", "ebook"]
}
```

## Troubleshooting Checklist

1. **Review the relationship**: Ensure you understand which class is the parent
2. **Check for typos**: Verify the correct class keys are used
3. **Review class generalization structure**: Superclass should be the more general concept

## Class Generalization Structure

```
        Product (superclass)
           /\
          /  \
         /    \
      Book   EBook (subclasses)
```

The superclass represents a more general concept that the subclasses specialize.

## Related Errors

- **E11005**: Superclass not found
- **E11006**: Subclass not found
- **E11015**: Duplicate subclass in subclass_keys
