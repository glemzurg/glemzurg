# Conversion Class Not Found (E21004)

A class key referenced in an association could not be found during model conversion.

## What Went Wrong

During model conversion, an association references a class key (via `from_class_key`, `to_class_key`, or `association_class_key`) that does not match any class in the expected scope. This means the referenced class either doesn't exist or has an incorrect key.

## How to Fix

1. Check that the referenced class exists in the correct subdomain
2. Verify the class key matches the directory name exactly (case-sensitive, snake_case)
3. For domain-level associations, use `subdomain/class` format
4. For model-level associations, use `domain/subdomain/class` format

## Key Format by Association Level

| Level | Format | Example |
|-------|--------|---------|
| Subdomain | `class_name` | `"order"` |
| Domain | `subdomain/class` | `"default/order"` |
| Model | `domain/subdomain/class` | `"sales/default/order"` |

## Common Mistakes

```json
// WRONG: Class doesn't exist
{
    "from_class_key": "orders",
    "to_class_key": "line_item"
}
// Fix: Check that class directory "orders" exists. Maybe it should be "order"?

// WRONG: Wrong scope format for domain-level
{
    "from_class_key": "order",
    "to_class_key": "product"
}
// Fix: For domain-level associations, use "subdomain/class" format:
{
    "from_class_key": "default/order",
    "to_class_key": "catalog/product"
}
```

## Related Errors

- **E11002**: Association from_class_key not found (caught during tree validation)
- **E11003**: Association to_class_key not found (caught during tree validation)
- **E11004**: Association class key not found (caught during tree validation)
