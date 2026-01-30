# JSON Model Format Proposal

## Overview

This document proposes a new JSON-based format for defining requirements models. The format is designed to be:

1. **AI-friendly** - Easy for AI to generate and modify individual files
2. **Human-reviewable** - Each file is small and focused, making corrections easy
3. **Intuitive structure** - Directory layout mirrors the conceptual model hierarchy

## What This Model Represents

This model captures **requirements** — the distilled behavior of a system as understood by human readers. It describes *what* the system does, not *how* it is implemented.

### Abstraction Over Implementation

The model represents a **logical abstraction** that is independent of deployment architecture. It does not distinguish between client and server, frontend and backend, or any particular technical boundary. Instead, it describes behavior as a unified whole:

- **Objects and State**: A class's attributes and state represent the complete logical state of that concept, regardless of where the data physically resides. An "Order" object might have some data stored in a client-side cache, some in a server database, and some computed on-the-fly — but in the requirements model, it is simply an "Order" with its attributes.

- **State Machines**: When an object transitions between states, the model captures the business meaning of that transition. Whether the transition is triggered by a user action on a mobile app, a webhook from a payment processor, or a scheduled job on a server is an implementation detail not expressed here.

- **Actions**: An action like "Send Confirmation Email" describes the business effect. The model does not specify whether this happens synchronously, asynchronously, on which server, or through which email provider.

- **Queries**: A query like "Get Available Products" describes what information is returned and under what conditions. The query might involve filtering options that only exist in a client application's UI, database queries on a server, and aggregation across multiple microservices — but the requirements model presents it as a single coherent operation.

### Why This Matters

This abstraction serves several purposes:

1. **Human Understanding**: Stakeholders can review and validate system behavior without needing to understand technical architecture. A product manager can verify that "an order can only be cancelled before it ships" without knowing which services enforce this rule.

2. **Implementation Flexibility**: The same requirements model can be implemented in various architectures. A monolithic application, a microservices deployment, or a serverless architecture could all satisfy the same requirements.

3. **Focus on Behavior**: By removing implementation concerns, the model keeps focus on what matters most during requirements gathering: what the system should do for its users.

4. **Communication Bridge**: The model serves as a shared language between technical and non-technical team members, capturing business logic in a form both can understand and validate.

### Example: Cross-Boundary Behavior

Consider a "Shopping Cart" class with a "total" attribute and a "Calculate Total" action:

- The cart items might be stored in browser local storage
- The product prices might come from a server API
- Tax calculations might involve a third-party tax service
- Discount rules might be evaluated on the server
- The final total might be displayed in the mobile app UI

In the requirements model, none of this complexity appears. There is simply a Shopping Cart with items, and when you ask for the total, you get the correct value reflecting prices, taxes, and discounts. The model describes the *behavior* humans expect, not the *mechanism* that delivers it.

## Design Principles

- **One concept per file** - Each actor, class, etc. is its own JSON file
- **Directory structure reflects hierarchy** - Folders organize content by domain and type
- **Scoped keys in references** - Cross-references use minimal keys scoped to the current context for human readability
- **Minimal required fields** - Only essential fields are required; everything else is optional
- **Separate complex concerns** - State machines, actions, and queries are separate files since they will become complex

## Proposed Directory Structure

