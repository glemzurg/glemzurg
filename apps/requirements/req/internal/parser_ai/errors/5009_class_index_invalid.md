# Class Index Invalid (E5009)

An index in the class contains an empty or whitespace-only attribute key.

## What Went Wrong

The class has an `indexes` array, and one of the indexes contains an attribute key that is either an empty string (`""`) or contains only whitespace characters. Every attribute key in an index must be a valid, non-empty reference.

## File Location

Class files are located within domain directories:

```
your_model/
├── model.json
└── order_management/           <-- Domain directory
    ├── domain.json
    └── order.class.json        <-- An index in this file has an invalid key
```

## How to Fix

Ensure every attribute key in each index is a valid, non-empty string:

```json
{
    "name": "Order",
    "attributes": {
        "order_date": {
            "name": "Order Date"
        }
    },
    "indexes": [
        ["order_date"]
    ]
}
```

## Understanding the Error Field

The error message includes the path to the problematic index entry:

| Error Field | Meaning |
|-------------|---------|
| `indexes[0][0]` | First index, first attribute key is invalid |
| `indexes[0][1]` | First index, second attribute key is invalid |
| `indexes[1][0]` | Second index, first attribute key is invalid |
| `indexes[2][1]` | Third index, second attribute key is invalid |

## Invalid Examples

```json
// WRONG: Empty attribute key in index
{
    "name": "Order",
    "attributes": {
        "order_date": { "name": "Order Date" }
    },
    "indexes": [
        [""]
    ]
}

// WRONG: Whitespace-only attribute key
{
    "name": "Order",
    "attributes": {
        "order_date": { "name": "Order Date" }
    },
    "indexes": [
        ["   "]
    ]
}

// WRONG: Empty key in compound index
{
    "name": "Order",
    "attributes": {
        "order_date": { "name": "Order Date" },
        "status": { "name": "Status" }
    },
    "indexes": [
        ["order_date", ""]
    ]
}

// WRONG: Multiple invalid keys
{
    "name": "Order",
    "attributes": {
        "order_date": { "name": "Order Date" }
    },
    "indexes": [
        ["order_date"],
        ["", "   "]
    ]
}
```

## Valid Examples

```json
// Single-attribute index
{
    "name": "Order",
    "attributes": {
        "order_number": { "name": "Order Number" }
    },
    "indexes": [
        ["order_number"]
    ]
}

// Compound index
{
    "name": "Order",
    "attributes": {
        "customer_id": { "name": "Customer ID" },
        "order_date": { "name": "Order Date" }
    },
    "indexes": [
        ["customer_id", "order_date"]
    ]
}

// Multiple indexes
{
    "name": "Order",
    "attributes": {
        "order_number": { "name": "Order Number" },
        "customer_id": { "name": "Customer ID" },
        "order_date": { "name": "Order Date" },
        "status": { "name": "Status" }
    },
    "indexes": [
        ["order_number"],
        ["customer_id", "order_date"],
        ["status"]
    ]
}
```

## Understanding Indexes

Indexes define which attribute combinations should be indexed for efficient lookups:

```json
{
    "indexes": [
        ["order_number"],              // Single-attribute index
        ["customer_id", "order_date"]  // Compound index (both attributes together)
    ]
}
```

Each index entry must reference an attribute key defined in the `attributes` map.

## Index Structure

```
indexes: [
    ["attr1"],           // First index (single attribute)
    ["attr2", "attr3"]   // Second index (compound)
]
  │
  ├── indexes[0] = ["attr1"]
  │       └── indexes[0][0] = "attr1"
  │
  └── indexes[1] = ["attr2", "attr3"]
          ├── indexes[1][0] = "attr2"
          └── indexes[1][1] = "attr3"
```

## Troubleshooting Checklist

1. **Check for empty strings**: Look for `""` in the index arrays
2. **Check for whitespace-only**: Look for `"   "` entries
3. **Verify attribute exists**: Each key should match an attribute in the `attributes` map
4. **Check for typos**: Attribute keys are case-sensitive

### Checking with Command Line

```bash
# View all indexes
cat order.class.json | jq '.indexes'

# Check for empty strings in indexes
cat order.class.json | jq '.indexes[][] | select(. == "" or (. | test("^\\s*$")))'
```

## Related Errors

- **E5004**: Schema violation (general)
- **E5008**: Attribute name is empty
- **E5010**: Index references non-existent attribute (reference validation)
