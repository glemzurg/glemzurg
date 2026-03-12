# Association Filename Invalid Component (E11028)

A component in an association filename is not valid snake_case.

## Component Rules

Each component (from_class, to_class, name, and optional subdomain/domain parts) must:
- Start with a lowercase letter (a-z)
- Contain only lowercase letters, digits (0-9), and underscores
- Not start/end with underscore or have consecutive underscores

## How to Fix

The error message identifies which component is invalid and its value. Rename the file so that component is valid snake_case.

```
# WRONG
BookOrder--line--orders.assoc.json
book-order--line--orders.assoc.json

# CORRECT
book_order--line--orders.assoc.json
```

## Filename Formats

- Subdomain level: `{from_class}--{to_class}--{name}.assoc.json`
- Domain level: `{from_sub}.{from_class}--{to_sub}.{to_class}--{name}.assoc.json`
- Model level: `{from_dom}.{from_sub}.{from_class}--{to_dom}.{to_sub}.{to_class}--{name}.assoc.json`

## Related Errors

- **E11026**: Key invalid format
- **E11027**: Association filename wrong number of parts