```
models_json/
└── {model_name}/
    ├── model.json                    # Model metadata
    ├── actors/
    │   ├── customer.json
    │   ├── manager.json
    │   └── publisher.json
    ├── associations/                 # Model-level associations (cross-domain)
    │   └── order_fulfillment.default.book_order_line--inventory.default.inventory_item--order_inventory.json
    └── domains/
        └── {domain_name}/
            ├── domain.json           # Domain metadata
            ├── associations/         # Domain-level associations (cross-subdomain)
            │   └── orders.book_order--shipping.shipment--order_shipment.json
            └── subdomains/
                └── {subdomain_name}/
                    ├── subdomain.json    # Subdomain metadata (optional)
                    ├── associations/     # Subdomain-level associations (within subdomain)
                    │   ├── book_order--book_order_line--order_lines.json
                    │   └── book_order--customer--customer_orders.json
                    ├── classes/
                    │   ├── book_order/
                    │   │   ├── class.json        # Class definition (attributes only)
                    │   │   ├── state_machine.json # States, events, guards, transitions
                    │   │   ├── actions/
                    │   │   │   ├── calculate_total.json
                    │   │   │   ├── notify_warehouse.json
                    │   │   │   └── refund_payment.json
                    │   │   └── queries/
                    │   │       ├── get_subtotal.json
                    │   │       └── is_cancellable.json
                    │   ├── book_order_line/
                    │   │   ├── class.json
                    │   │   └── state_machine.json
                    │   └── customer/
                    │       └── class.json
                    └── generalizations/
                        └── medium.json
```

## File Formats

### model.json

Based on `req_model.Model`:

```json
{
  "name": "Web Books",
  "details": "An online bookstore application."
}
```

**Fields:**
- `name` (required): Display name of the model
- `details` (optional): Markdown description

### actors/{actor_name}.json

Based on `model_actor.Actor`:

```json
{
  "name": "Customer",
  "type": "person",
  "details": "A person who purchases books from the online store.",
  "uml_comment": "Primary user"
}
```

**Fields:**
- `name` (required): Display name of the actor
- `type` (required): Either `"person"` or `"system"`
- `details` (optional): Markdown description
- `uml_comment` (optional): Comment for UML diagrams

### domains/{domain}/domain.json

Based on `model_domain.Domain`:

```json
{
  "name": "Order Fulfillment",
  "details": "Handles the lifecycle of customer orders.",
  "realized": false,
  "uml_comment": "Core business domain"
}
```

**Fields:**
- `name` (required): Display name of the domain
- `details` (optional): Markdown description
- `realized` (optional, default false): If true, this domain has no semantic model because it already exists
- `uml_comment` (optional): Comment for UML diagrams

### classes/{class_name}/class.json

Based on `model_class.Class` and `model_class.Attribute`:

```json
{
  "name": "Book Order",
  "details": "Represents a customer's order for books.",
  "actor_key": "customer",
  "uml_comment": "Aggregate root",

  "attributes": {
    "id": {
      "name": "ID",
      "data_type_rules": "int",
      "details": "Unique identifier",
      "derivation_policy": "",
      "nullable": false,
      "uml_comment": ""
    },
    "status": {
      "name": "Status",
      "data_type_rules": "enum(pending, confirmed, shipped, delivered)",
      "details": "Current order status",
      "derivation_policy": "",
      "nullable": false,
      "uml_comment": ""
    },
    "total": {
      "name": "Total",
      "data_type_rules": "decimal",
      "details": "Order total amount",
      "derivation_policy": "",
      "nullable": false,
      "uml_comment": ""
    }
  },

  "indexes": [
    ["id"],
    ["status", "total"]
  ]
}
```

**Class Fields (from `model_class.Class`):**
- `name` (required): Display name of the class
- `details` (optional): Markdown description
- `actor_key` (optional): Actor name (scoped to model actors)
- `uml_comment` (optional): Comment for UML diagrams

**Attribute Fields (from `model_class.Attribute`):**

Attributes are stored as an object keyed by attribute key.

- `name` (required): Display name of the attribute
- `data_type_rules` (optional): Data type specification string
- `details` (optional): Markdown description
- `derivation_policy` (optional): How this derived attribute is computed
- `nullable` (optional, default false): Whether this attribute can be null
- `uml_comment` (optional): Comment for UML diagrams

**Indexes:**

- `indexes` (optional): Array of arrays of attribute keys. Each inner array defines one index.

### associations/{association_name}.json

Based on `model_class.Association`. Associations exist at three levels depending on which classes they connect:

- **Subdomain-level** (`subdomains/{subdomain}/associations/`): Classes within the same subdomain
- **Domain-level** (`domains/{domain}/associations/`): Classes from different subdomains within the same domain
- **Model-level** (`associations/`): Classes from different domains

