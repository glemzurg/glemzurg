# Class Index Attribute Not Found (E11007)

A class index references an attribute that does not exist in the class.

## What Went Wrong

A `class.json` file has an `indexes` array that references an attribute key, but no attribute with that key exists in the class's `attributes` map.

## How Indexes Work

Indexes specify combinations of attributes that should be indexed for efficient lookup. Each entry in the `indexes` array is an array of attribute keys.

```json
{
    "name": "Book Order",
    "attributes": {
        "id": {"name": "ID", "data_type_rules": "int"},
        "status": {"name": "Status", "data_type_rules": "string"},
        "customer_id": {"name": "Customer ID", "data_type_rules": "int"}
    },
    "indexes": [
        ["id"],                    // Single-attribute index
        ["status", "customer_id"]  // Composite index
    ]
}
```

## How to Fix

### Option 1: Add the Missing Attribute

Add the attribute to the class:

```json
{
    "attributes": {
        "missing_attr": {
            "name": "Missing Attribute",
            "data_type_rules": "string"
        }
    }
}
```

### Option 2: Fix the Reference

Update the index to reference an existing attribute:

```json
{
    "indexes": [
        ["existing_attribute"]
    ]
}
```

### Option 3: Remove the Index

Remove the index entry that references the missing attribute:

```json
{
    "indexes": [
        ["id"]
    ]
}
```

## Troubleshooting Checklist

1. **Check spelling**: Ensure the attribute key in the index matches exactly
2. **Check attribute exists**: The attribute must be defined in the `attributes` map
3. **Check for typos**: Common issues include extra underscores or wrong casing

## Common Mistakes

```json
// WRONG: Using attribute name instead of key
{
    "attributes": {
        "customer_id": {"name": "Customer ID"}
    },
    "indexes": [
        ["Customer ID"]
    ]
}
// Should be:
{
    "indexes": [
        ["customer_id"]
    ]
}

// WRONG: Typo in attribute key
{
    "indexes": [
        ["cusotmer_id"]
    ]
}
```

## Related Errors

- **E5008**: Attribute name is empty
- **E5009**: Index is invalid
