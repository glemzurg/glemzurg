# Association Filename Invalid Format (E11027)

An association filename does not follow the required compound format with `--` separators.

## What Went Wrong

Association filenames must follow a specific compound format that encodes the relationship between classes. The filename must have exactly three parts separated by `--` (double hyphen):

1. **from** - The source class (or path to it)
2. **to** - The target class (or path to it)
3. **name** - The association name

## Required Format by Level

### Subdomain-Level Associations

Located at: `domains/{domain}/subdomains/{subdomain}/associations/`

Format: `{from_class}--{to_class}--{name}.assoc.json`

Examples:
- `book_order--book_order_line--order_lines.assoc.json`
- `customer--order--customer_orders.assoc.json`
- `product--category--product_categories.assoc.json`

### Domain-Level Associations

Located at: `domains/{domain}/associations/`

Format: `{from_subdomain}.{from_class}--{to_subdomain}.{to_class}--{name}.assoc.json`

Examples:
- `orders.book_order--shipping.shipment--order_shipment.assoc.json`
- `customers.customer--orders.order--customer_orders.assoc.json`

### Model-Level Associations

Located at: `associations/`

Format: `{from_domain}.{from_subdomain}.{from_class}--{to_domain}.{to_subdomain}.{to_class}--{name}.assoc.json`

Examples:
- `order_fulfillment.default.book_order_line--inventory.default.inventory_item--order_inventory.assoc.json`
- `sales.orders.order--finance.billing.invoice--order_invoice.assoc.json`

## Common Mistakes

### Missing Parts

```
# WRONG - Only 2 parts
book_order--book_order_line.assoc.json

# CORRECT - All 3 parts
book_order--book_order_line--order_lines.assoc.json
```

### Too Many Parts

```
# WRONG - 4 parts
book_order--book_order_line--order_lines--extra.assoc.json

# CORRECT - Exactly 3 parts
book_order--book_order_line--order_lines.assoc.json
```

### Using Single Hyphen

```
# WRONG - Single hyphen doesn't separate parts
book_order-book_order_line-order_lines.assoc.json

# CORRECT - Double hyphen (--) separates parts
book_order--book_order_line--order_lines.assoc.json
```

### Wrong Depth for Level

```
# WRONG - Subdomain level with domain prefix
default.book_order--default.line--order_lines.assoc.json

# CORRECT - Subdomain level uses simple class names
book_order--book_order_line--order_lines.assoc.json
```

## How to Fix

### Step 1: Identify the Problem

The error message tells you:
- The filename that's invalid
- How many parts were found (expected 3)
- The file path

### Step 2: Rename the File

Ensure your filename has exactly three parts:

```bash
# Example: Add missing name part
mv book_order--book_order_line.assoc.json book_order--book_order_line--order_lines.assoc.json

# Example: Fix single hyphens
mv book-order--line-item--lines.assoc.json book_order--line_item--lines.assoc.json
```

### Step 3: Check the Level

Make sure you're using the right format for where the file is located:

| Location | Format |
|----------|--------|
| `subdomains/{sub}/associations/` | `class--class--name` |
| `domains/{dom}/associations/` | `sub.class--sub.class--name` |
| `associations/` (model root) | `dom.sub.class--dom.sub.class--name` |

## Why This Format?

### 1. Self-Documenting

The filename tells you exactly what the association connects:
- `book_order--book_order_line--order_lines.assoc.json`
- Connects `book_order` to `book_order_line`, named `order_lines`

### 2. Unique Identification

The compound format ensures uniqueness:
- Two classes can have multiple associations with different names
- `book_order--product--ordered_products.assoc.json`
- `book_order--product--featured_products.assoc.json`

### 3. Path Resolution

For cross-subdomain and cross-domain associations, the path components (dot-separated) tell you exactly where to find each class.

## Related Errors

- **E11026**: Key invalid format (for simple keys like class names)
- **E11028**: Association filename invalid component (when a part isn't valid snake_case)
