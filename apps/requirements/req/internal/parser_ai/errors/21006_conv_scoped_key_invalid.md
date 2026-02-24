# Conversion Scoped Key Invalid (E21006)

A scoped class key reference has an invalid format during model conversion.

## What Went Wrong

A class key reference in an association or other cross-reference uses an incorrect format for its scope level. Scoped keys must follow specific formats depending on where the association is defined.

## Expected Formats

| Association Level | Key Format | Example |
|-------------------|------------|---------|
| Domain-level | `subdomain/class` | `"default/order"` |
| Model-level | `domain/subdomain/class` | `"sales/default/order"` |

## How to Fix

Check the `from_class_key` and `to_class_key` values in your association file:

### Domain-Level Association
```json
// File: domains/sales/associations/order_to_product.assoc.json
{
    "name": "Order Products",
    "from_class_key": "default/order",
    "to_class_key": "catalog/product",
    "from_multiplicity": "1",
    "to_multiplicity": "1..*"
}
```

### Model-Level Association
```json
// File: associations/order_to_warehouse.assoc.json
{
    "name": "Order Fulfillment",
    "from_class_key": "sales/default/order",
    "to_class_key": "logistics/warehousing/warehouse",
    "from_multiplicity": "1",
    "to_multiplicity": "0..1"
}
```

## Common Mistakes

```json
// WRONG: Missing subdomain part in domain-level association
{
    "from_class_key": "order"
}
// Should be:
{
    "from_class_key": "default/order"
}

// WRONG: Too many parts in domain-level association
{
    "from_class_key": "sales/default/order"
}
// This format is for model-level associations only
```

## Related Errors

- **E21004**: Class not found during conversion
- **E11002**: Association from_class_key not found
- **E11003**: Association to_class_key not found
