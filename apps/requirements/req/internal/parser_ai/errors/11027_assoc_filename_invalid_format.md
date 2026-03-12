# Association Filename Invalid Format (E11027)

An association filename does not have the required three parts separated by `--`.

## How to Fix

### Step 1: Determine the correct association level

Decide where this association belongs based on where the two classes live:

| Both classes in... | Association level | File location |
|--------------------|-------------------|---------------|
| Same subdomain | Subdomain | `domains/{dom}/subdomains/{sub}/class_associations/` |
| Same domain, different subdomains | Domain | `domains/{dom}/class_associations/` |
| Different domains | Model | `class_associations/` |

### Step 2: Use the correct filename format for that level

All formats have exactly three `--`-separated parts: `from--to--name.assoc.json`

| Level | Format |
|-------|--------|
| Subdomain | `{from_class}--{to_class}--{name}.assoc.json` |
| Domain | `{from_sub}.{from_class}--{to_sub}.{to_class}--{name}.assoc.json` |
| Model | `{from_dom}.{from_sub}.{from_class}--{to_dom}.{to_sub}.{to_class}--{name}.assoc.json` |

### Step 3: Create any missing classes

Only after the association is at the correct level with the correct filename format, create any referenced classes that do not yet exist.

## Common Mistakes

- Using single hyphens (`-`) instead of double hyphens (`--`) between parts
- Wrong number of parts (must be exactly 3)
- Using the wrong depth for the level (e.g., domain-level format in a subdomain directory)

## Related Errors

- **E11026**: Key invalid format
- **E11028**: Association filename invalid component (bad snake_case)
