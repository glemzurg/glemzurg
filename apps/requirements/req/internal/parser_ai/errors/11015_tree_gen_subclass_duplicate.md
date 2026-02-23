# Class Generalization Subclass Duplicate (E11015)

A class generalization lists the same class multiple times in its subclasses.

## What Went Wrong

A class generalization file has the same class key appearing more than once in the `subclass_keys` array. Each subclass should only be listed once.

## Example of the Error

```json
{
    "name": "Medium",
    "superclass_key": "product",
    "subclass_keys": ["book", "ebook", "book"]
}
```

In this example, `book` appears twice in the subclass list.

## How to Fix

Remove the duplicate entry:

```json
{
    "name": "Medium",
    "superclass_key": "product",
    "subclass_keys": ["book", "ebook"]
}
```

## Troubleshooting Checklist

1. **Check for copy-paste errors**: Duplicates often occur from copying entries
2. **Review all entries**: Ensure each subclass is unique
3. **Check for similar names**: Ensure typos didn't create apparent duplicates

## Related Errors

- **E11006**: Subclass not found
- **E11014**: Superclass cannot also be a subclass
- **E10007**: subclass_keys cannot be empty
