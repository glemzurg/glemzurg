# Single Subdomain Must Be Named Default (E11030)

A domain has only one subdomain, but it is not named "default".

## What Went Wrong

When a domain contains exactly one subdomain, that subdomain must be named "default". This convention indicates that the domain has not yet been split into multiple logical groupings.

## File Location

The error message includes the path to the incorrectly named subdomain:

```
domains/{domain}/subdomains/{subdomain_name}/
```

## How to Fix

Rename the subdomain directory to "default".

### Step 1: Rename the Subdomain Directory

```bash
# Example: Rename "core" subdomain to "default"
mv domains/order_fulfillment/subdomains/core \
   domains/order_fulfillment/subdomains/default
```

### Step 2: Update the subdomain.json Name

Edit `domains/{domain}/subdomains/default/subdomain.json`:

```json
{
    "name": "Default",
    "details": "Primary subdomain for order fulfillment functionality"
}
```

## Example

### Before (Invalid Structure)

```
domains/
└── order_fulfillment/
    ├── domain.json
    └── subdomains/
        └── core/              <- ERROR: Single subdomain not named "default"
            ├── subdomain.json
            └── classes/
```

### After (Valid Structure)

```
domains/
└── order_fulfillment/
    ├── domain.json
    └── subdomains/
        └── default/           <- Renamed to "default"
            ├── subdomain.json
            └── classes/
```

## Why This Convention Exists

### 1. Clear Intent

The name "default" signals that the domain hasn't been intentionally split into multiple subdomains. It's a placeholder that will be renamed when the domain grows large enough to warrant multiple subdomains.

### 2. Consistent Starting Point

All domains start with a "default" subdomain. This provides a consistent structure across the model and makes it clear which domains have been intentionally organized into multiple subdomains versus those that haven't.

### 3. Migration Path

When a domain grows to 20-40+ classes, you'll split it into multiple meaningful subdomains. At that point, the "default" name goes away and is replaced with descriptive names like "cart", "checkout", and "fulfillment".

## When to Split into Multiple Subdomains

Keep the "default" subdomain until:
- The domain exceeds **20-40 classes**
- You can identify **2+ cohesive groupings** of classes
- Classes have **different lifecycles** or **different actors**

When you split, rename "default" to something meaningful and create additional subdomains. See E11031 for rules about multiple subdomains.

## Related Errors

- **E11019**: Domain has no subdomains
- **E11031**: Multiple subdomains cannot include one named "default"
- **E11020**: Subdomain has too few classes
