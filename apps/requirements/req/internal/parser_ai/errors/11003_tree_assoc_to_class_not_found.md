# Association To Class Not Found (E11003)

An association references a target class (`to_class_key`) that does not exist.

## What Went Wrong

An association file references a class in its `to_class_key` field, but that class does not exist at the expected location. The expected location depends on the association's scope:

- **Subdomain-level**: Class must exist within the same subdomain
- **Domain-level**: Class must exist within the specified subdomain of the same domain
- **Model-level**: Class must exist within the specified domain/subdomain

## Key Formats by Scope

| Scope | Key Format | Example |
|-------|-----------|---------|
| Subdomain | `class_name` | `line_item` |
| Domain | `subdomain/class` | `shipping/shipment` |
| Model | `domain/subdomain/class` | `inventory/stock/inventory_item` |

## How to Fix

### Option 1: Create the Missing Class

Create the class directory and `class.json` file at the appropriate location:

```
domains/{domain}/subdomains/{subdomain}/classes/{class_name}/
└── class.json
```

### Option 2: Fix the Reference

Update the association to reference an existing class:

```json
{
    "name": "Order Lines",
    "from_class_key": "book_order",
    "from_multiplicity": "1",
    "to_class_key": "existing_class",
    "to_multiplicity": "*"
}
```

### Option 3: Fix the Key Format

Ensure the key format matches the association's scope.

## Troubleshooting Checklist

1. **Check spelling**: Ensure class names match exactly
2. **Check key format**: Ensure the scope matches the association's location
3. **Check class directory exists**: Classes are directories, not files
4. **Check class.json exists**: Each class directory must contain `class.json`

## Common Mistakes

```json
// WRONG: Wrong scope for subdomain-level association
{
    "to_class_key": "orders/line_item"
}
// Should be (for subdomain-level):
{
    "to_class_key": "line_item"
}
```

## Related Errors

- **E11002**: Association from_class_key not found
- **E11004**: Association association_class_key not found
- **E6006**: to_class_key field is required
