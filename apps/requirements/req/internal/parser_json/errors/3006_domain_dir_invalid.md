# Domain Directory Invalid (E3006)

A domain directory in AI JSON does not match the required layout.

## Common causes

1. The domain folder name does not match the key naming convention.
2. The domain contains `classes/` or `use_cases/` at the domain root (the human YAML layout). AI JSON requires `subdomains/{subdomain}/classes/` and `subdomains/{subdomain}/use_cases/` instead. For a single-subdomain domain, use `subdomains/default/`.

## How to Fix

Domain directories must be lowercase snake_case: `^[a-z][a-z0-9]*(_[a-z0-9]+)*$`

Move domain-root model content under an explicit subdomain directory, typically `subdomains/default/` when the domain has only one subdomain.