The filename includes enough context to be unique at that level. Separators:
- `.` separates domain from subdomain from class
- `--` separates the from-class, to-class, and distilled name

The distilled name is the full association name, lowercase with `_` between words.

| Level | Filename Pattern | Example |
|-------|-----------------|---------|
| Subdomain | `{from_class}--{to_class}--{name}.json` | `book_order--book_order_line--order_lines.json` |
| Domain | `{from_subdomain}.{from_class}--{to_subdomain}.{to_class}--{name}.json` | `orders.book_order--shipping.shipment--order_shipment.json` |
| Model | `{from_domain}.{from_subdomain}.{from_class}--{to_domain}.{to_subdomain}.{to_class}--{name}.json` | `order_fulfillment.default.book_order_line--inventory.default.inventory_item--order_inventory.json` |

**Subdomain-level association** (`subdomains/default/associations/book_order--customer--customer_orders.json`):

Keys are scoped to the subdomain (just class names).

```json
{
  "name": "Customer Orders",
  "details": "Links an order to the customer who placed it",
  "from_class_key": "book_order",
  "from_multiplicity": "*",
  "to_class_key": "customer",
  "to_multiplicity": "1",
  "association_class_key": null,
  "uml_comment": ""
}
```

**Domain-level association** (`domains/order_fulfillment/associations/orders.book_order--shipping.shipment--order_shipment.json`):

Keys include subdomain to disambiguate (subdomain/class).

```json
{
  "name": "Order Shipment",
  "details": "Links an order to its shipment tracking",
  "from_class_key": "orders/book_order",
  "from_multiplicity": "1",
  "to_class_key": "shipping/shipment",
  "to_multiplicity": "0..1",
  "association_class_key": null,
  "uml_comment": ""
}
```

**Model-level association** (`associations/order_fulfillment.default.book_order_line--inventory.default.inventory_item--order_inventory.json`):

Keys include domain and subdomain (domain/subdomain/class).

```json
{
  "name": "Order Inventory",
  "details": "Links order lines to inventory items across domains",
  "from_class_key": "order_fulfillment/default/book_order_line",
  "from_multiplicity": "*",
  "to_class_key": "inventory/default/inventory_item",
  "to_multiplicity": "1",
  "association_class_key": null,
  "uml_comment": ""
}
```

**Association Fields (from `model_class.Association`):**
- `name` (required): Display name of the association
- `from_class_key` (required): Scoped key of the source class
- `from_multiplicity` (required): Multiplicity on the "from" side
- `to_class_key` (required): Scoped key of the target class
- `to_multiplicity` (required): Multiplicity on the "to" side
- `association_class_key` (optional): Scoped key of association class if any
- `details` (optional): Markdown description
- `uml_comment` (optional): Comment for UML diagrams

**Multiplicity Format:**

Multiplicity defines cardinality constraints on associations. Valid formats:

| Format | Meaning | Example |
|--------|---------|---------|
| `"1"` | Exactly one | Required 1-to-1 |
| `"0..1"` | Zero or one | Optional relationship |
| `"*"` | Zero or more (unbounded) | Optional many |
| `"1..*"` | One or more | Required many |
| `"n"` | Exactly n | `"3"` means exactly 3 |
| `"n..m"` | Range from n to m | `"2..5"` means 2 to 5 |
| `"n..*"` | n or more | `"3..*"` means 3 or more |
| `"0..n"` | Zero to n | `"0..3"` means 0 to 3 |

Rules:
- Single number `"n"` means exactly n (both lower and upper bound)
- `"*"` represents unbounded
- Upper bound must be >= lower bound (unless upper is unbounded)

### classes/{class_name}/state_machine.json

Based on `model_state.State`, `model_state.Event`, `model_state.Guard`, and `model_state.Transition`:

