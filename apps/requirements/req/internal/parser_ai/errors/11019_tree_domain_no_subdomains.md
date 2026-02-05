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
- Contain between **20-40 classes** for optimal manageability

### Subdomain Size Guidelines

Subdomains should be sized appropriately:

- **Minimum**: At least 2 classes (a single class cannot have relationships)
- **Target**: 20-40 classes for a well-organized subdomain
- **Maximum**: When approaching 40+ classes, consider splitting

When a subdomain grows beyond 40 classes, it becomes difficult to understand and maintain. Split it into multiple subdomains based on cohesive groupings of classes.

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
- For simple domains, `core` or `default` are acceptable names

## When to Create Multiple Subdomains

Create separate subdomains when:
1. A subdomain exceeds **40 classes** — split based on cohesive groupings
2. Classes have **different lifecycles** (Cart items are temporary, Orders are permanent)
3. Classes are **accessed by different actors** (Admin vs Customer)
4. Classes represent **different stages** of a process (Draft vs Published)
5. You want to **isolate complexity** (Payment processing vs Order display)

## Starting with a Single Subdomain

For new domains, start with a single subdomain and split when needed:

```
domains/
└── simple_domain/
    ├── domain.json
    └── subdomains/
        └── core/              <-- Start with one subdomain
            ├── subdomain.json
            └── classes/
```

As the domain grows beyond 20-40 classes, identify cohesive groups and split into multiple subdomains.

## Related Errors

- **E4001**: Subdomain name is required
- **E11020**: Subdomain has too few classes
- **E11021**: Subdomain has no associations
