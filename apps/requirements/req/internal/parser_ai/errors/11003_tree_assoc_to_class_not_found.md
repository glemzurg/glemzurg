# Association To Class Not Found (E11003)

The `to_class_key` in an association references a class that does not exist.

## How to Fix

### Step 1: Determine the correct association level

Decide which scope this association belongs to based on where the two classes live:

| Both classes in... | Association level | File location |
|--------------------|-------------------|---------------|
| Same subdomain | Subdomain | `domains/{dom}/subdomains/{sub}/associations/` |
| Same domain, different subdomains | Domain | `domains/{dom}/associations/` |
| Different domains | Model | `associations/` |

If the association file is at the wrong level, move it to the correct location.

### Step 2: Fix the `to_class_key` format for that level

| Level | Key format | Example |
|-------|-----------|---------|
| Subdomain | `class_name` | `line_item` |
| Domain | `subdomain/class` | `shipping/shipment` |
| Model | `domain/subdomain/class` | `inventory/stock/item` |

### Step 3: Create the class if it does not exist

Only after confirming the association is at the right level and the key format is correct, create the missing class file if needed:

```
domains/{domain}/subdomains/{subdomain}/classes/{class_name}/class.json
```

## Related Errors

- **E11002**: `from_class_key` not found (same resolution steps)
- **E11004**: `association_class_key` not found