```json
{
  "states": {
    "pending": {
      "name": "Pending",
      "details": "Order created but not yet confirmed",
      "uml_comment": "",
      "actions": [
        {
          "action_key": "log_pending",
          "when": "entry"
        }
      ]
    },
    "confirmed": {
      "name": "Confirmed",
      "details": "Order confirmed and ready for processing",
      "uml_comment": "",
      "actions": [
        {
          "action_key": "reserve_inventory",
          "when": "entry"
        }
      ]
    },
    "shipped": {
      "name": "Shipped",
      "details": "Order has been shipped to customer",
      "uml_comment": "",
      "actions": []
    },
    "delivered": {
      "name": "Delivered",
      "details": "Order delivered to customer",
      "uml_comment": "",
      "actions": []
    },
    "cancelled": {
      "name": "Cancelled",
      "details": "Order was cancelled",
      "uml_comment": "",
      "actions": [
        {
          "action_key": "release_inventory",
          "when": "entry"
        }
      ]
    }
  },

  "events": {
    "place": {
      "name": "place",
      "details": "Customer places the order",
      "parameters": [
        {"name": "items", "source": "cart.items"},
        {"name": "shipping_address", "source": "customer.address"}
      ]
    },
    "confirm": {
      "name": "confirm",
      "details": "Order payment confirmed",
      "parameters": [
        {"name": "payment_id", "source": "payment.id"}
      ]
    },
    "ship": {
      "name": "ship",
      "details": "Order shipped",
      "parameters": [
        {"name": "tracking_number", "source": "shipment.tracking"}
      ]
    },
    "deliver": {
      "name": "deliver",
      "details": "Order delivered",
      "parameters": []
    },
    "cancel": {
      "name": "cancel",
      "details": "Order cancelled",
      "parameters": [
        {"name": "reason", "source": "user.input"}
      ]
    }
  },

  "guards": {
    "has_items": {
      "name": "hasItems",
      "details": "Order has at least one line item"
    },
    "payment_valid": {
      "name": "paymentValid",
      "details": "Payment has been validated"
    },
    "in_stock": {
      "name": "inStock",
      "details": "All items are in stock"
    },
    "not_shipped": {
      "name": "notShipped",
      "details": "Order has not been shipped yet"
    }
  },

  "transitions": [
    {
      "from_state_key": null,
      "to_state_key": "pending",
      "event_key": "place",
      "guard_key": "has_items",
      "action_key": "calculate_total",
      "uml_comment": "Initial transition"
    },
    {
      "from_state_key": "pending",
      "to_state_key": "confirmed",
      "event_key": "confirm",
      "guard_key": "payment_valid",
      "action_key": "notify_warehouse",
      "uml_comment": ""
    },
    {
      "from_state_key": "confirmed",
      "to_state_key": "shipped",
      "event_key": "ship",
      "guard_key": "in_stock",
      "action_key": "send_tracking_email",
      "uml_comment": ""
    },
    {
      "from_state_key": "shipped",
      "to_state_key": "delivered",
      "event_key": "deliver",
      "guard_key": null,
      "action_key": "send_delivery_confirmation",
      "uml_comment": ""
    },
    {
      "from_state_key": "pending",
      "to_state_key": "cancelled",
      "event_key": "cancel",
      "guard_key": null,
      "action_key": "refund_payment",
      "uml_comment": ""
    },
    {
      "from_state_key": "confirmed",
      "to_state_key": "cancelled",
      "event_key": "cancel",
      "guard_key": "not_shipped",
      "action_key": "refund_payment",
      "uml_comment": ""
    }
  ]
}
```

**State Fields (from `model_state.State`):**

States are stored as an object keyed by state key.

- `name` (required): Display name of the state
- `details` (optional): Markdown description
- `uml_comment` (optional): Comment for UML diagrams
- `actions` (optional): Array of state actions (entry/exit/do)

**State Action Fields (from `model_state.StateAction`):**
- `action_key` (required): Key of the action to execute
- `when` (required): When to execute - `"entry"`, `"exit"`, or `"do"`

**Event Fields (from `model_state.Event`):**

