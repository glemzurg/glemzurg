# Class Generalization Superclass Not Found (E11005)

A class generalization references a superclass that does not exist in the subdomain.

## What Went Wrong

A class generalization file has a `superclass_key` field that references a class, but no class with that key exists in the same subdomain as the class generalization.

## How Class Generalizations Work

Class generalizations define inheritance relationships between classes. All classes referenced by a class generalization must exist within the same subdomain.

```
your_model/
└── domains/
    └── products/
        └── subdomains/
            └── default/
                ├── classes/
                │   ├── product/           <-- Superclass must exist here
                │   │   └── class.json
                │   ├── book/
                │   │   └── class.json
                │   └── ebook/
                │       └── class.json
                └── generalizations/
                    └── medium.gen.json    <-- References superclass "product"
```

## How to Fix

### Option 1: Create the Missing Superclass

Create the class directory and `class.json` file:

```
domains/{domain}/subdomains/{subdomain}/classes/{superclass_key}/
└── class.json
```

### Option 2: Fix the Reference

Update the class generalization to reference an existing class:

```json
{
    "name": "Medium",
    "superclass_key": "existing_class",
    "subclass_keys": ["book", "ebook"]
}
```

## Troubleshooting Checklist

1. **Check spelling**: Ensure the superclass key matches the class directory name exactly
2. **Check location**: The superclass must be in the same subdomain as the class generalization
3. **Check class exists**: Ensure the class directory contains a valid `class.json`

## Common Mistakes

```json
// WRONG: Using class name instead of key
{
    "superclass_key": "Product"
}
// Should be:
{
    "superclass_key": "product"
}

// WRONG: Including path in key
{
    "superclass_key": "classes/product"
}
```

## Related Errors

- **E11006**: Class generalization subclass not found
- **E11014**: Superclass cannot also be a subclass
- **E10005**: superclass_key field is required
