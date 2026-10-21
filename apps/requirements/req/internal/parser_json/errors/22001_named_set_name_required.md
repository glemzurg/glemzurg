# Named Set Name Required (E22001)

The named set JSON file is missing the required `name` field.

## How to Fix

Add a `name` field:

```json
{
    "name": "Order Statuses",
    "values": ["pending", "confirmed", "shipped"]
}
```
