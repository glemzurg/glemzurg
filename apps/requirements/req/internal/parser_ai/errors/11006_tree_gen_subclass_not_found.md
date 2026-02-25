# Class Generalization Subclass Not Found (E11006)

A class generalization references a subclass that does not exist in the subdomain.

## What Went Wrong

A class generalization file has a `subclass_keys` array that includes a class key, but no class with that key exists in the same subdomain as the class generalization.

## How Class Generalizations Work

Class generalizations define inheritance relationships between classes. All classes referenced by a class generalization (both superclass and subclasses) must exist within the same subdomain.

```
your_model/
└── domains/
    └── products/
        └── subdomains/
            └── default/
                ├── classes/
                │   ├── product/
                │   │   └── class.json
                │   ├── book/              <-- Each subclass must exist
                │   │   └── class.json
                │   └── ebook/             <-- Each subclass must exist
                │       └── class.json
                └── generalizations/
                    └── medium.gen.json    <-- References subclasses
```

## How to Fix

### Option 1: Create the Missing Subclass

Create the class directory and `class.json` file:

```
domains/{domain}/subdomains/{subdomain}/classes/{subclass_key}/
└── class.json
```

### Option 2: Fix the Reference

Update the class generalization to reference existing classes only:

```json
{
    "name": "Medium",
    "superclass_key": "product",
    "subclass_keys": ["book", "existing_class"]
}
```

### Option 3: Remove the Reference

Remove the missing subclass from the array:

```json
{
    "name": "Medium",
    "superclass_key": "product",
    "subclass_keys": ["book"]
}
```

## Troubleshooting Checklist

1. **Check spelling**: Ensure each subclass key matches its class directory name exactly
2. **Check location**: All subclasses must be in the same subdomain as the class generalization
3. **Check all entries**: Each entry in `subclass_keys` must reference an existing class

## Common Mistakes

```json
// WRONG: Typo in subclass key
{
    "subclass_keys": ["book", "e-book"]
}
// Should be:
{
    "subclass_keys": ["book", "ebook"]
}
```

## Related Errors

- **E11005**: Class generalization superclass not found
- **E11015**: Duplicate subclass in class generalization
- **E10006**: subclass_keys field is required
