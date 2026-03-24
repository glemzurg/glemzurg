# Class Index Attribute Not Found (E5010)

A class index references an attribute key that does not exist in the class.

## How to Fix

Ensure every attribute key listed in the index exists in the class's `attributes` map:

```json
{
    "name": "Order",
    "attributes": {
        "order_date": { "name": "Order Date" },
        "customer_id": { "name": "Customer ID" }
    },
    "indexes": [
        ["customer_id", "order_date"]
    ]
}
```

## Related Errors

- **E5009**: Index format invalid
- **E11007**: Tree-level index attribute validation