Events are stored as an object keyed by event key.

- `name` (required): Display name of the event
- `details` (optional): Description
- `parameters` (optional): Array of event parameters

**Event Parameter Fields (from `model_state.EventParameter`):**
- `name` (required): Parameter name
- `source` (required): Where the parameter value comes from

**Guard Fields (from `model_state.Guard`):**

Guards are stored as an object keyed by guard key.

- `name` (required): Simple name for internal use
- `details` (required): Description of the guard condition (shown in UML)

**Transition Fields (from `model_state.Transition`):**
- `from_state_key` (optional): State key to transition from (null for initial transitions)
- `to_state_key` (optional): State key to transition to (null for final transitions)
- `event_key` (required): Event that triggers this transition
- `guard_key` (optional): Guard condition for this transition
- `action_key` (optional): Action to execute during transition
- `uml_comment` (optional): Comment for UML diagrams

Note: At least one of `from_state_key` or `to_state_key` must be specified.

### classes/{class_name}/actions/{action_name}.json

Based on `model_state.Action`. Each action is its own file. The filename (without .json) becomes the action key.

```json
{
  "name": "calculateTotal",
  "details": "Sum up line item prices and apply taxes/discounts",
  "requires": [
    "Order has at least one line item",
    "All line items have valid prices"
  ],
  "guarantees": [
    "Order.total is set to the computed value",
    "Total is non-negative"
  ]
}
```

**Action Fields (from `model_state.Action`):**
- `name` (required): Display name of the action
- `details` (optional): Description of what the action does
- `requires` (optional): Array of preconditions
- `guarantees` (optional): Array of postconditions

Another example (`actions/notify_warehouse.json`):

```json
{
  "name": "notifyWarehouse",
  "details": "Send order details to warehouse for fulfillment",
  "requires": [
    "Order is in confirmed state",
    "All items have been validated"
  ],
  "guarantees": [
    "Warehouse notification message is queued",
    "Order.warehouse_notified is set to true"
  ]
}
```

### classes/{class_name}/queries/{query_name}.json

Based on `model_state.Query`. Each query is its own file. The filename (without .json) becomes the query key.

```json
{
  "name": "getSubtotal",
  "details": "Calculate subtotal before tax and discounts",
  "requires": [
    "Order exists"
  ],
  "guarantees": [
    "Returns sum of (line.price * line.quantity) for all lines",
    "Return value is non-negative"
  ]
}
```

**Query Fields (from `model_state.Query`):**
- `name` (required): Display name of the query
- `details` (optional): Description of what the query returns
- `requires` (optional): Array of preconditions
- `guarantees` (optional): Array of postconditions/return value guarantees

Another example (`queries/is_cancellable.json`):

```json
{
  "name": "isCancellable",
  "details": "Check if order can still be cancelled",
  "requires": [],
  "guarantees": [
    "Returns true if state is pending or confirmed",
    "Returns false otherwise"
  ]
}
```

### generalizations/{generalization_name}.json

Based on `model_class.Generalization`:

```json
{
  "name": "Medium",
  "details": "Different formats a book can be published in",
  "superclass_key": "product",
  "subclass_keys": ["book", "ebook", "audiobook"],
  "is_complete": true,
  "is_static": true,
  "uml_comment": ""
}
```

**Generalization Fields (from `model_class.Generalization`):**
- `name` (required): Display name of the generalization
- `superclass_key` (required): Class key for the superclass (scoped to same subdomain)
- `subclass_keys` (required): Array of class keys for the subclasses (at least one required, scoped to same subdomain)
- `details` (optional): Markdown description
- `is_complete` (optional, default false): Are the specializations complete, or can an instantiation exist without a specialization
- `is_static` (optional, default false): Are the specializations static and unchanging, or can they change at runtime
- `uml_comment` (optional): Comment for UML diagrams

## Key Design Decisions

### 1. Scoped Keys

**Decision**: Use minimal scoped keys that include only the context needed for unambiguous reference.

