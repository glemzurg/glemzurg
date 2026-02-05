# Multiple Subdomains Cannot Include Default (E11031)

A domain has multiple subdomains, but one of them is named "default".

## What Went Wrong

When a domain contains more than one subdomain, none of them should be named "default". The "default" name is reserved for domains that have not yet been split into multiple logical groupings. Once you have multiple subdomains, each should have a descriptive name reflecting its purpose.

## File Location

The error message points to the "default" subdomain that needs to be renamed:

```
domains/{domain}/subdomains/default/
```

## How to Fix

Rename the "default" subdomain to something that describes its purpose.

### Step 1: Identify the Purpose

Look at the classes in the "default" subdomain and determine what they represent. Common patterns:
- **Core functionality**: `core`, `foundation`, `base`
- **Primary workflow**: `ordering`, `checkout`, `processing`
- **Main entities**: `products`, `customers`, `inventory`

### Step 2: Rename the Subdomain Directory

```bash
# Example: Rename "default" to "ordering" based on its classes
mv domains/order_fulfillment/subdomains/default \
   domains/order_fulfillment/subdomains/ordering
```

### Step 3: Update the subdomain.json Name

Edit `domains/{domain}/subdomains/{new_name}/subdomain.json`:

```json
{
    "name": "Ordering",
    "details": "Handles order creation, modification, and status tracking"
}
```

## Example

### Before (Invalid Structure)

```
domains/
└── order_fulfillment/
    ├── domain.json
    └── subdomains/
        ├── default/           <- ERROR: Cannot use "default" with multiple subdomains
        │   ├── subdomain.json
        │   └── classes/
        │       ├── order/
        │       └── order_line/
        └── shipping/
            ├── subdomain.json
            └── classes/
                ├── shipment/
                └── tracking/
```

### After (Valid Structure)

```
domains/
└── order_fulfillment/
    ├── domain.json
    └── subdomains/
        ├── ordering/          <- Renamed to describe its purpose
        │   ├── subdomain.json
        │   └── classes/
        │       ├── order/
        │       └── order_line/
        └── shipping/
            ├── subdomain.json
            └── classes/
                ├── shipment/
                └── tracking/
```

## Why This Convention Exists

### 1. Meaningful Organization

When you have multiple subdomains, you've made a conscious decision to organize classes into logical groups. Each group should have a name that communicates its purpose. "Default" conveys no meaning.

### 2. Clear Communication

Subdomain names appear in documentation, breadcrumbs, and navigation. Names like "Cart", "Checkout", and "Fulfillment" help readers understand the system organization. "Default" tells them nothing.

### 3. Avoid Confusion

If "default" were allowed alongside named subdomains, it would be unclear what belongs in "default" versus the named subdomains. Is it a catch-all? Legacy code? The naming convention eliminates this ambiguity.

## Choosing Good Subdomain Names

Good subdomain names are:
- **Descriptive**: Reflect the classes they contain
- **Concise**: 1-2 words, lowercase with underscores
- **Domain-specific**: Use terminology from the problem domain

### Examples by Domain

**Orders Domain:**
- `cart` - ShoppingCart, CartItem, Coupon
- `checkout` - Order, Payment, Address
- `fulfillment` - Shipment, Package, Tracking

**Users Domain:**
- `identity` - User, Credential, Session
- `profile` - Profile, Preference, Settings
- `permissions` - Role, Permission, AccessGrant

**Products Domain:**
- `catalog` - Product, Category, Brand
- `inventory` - Stock, Warehouse, Location
- `pricing` - Price, Discount, Promotion

## Subdomain Size Guidelines

When splitting a domain:
- Target **20-40 classes** per subdomain
- Group classes that **work together** frequently
- Minimize **cross-subdomain associations**

## Related Errors

- **E11019**: Domain has no subdomains
- **E11030**: Single subdomain must be named "default"
- **E11020**: Subdomain has too few classes
