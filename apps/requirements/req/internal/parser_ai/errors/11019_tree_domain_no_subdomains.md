# Domain Has No Subdomains (E11019)

Every domain must have at least one subdomain to organize its classes.

## What Went Wrong

A domain directory exists but its `subdomains/` directory is empty or does not contain any subdomain subdirectories with `subdomain.json` files.

## Context

Subdomains break down a domain into more focused areas. They provide a middle layer of organization between domains and classes.

```
your_model/
└── domains/
    └── orders/
        ├── domain.json
        └── subdomains/               <-- Must contain at least one subdomain
            └── fulfillment/          <-- Subdomain directory
                ├── subdomain.json    <-- Subdomain definition
                └── classes/
                    └── ...
```

## How to Fix

### Step 1: Create a Subdomain Directory

Create a directory under the domain's `subdomains/` folder:

```
domains/{domain_key}/subdomains/{subdomain_key}/
```

### Step 2: Create the Subdomain Definition

Add a `subdomain.json` file:

```json
{
    "name": "Fulfillment",
    "details": "Manages order fulfillment including packing, shipping, and delivery tracking"
}
```

### Step 3: Add Classes

Each subdomain needs at least 2 classes. Continue by creating classes in the subdomain.

## Subdomain Design Guidelines

### What Makes a Good Subdomain?

A subdomain should:
- Contain **closely related classes** that work together
- Have **minimal dependencies** on other subdomains
- Represent a **cohesive sub-area** of the domain
- Be small enough to understand quickly

### Subdomain Examples

Within an "Orders" domain:
- **Cart** - ShoppingCart, CartItem, Coupon
- **Checkout** - Order, Payment, ShippingAddress
- **Fulfillment** - Shipment, Package, TrackingEvent

Within a "Customers" domain:
- **Identity** - User, Credential, Session
- **Profile** - Profile, Preference, Address
- **Communication** - EmailPreference, Notification

### Naming Conventions

- Use lowercase with underscores: `order_processing`, `user_management`
- Be specific: `shipping` not `stuff`
- Avoid generic names: `core`, `common`, `utils`

## When to Create Multiple Subdomains

Create separate subdomains when:
1. Classes have **different lifecycles** (Cart items are temporary, Orders are permanent)
2. Classes are **accessed by different actors** (Admin vs Customer)
3. Classes represent **different stages** of a process (Draft vs Published)
4. You want to **isolate complexity** (Payment processing vs Order display)

## Single Subdomain Pattern

For simple domains, a single subdomain is acceptable:

```
domains/
└── simple_domain/
    ├── domain.json
    └── subdomains/
        └── core/              <-- Single subdomain
            ├── subdomain.json
            └── classes/
```

Name it something meaningful like `core`, `default`, or describe its purpose.

## Related Errors

- **E4001**: Subdomain name is required
- **E11020**: Subdomain has too few classes
- **E11021**: Subdomain has no associations
