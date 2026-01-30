# Subdomain Has Too Few Classes (E11020)

Every subdomain must have at least 2 classes to represent meaningful relationships.

## What Went Wrong

A subdomain contains fewer than 2 classes. A single class in isolation cannot have relationships with other classes, making the subdomain incomplete.

## Context

Classes within a subdomain work together to model a cohesive concept. Associations between classes define how they relate to each other.

```
your_model/
└── domains/
    └── orders/
        └── subdomains/
            └── checkout/
                ├── subdomain.json
                └── classes/              <-- Must contain at least 2 classes
                    ├── order/
                    │   └── class.json
                    └── line_item/        <-- Second class needed
                        └── class.json
```

## How to Fix

### Step 1: Identify Related Classes

Think about what classes naturally group together:
- **Order** and **LineItem** (orders contain items)
- **User** and **Address** (users have addresses)
- **Product** and **Category** (products belong to categories)

### Step 2: Create Class Directories

For each class, create a directory:

```
domains/{domain}/subdomains/{subdomain}/classes/{class_key}/
```

### Step 3: Create Class Definitions

Add a `class.json` file for each class:

```json
{
    "name": "Order",
    "details": "Represents a customer's order containing one or more items",
    "attributes": {
        "order_number": {
            "name": "Order Number",
            "data_type_rules": "string: unique, auto-generated",
            "details": "Unique identifier displayed to customers"
        },
        "status": {
            "name": "Status",
            "data_type_rules": "enum: pending, confirmed, shipped, delivered",
            "details": "Current state of the order"
        }
    }
}
```

### Step 4: Define Relationships

Create associations to show how classes relate:

```
domains/{domain}/subdomains/{subdomain}/associations/order_has_line_items.assoc.json
```

```json
{
    "name": "Order Has Line Items",
    "from_class_key": "order",
    "from_multiplicity": "1",
    "to_class_key": "line_item",
    "to_multiplicity": "1..*",
    "details": "Each order contains at least one line item"
}
```

## Common Class Groupings

### Entities and Their Items
- Order → LineItem
- Invoice → InvoiceItem
- Cart → CartItem

### Parent-Child Relationships
- Category → Product
- Department → Employee
- Folder → Document

### User and Related Data
- User → Profile
- User → Address
- User → PaymentMethod

### Status and History
- Order → StatusChange
- Document → Revision
- Account → Transaction

## Design Considerations

### When Classes Belong Together

Classes should be in the same subdomain when they:
1. **Change together** - Modifying one often requires modifying the other
2. **Reference each other** - Have direct associations
3. **Share business rules** - Validations span both classes
4. **Belong to same aggregate** - One is the root, others are parts

### When to Split into Separate Subdomains

Consider separate subdomains when:
1. Classes have **different owners** - Different teams manage them
2. Classes have **different lifecycles** - One is long-lived, other is transient
3. Classes are **loosely coupled** - Only occasional references

## Related Errors

- **E5001**: Class name is required
- **E11021**: Subdomain has no associations
- **E11022**: Class has no attributes
