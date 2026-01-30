# Model Has No Domains (E11018)

The model must have at least one domain defined to organize the system's functionality.

## What Went Wrong

The model's `domains/` directory is empty or does not contain any domain subdirectories with `domain.json` files. Every model needs at least one domain to structure the application's concepts.

## Context

Domains are high-level subject areas that group related functionality. They help organize complex systems into manageable chunks.

```
your_model/
├── model.json
├── actors/
│   └── customer.actor.json
└── domains/                      <-- This directory must contain at least one domain
    └── orders/                   <-- Domain directory
        ├── domain.json           <-- Domain definition file
        └── subdomains/
            └── ...
```

## How to Fix

### Step 1: Create a Domain Directory

Create a directory under `domains/` with your domain name:

```
domains/{domain_key}/
```

### Step 2: Create the Domain Definition

Add a `domain.json` file in the domain directory:

```json
{
    "name": "Orders",
    "details": "Handles all order-related functionality including cart management, checkout, and order fulfillment"
}
```

### Step 3: Add Subdomains

Each domain needs at least one subdomain. Continue with creating subdomains inside your domain.

## Domain Design Guidelines

### What Makes a Good Domain?

A domain should:
- Represent a **cohesive business area** (Orders, Inventory, Customers)
- Be **relatively independent** from other domains
- Have **clear boundaries** with other domains
- Contain related concepts that change together

### Common Domain Examples

E-commerce system:
- **Orders** - Shopping cart, checkout, order tracking
- **Inventory** - Products, stock levels, warehouses
- **Customers** - User accounts, preferences, addresses
- **Payments** - Transactions, refunds, billing

Healthcare system:
- **Patients** - Patient records, demographics
- **Scheduling** - Appointments, availability
- **Clinical** - Diagnoses, treatments, prescriptions
- **Billing** - Claims, insurance, payments

### Anti-patterns to Avoid

- **Too broad**: A single "Application" domain that contains everything
- **Too narrow**: Domains for individual features (use subdomains instead)
- **Technical splits**: Domains based on technology rather than business concepts

## Realized vs Unrealized Domains

Domains can be marked as `realized` (true/false):
- **Realized**: This system implements this domain
- **Unrealized**: Domain exists conceptually but is implemented elsewhere

```json
{
    "name": "Payments",
    "details": "Payment processing handled by Stripe",
    "realized": false
}
```

## Related Errors

- **E3001**: Domain name is required
- **E11019**: Domain has no subdomains
