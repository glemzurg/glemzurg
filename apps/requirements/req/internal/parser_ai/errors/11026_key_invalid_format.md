# Key Invalid Format (E11026)

A file or directory name used as a key does not follow the required snake_case format.

## What Went Wrong

Keys in the model are derived from file and directory names. They must follow strict snake_case formatting rules to ensure consistency, cross-referencing, and code generation compatibility.

**Note:** This error applies to simple keys (actors, domains, subdomains, classes, actions, queries, generalizations). Association filenames follow a different compound format - see "Association Filenames" below.

## Key Naming Rules

### Valid Format

Keys must:
- Start with a **lowercase letter** (a-z)
- Contain only **lowercase letters** (a-z), **numbers** (0-9), and **underscores** (_)
- Use underscores to separate words
- Not start or end with an underscore
- Not have consecutive underscores

### Valid Examples

| Key | Derived From |
|-----|--------------|
| `order` | `order/` directory or `order.actor.json` file |
| `book_order` | `book_order/` directory |
| `order_line_item` | `order_line_item/` directory |
| `customer2` | `customer2.actor.json` file |
| `v2_order` | `v2_order/` directory |
| `get_subtotal` | `get_subtotal.json` action file |

### Invalid Examples

| Invalid Key | Problem | Correct Key |
|-------------|---------|-------------|
| `BookOrder` | Contains uppercase | `book_order` |
| `book-order` | Contains hyphen | `book_order` |
| `2order` | Starts with number | `order2` or `v2_order` |
| `order line` | Contains space | `order_line` |
| `order.line` | Contains period | `order_line` |
| `_order` | Starts with underscore | `order` |
| `order_` | Ends with underscore | `order` |
| `order__line` | Consecutive underscores | `order_line` |
| `Order` | Contains uppercase | `order` |

## Where Keys Come From

Keys are derived from filenames and directory names throughout the model structure:

```
your_model/
├── model.json
├── actors/
│   └── customer.actor.json          <- actor_key: "customer"
├── domains/
│   └── order_fulfillment/           <- domain_key: "order_fulfillment"
│       ├── domain.json
│       └── subdomains/
│           └── default/             <- subdomain_key: "default"
│               ├── subdomain.json
│               ├── associations/
│               │   └── order_lines.assoc.json  <- association_key: "order_lines"
│               ├── generalizations/
│               │   └── medium.gen.json         <- generalization_key: "medium"
│               └── classes/
│                   └── book_order/             <- class_key: "book_order"
│                       ├── class.json
│                       ├── state_machine.json
│                       ├── actions/
│                       │   └── calculate_total.json  <- action_key: "calculate_total"
│                       └── queries/
│                           └── get_subtotal.json     <- query_key: "get_subtotal"
```

## How to Fix

### Step 1: Identify the Invalid Key

The error message will tell you which key is invalid and in which file.

### Step 2: Rename the File or Directory

Rename the file or directory to use valid snake_case:

```bash
# Example: Rename a directory
mv BookOrder/ book_order/

# Example: Rename an actor file
mv Customer.actor.json customer.actor.json

# Example: Rename an action file
mv calculateTotal.json calculate_total.json
```

### Step 3: Update Any References

If other files reference the old key, update them to use the new key:

```json
// In class.json, update actor_key
{
    "name": "Book Order",
    "actor_key": "customer"  // Must match the actor filename
}

// In state_machine.json, update action_key references
{
    "transitions": [
        {
            "action_key": "calculate_total"  // Must match the action filename
        }
    ]
}
```

## Why Snake_Case?

### 1. Filesystem Compatibility

Snake_case works on all operating systems without escaping or special handling. Names with spaces, special characters, or case sensitivity issues can cause problems on different platforms.

### 2. Code Generation

Keys map directly to identifiers in generated code. Snake_case converts cleanly to various naming conventions:

| Snake_case Key | Language Convention | Generated Name |
|----------------|---------------------|----------------|
| `book_order` | Go (exported) | `BookOrder` |
| `book_order` | Go (unexported) | `bookOrder` |
| `book_order` | Python | `book_order` |
| `book_order` | JavaScript | `bookOrder` |
| `book_order` | SQL | `book_order` |

### 3. Readability

Snake_case provides clear word boundaries:
- `orderlineitem` vs `order_line_item` — underscores make words obvious
- No ambiguity like camelCase (`iPhone` — is it `IPhone` or `Iphone`?)

### 4. Consistency

One canonical form prevents variations. Without strict rules, you might end up with:
- `BookOrder`, `bookOrder`, `book_order`, `BOOK_ORDER` — all referring to the same thing

With snake_case enforced, there's only one valid form: `book_order`.

### 5. Cross-Referencing

Keys appear in JSON files and must be typed exactly. Snake_case is:
- Easy to type (no shift key for underscores on most keyboards)
- Easy to read in JSON
- Unambiguous for matching

## Common Mistakes

### Using Display Names Instead of Keys

```json
// WRONG: Using the display name as the key
{
    "actor_key": "Customer"
}

// CORRECT: Using the file-derived key
{
    "actor_key": "customer"  // From customer.actor.json
}
```

### Copying from Other Systems

If you're converting from another system, names may not follow snake_case:

```bash
# Original: BookOrder.java
# Create as: book_order/class.json

# Original: calculateTotal()
# Create as: actions/calculate_total.json
```

### Acronyms and Abbreviations

```
# WRONG
API_endpoint/
HTMLParser/

# CORRECT
api_endpoint/
html_parser/
```

## Association Filenames

Association filenames follow a **different format** than simple keys. They use a compound structure with `--` separators to encode relationship information:

### Subdomain-Level Associations

Format: `{from_class}--{to_class}--{name}.assoc.json`

Example: `book_order--book_order_line--order_lines.assoc.json`

Each component (`book_order`, `book_order_line`, `order_lines`) must be valid snake_case.

### Domain-Level Associations

Format: `{from_subdomain}.{from_class}--{to_subdomain}.{to_class}--{name}.assoc.json`

Example: `orders.book_order--shipping.shipment--order_shipment.assoc.json`

### Model-Level Associations

Format: `{from_domain}.{from_subdomain}.{from_class}--{to_domain}.{to_subdomain}.{to_class}--{name}.assoc.json`

Example: `order_fulfillment.default.book_order_line--inventory.default.inventory_item--order_inventory.assoc.json`

### Association Filename Rules

- Use `--` (double hyphen) to separate the three main parts: from, to, and name
- Use `.` (dot) to separate path components within from/to (domain.subdomain.class)
- Each individual component (domain, subdomain, class, name) must be valid snake_case
- The association name (last component) describes the relationship

## Related Errors

- **E2008**: Actor filename invalid
- **E3006**: Domain directory invalid
- **E4006**: Subdomain directory invalid
- **E5006**: Class directory invalid
- **E6013**: Association filename invalid
- **E8006**: Action filename invalid
- **E9006**: Query filename invalid
- **E10011**: Class generalization filename invalid