**Rationale**:
- Human-readable: `book_order` is clearer than `domain/order_fulfillment/subdomain/default/class/book_order`
- Context is implicit from file location
- Keys grow only when needed to disambiguate (e.g., cross-subdomain needs `subdomain/class`)

**Scoping rules:**
- Within a subdomain: just the entity name (e.g., `book_order`, `customer`)
- Within a domain (cross-subdomain): `subdomain/entity` (e.g., `orders/book_order`)
- Within a model (cross-domain): `domain/subdomain/entity` (e.g., `order_fulfillment/default/book_order_line`)
- Actors are always model-scoped: just the actor name (e.g., `customer`)

### 2. Associations at Appropriate Levels

**Decision**: Associations are stored in separate files at model, domain, or subdomain level based on which classes they connect.

**Rationale**:
- Models can have many associations; keeping them in one list is unwieldy.
- The level (model/domain/subdomain) indicates the scope of the relationship
- Subdomain associations connect classes within the same subdomain
- Domain associations connect classes across subdomains within the same domain
- Model associations connect classes across different domains
- Each association can be reviewed and modified independently
- Easy to see all relationships at a given scope by listing the associations directory

### 3. State Machine as Separate File

**Decision**: State machine (states, events, guards, transitions) is in its own file.

**Rationale**:
- State machines can become complex with many states and transitions
- Easier to review and modify state behavior independently
- AI can focus on state machine logic without class structure noise
- Classes without state machines simply don't have this file

### 4. Each Action is Its Own File

**Decision**: Each action is a separate file in an `actions/` directory.

**Rationale**:
- Actions will eventually contain detailed implementation logic
- Each action can be refined independently without touching others
- Easy to say "fix the calculate_total action" and know exactly which file
- Clear separation between "what happens" (state machine) and "how it happens" (actions)
- Filename becomes the action key

### 5. Each Query is Its Own File

**Decision**: Each query is a separate file in a `queries/` directory.

**Rationale**:
- Queries represent read-only computations that can become complex
- Each query can be refined independently
- Easy to say "fix the get_subtotal query" and know exactly which file
- Separating them makes it clear what data is stored vs computed
- Filename becomes the query key

### 6. Directory Names as Keys

**Decision**: File and directory names become the sub-keys for identity.

**Rationale**:
- Intuitive mapping from filesystem to model
- Easy to see what exists without opening files
- AI can create new entities by creating new files

## Constraints

1. **Default subdomain only**: For the initial implementation, only the `default` subdomain is supported. All classes must be in `subdomains/default/`.

2. **Actions are independent**: Actions do not reference or call other actions.

## Implementation Architecture

### Separate Go Structs for JSON Import

The JSON import package will use its own set of Go structs, separate from the `req_model` classes. This separation exists for several reasons:

1. **Optimized input shapes**: The JSON format can use structures that are easier for input (e.g., superclass/subclass defined in generalization rather than spread across class files).

2. **Distinct error handling**: Each validation error has two distinct components:
   - **Unique error number**: Every error type has its own identifier for programmatic handling and documentation reference.
   - **Detailed output with construction advice**: Error messages include specific guidance on how to correct the input, helping AI or human authors fix issues quickly.

   These error types are specific to import validation and don't belong in the main model tree.

3. **Clear separation of concerns**: Import structs handle parsing and validation of external input; `req_model` structs represent the canonical internal model.

4. **Conversion layer**: After successful validation, import structs are converted to `req_model` structs for use in the rest of the system.

### JSON Schema Validation

Each JSON file type will have a corresponding JSON Schema for validation:

- `model.schema.json`
- `actor.schema.json`
- `domain.schema.json`
- `subdomain.schema.json`
- `class.schema.json`
- `association.schema.json`
- `state_machine.schema.json`
- `action.schema.json`
- `query.schema.json`
- `generalization.schema.json`

These schemas provide:
- Early validation before Go parsing
- Clear documentation of expected structure
- IDE support for editing JSON files
- Consistent error messages for structural issues
