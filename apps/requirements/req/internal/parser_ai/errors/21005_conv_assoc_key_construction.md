# Conversion Association Key Construction Failed (E21005)

A class association key could not be constructed during model conversion.

## What Went Wrong

When creating the internal key for a class association, the key construction failed. This can happen when:

- The association name is empty or whitespace-only
- The from/to class keys are not properly formed class keys
- For domain-level associations, both classes are in the same subdomain (should use subdomain-level instead)
- For model-level associations, both classes are in the same domain (should use domain-level instead)

## How to Fix

### Check Association Name
Every association must have a non-empty `name`:
```json
{
    "name": "Order Lines",
    "from_class_key": "order",
    "to_class_key": "line_item",
    "from_multiplicity": "1",
    "to_multiplicity": "1..*"
}
```

### Check Association Level
Place your association file at the correct level:

| Scenario | Level | Directory |
|----------|-------|-----------|
| Both classes in same subdomain | Subdomain | `domains/{d}/subdomains/{s}/associations/` |
| Classes in different subdomains, same domain | Domain | `domains/{d}/associations/` |
| Classes in different domains | Model | `associations/` |

## Related Errors

- **E6001**: Association name required
- **E6013**: Association filename invalid
- **E21004**: Class not found during conversion
