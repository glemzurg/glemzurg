# Subdomain Has No Associations (E11021)

Every subdomain must have at least one association to describe how its classes relate to each other.

## What Went Wrong

A subdomain has classes defined but no associations between them. Associations are essential for describing the relationships and cardinality between classes.

## Context

Associations define how classes relate to each other. Without associations, classes exist in isolation without any defined connections.

```
your_model/
└── domains/
    └── orders/
        └── subdomains/
            └── checkout/
                ├── subdomain.json
                ├── classes/
                │   ├── order/
                │   │   └── class.json
                │   └── line_item/
                │       └── class.json
                └── associations/          <-- Must contain at least one association
                    └── order_has_items.assoc.json
```

## How to Fix

### Step 1: Identify Relationships

Look at your classes and ask:
- Does one class **contain** another? (Order contains LineItems)
- Does one class **reference** another? (LineItem references Product)
- Is there a **many-to-many** relationship? (Product has many Categories)

### Step 2: Create Association File

Create an association file in the `associations/` directory:

```
{subdomain}/associations/{descriptive_name}.assoc.json
```

### Step 3: Define the Association

```json
{
    "name": "Order Has Line Items",
    "from_class_key": "order",
    "from_multiplicity": "1",
    "to_class_key": "line_item",
    "to_multiplicity": "1..*",
    "details": "Each order contains at least one line item; line items belong to exactly one order"
}
```

## Understanding Multiplicities

### Common Patterns

| From | To | Meaning |
|------|-----|---------|
| `1` | `*` | One-to-many (Order has many LineItems) |
| `1` | `1..*` | One-to-many, at least one required |
| `1` | `0..1` | One-to-zero-or-one (User has optional Profile) |
| `*` | `*` | Many-to-many (Student has many Courses) |
| `1` | `1` | One-to-one (User has exactly one Account) |

### Multiplicity Notation

- `1` - Exactly one
- `*` - Zero or more (unbounded)
- `0..1` - Zero or one (optional)
- `1..*` - One or more (at least one required)
- `2..5` - Between 2 and 5

## Association Examples

### Composition (Contains)
```json
{
    "name": "Order Contains Line Items",
    "from_class_key": "order",
    "from_multiplicity": "1",
    "to_class_key": "line_item",
    "to_multiplicity": "1..*",
    "details": "Line items cannot exist without their order"
}
```

### Reference (Links To)
```json
{
    "name": "Line Item References Product",
    "from_class_key": "line_item",
    "from_multiplicity": "*",
    "to_class_key": "product",
    "to_multiplicity": "1",
    "details": "Each line item refers to one product"
}
```

### Many-to-Many
```json
{
    "name": "Product Has Categories",
    "from_class_key": "product",
    "from_multiplicity": "*",
    "to_class_key": "category",
    "to_multiplicity": "*",
    "details": "Products can belong to multiple categories; categories contain multiple products"
}
```

## Association Classes

For relationships that have their own attributes, use an association class:

```json
{
    "name": "Student Enrollment",
    "from_class_key": "student",
    "from_multiplicity": "*",
    "to_class_key": "course",
    "to_multiplicity": "*",
    "association_class_key": "enrollment",
    "details": "Enrollment tracks the grade and semester for each student-course pairing"
}
```

## Naming Conventions

Use descriptive names that describe the relationship:
- `order_has_line_items` - Clear parent-child relationship
- `user_manages_account` - Action-oriented relationship
- `product_in_category` - Membership relationship

## Related Errors

- **E6001**: Association name is required
- **E11002**: Association from_class_key not found
- **E11003**: Association to_class_key not found
- **E11016**: Invalid multiplicity format
