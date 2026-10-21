# Subdomain Association Problem Key Required (E17101)

A subdomain association file is missing the required `problem_subdomain_key` field.

## Where Subdomain Associations Appear

Subdomain associations live at domain level:

```
domains/{domain}/subdomain_associations/{problem}.{solution}.subdomain_assoc.json
```

## How to Fix

Add `problem_subdomain_key` referencing an existing subdomain in the same domain.