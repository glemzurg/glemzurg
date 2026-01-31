# Association Filename Invalid Component (E11028)

A component within an association filename does not follow the required snake_case format.

## What Went Wrong

Association filenames are compound names with multiple components separated by `--` and `.`. Each individual component must be valid snake_case:

- **from_class**, **to_class**, **name** - The three main parts
- **from_subdomain**, **to_subdomain** - For domain-level associations
- **from_domain**, **to_domain** - For model-level associations

The error indicates that one of these components contains invalid characters or formatting.

## Valid Component Format

Each component must:
- Start with a **lowercase letter** (a-z)
- Contain only **lowercase letters** (a-z), **numbers** (0-9), and **underscores** (_)
- Use underscores to separate words
- Not start or end with an underscore
- Not have consecutive underscores

## Common Mistakes

### Uppercase Letters

```
# WRONG - Uppercase in class name
BookOrder--book_order_line--order_lines.assoc.json

# CORRECT - All lowercase
book_order--book_order_line--order_lines.assoc.json
```

### Hyphens Instead of Underscores

```
# WRONG - Hyphens in component
book-order--book-order-line--order-lines.assoc.json

# CORRECT - Underscores in components
book_order--book_order_line--order_lines.assoc.json
```

Note: `--` (double hyphen) separates parts, but single hyphens within a part are invalid.

### Spaces

```
# WRONG - Spaces in component
book order--line item--order lines.assoc.json

# CORRECT - Use underscores
book_order--line_item--order_lines.assoc.json
```

### Starting with Number

```
# WRONG - Component starts with number
2nd_order--line--orders.assoc.json

# CORRECT - Start with letter
second_order--line--orders.assoc.json
```

### Invalid Path Component (Domain/Subdomain Level)

```
# WRONG - Uppercase subdomain at domain level
Orders.book_order--Shipping.shipment--order_shipment.assoc.json

# CORRECT - Lowercase subdomain
orders.book_order--shipping.shipment--order_shipment.assoc.json
```

## How to Fix

### Step 1: Identify the Problem Component

The error message tells you:
- Which component is invalid (e.g., `from_class`, `to_subdomain`, `name`)
- The invalid value
- What's wrong with it

### Step 2: Fix the Component

Convert the component to valid snake_case:

```bash
# Fix uppercase
mv BookOrder--line--orders.assoc.json book_order--line--orders.assoc.json

# Fix hyphens
mv book-order--line--orders.assoc.json book_order--line--orders.assoc.json

# Fix spaces (usually requires quoting)
mv "book order--line--orders.assoc.json" book_order--line--orders.assoc.json
```

### Step 3: Verify All Components

Check each part of the filename:

| Filename Part | Example | Must Be |
|---------------|---------|---------|
| From class | `book_order` | snake_case |
| To class | `book_order_line` | snake_case |
| Association name | `order_lines` | snake_case |
| Subdomain (if present) | `default` | snake_case |
| Domain (if present) | `order_fulfillment` | snake_case |

## Examples by Level

### Subdomain Level

Format: `{from_class}--{to_class}--{name}`

| Invalid | Problem | Correct |
|---------|---------|---------|
| `BookOrder--line--orders` | Uppercase from_class | `book_order--line--orders` |
| `order--LineItem--orders` | Uppercase to_class | `order--line_item--orders` |
| `order--line--OrderLines` | Uppercase name | `order--line--order_lines` |

### Domain Level

Format: `{from_sub}.{from_class}--{to_sub}.{to_class}--{name}`

| Invalid | Problem | Correct |
|---------|---------|---------|
| `Orders.order--shipping.ship--link` | Uppercase subdomain | `orders.order--shipping.ship--link` |
| `orders.Order--shipping.ship--link` | Uppercase class | `orders.order--shipping.ship--link` |

### Model Level

Format: `{from_dom}.{from_sub}.{from_class}--{to_dom}.{to_sub}.{to_class}--{name}`

| Invalid | Problem | Correct |
|---------|---------|---------|
| `Sales.orders.order--...` | Uppercase domain | `sales.orders.order--...` |

## Why Snake_Case Components?

### Consistency

Every key in the model follows the same rules:
- Class names: `book_order`
- Domain names: `order_fulfillment`
- Association names: `order_lines`

### Code Generation

Components map to identifiers in generated code. Invalid characters cause problems:
- `book-order` becomes invalid in most languages
- `BookOrder` may conflict with types
- Spaces are universally problematic

### Cross-Referencing

Components in filenames must match keys used elsewhere:
- `from_class_key` in the JSON file must match the filename component
- Invalid characters break these references

## Related Errors

- **E11026**: Key invalid format (for simple keys)
- **E11027**: Association filename invalid format (wrong number of parts)
