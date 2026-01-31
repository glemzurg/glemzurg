# Subdomain Not Default (E11030)

A subdomain has a name other than "default". Currently, only the "default" subdomain is supported.

## What Went Wrong

The model contains a subdomain directory that is not named "default". For the current implementation, all classes must be organized within a subdomain named "default".

## File Location

The error message includes the path to the non-default subdomain:

```
domains/{domain}/subdomains/{non_default_name}/
```

## How to Fix

Merge the contents of the non-default subdomain into the "default" subdomain.

### Step 1: Identify the Non-Default Subdomain

The error message tells you which subdomain needs to be merged:
- Path: `domains/{domain}/subdomains/{subdomain_name}/`

### Step 2: Merge Contents into Default

Move all contents from the non-default subdomain into the "default" subdomain:

```bash
# Example: Merge "orders" subdomain into "default"
# Move classes
mv domains/order_fulfillment/subdomains/orders/classes/* \
   domains/order_fulfillment/subdomains/default/classes/

# Move associations
mv domains/order_fulfillment/subdomains/orders/associations/* \
   domains/order_fulfillment/subdomains/default/associations/

# Move generalizations (if any)
mv domains/order_fulfillment/subdomains/orders/generalizations/* \
   domains/order_fulfillment/subdomains/default/generalizations/
```

### Step 3: Delete the Non-Default Subdomain

After merging, delete the now-empty subdomain directory:

```bash
rm -r domains/order_fulfillment/subdomains/orders/
```

### Step 4: Create Default Subdomain if Missing

If the "default" subdomain doesn't exist yet, create it:

```bash
mkdir -p domains/{domain}/subdomains/default/classes
mkdir -p domains/{domain}/subdomains/default/associations
```

Create the subdomain.json file:

```json
{
    "name": "Default"
}
```

## Example

### Before (Invalid Structure)

```
domains/
└── order_fulfillment/
    ├── domain.json
    └── subdomains/
        ├── orders/              <- ERROR: Not named "default"
        │   ├── subdomain.json
        │   ├── classes/
        │   │   └── book_order/
        │   └── associations/
        └── shipping/            <- ERROR: Not named "default"
            ├── subdomain.json
            ├── classes/
            │   └── shipment/
            └── associations/
```

### After (Valid Structure)

```
domains/
└── order_fulfillment/
    ├── domain.json
    └── subdomains/
        └── default/             <- All content merged here
            ├── subdomain.json
            ├── classes/
            │   ├── book_order/
            │   └── shipment/
            └── associations/
```

## Why Only "default" Is Supported

### 1. Simplified Model Structure

The current implementation focuses on a flat class structure within each domain. Multiple subdomains add complexity that is not yet supported.

### 2. Future Compatibility

When multiple subdomain support is added, models using "default" will automatically work. Migrating from a single subdomain is straightforward.

### 3. Cross-Subdomain References

Supporting multiple subdomains requires handling cross-subdomain class references in associations and generalizations. This adds complexity to key scoping that is not yet implemented.

## If You Need Multiple Logical Groupings

If you want to logically group classes, consider using:

1. **Multiple Domains**: Create separate domains for truly distinct areas of functionality
2. **Naming Conventions**: Use class name prefixes to indicate grouping (e.g., `order_book_order`, `shipping_shipment`)
3. **Comments/Details**: Document the logical grouping in class and domain details

## Related Errors

- **E11019**: Domain has no subdomains (must have at least "default")
- **E4006**: Subdomain directory name is invalid

## Future Support

Multiple subdomain support may be added in a future version. When available:
- Subdomains will allow organizing classes into logical groups within a domain
- Cross-subdomain associations will use `subdomain/class` key format
- The "default" subdomain will remain supported for backward compatibility
