# JSON Model Format

## Overview

This document describes the JSON-based format for defining requirements models. The format is designed to be:

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

## Deriving a Model from an Existing System

When examining another system to create a model, use these guidelines to identify model elements:

### Classes from Persistent Storage

Classes can be derived from the persistent storage in the existing system:
- **Database tables** → Classes with attributes matching columns
- **Cache structures** → Classes representing cached entities
- **Message queue payloads** → Classes for message data types
- **Configuration stores** → Classes for configurable entities

Each persistent entity typically becomes a class. The columns, fields, or properties become attributes.

### State Machines for Stateless Classes

If a class has no obvious state attribute (no `status`, `state`, or lifecycle column), create a minimal state machine with a single state called `existing`:

```json
{
  "states": {
    "existing": {
      "name": "Existing",
      "details": "The entity exists in the system"
    }
  },
  "events": {
    "existing": {
      "name": "existing",
      "details": "Initial event that creates the entity"
    }
  },
  "transitions": [
    {
      "from_state_key": null,
      "to_state_key": "existing",
      "event_key": "existing"
    }
  ]
}
```

This provides a valid state machine even when the class doesn't have explicit lifecycle states.

### Events and Queries from Server Protocols

Examine the system's server protocols (REST APIs, GraphQL, RPC, etc.) to identify events and queries:

| Protocol Call Type | Model Element | Criteria |
|--------------------|---------------|----------|
| **Query** | Query file | Call makes **no change** to system state (GET, read operations) |
| **Event** | Event + Transition + Action | Call **does change** system state (POST, PUT, DELETE, write operations) |

### Actions from State-Changing Calls

For each protocol call that changes system state:

1. **Create an event** in the state machine for the triggering call
2. **Create a transition** from `existing` to `existing` (or between appropriate states if the class has lifecycle states)
3. **Create an action file** that describes the business logic performed by that call

Example: A REST endpoint `POST /orders/{id}/cancel` becomes:

**Event in state_machine.json:**
```json
{
  "events": {
    "cancel": {
      "name": "cancel",
      "details": "Request to cancel the order"
    }
  }
}
```

**Transition in state_machine.json:**
```json
{
  "transitions": [
    {
      "from_state_key": "existing",
      "to_state_key": "existing",
      "event_key": "cancel",
      "action_key": "cancel_order"
    }
  ]
}
```

**Action file (actions/cancel_order.json):**
```json
{
  "name": "Cancel Order",
  "details": "Cancels the order and releases reserved inventory",
  "requires": [
    {"description": "Order has not been shipped"}
  ],
  "guarantees": [
    {"description": "Order status is set to cancelled"},
    {"description": "Reserved inventory is released"},
    {"description": "Customer is notified of cancellation"}
  ]
}
```

### Summary: Mapping System Elements to Model

| Existing System Element | Model Element |
|------------------------|---------------|
| Database table / Cache / Message type | Class |
| Table columns / Fields | Attributes |
| Status/state column values | States |
| Read-only API endpoints | Queries |
| State-changing API endpoints | Events + Transitions + Actions |
| API preconditions | Action `requires` |
| API effects/side effects | Action `guarantees` |

## Design Principles

- **One concept per file** - Each actor, class, etc. is its own JSON file
- **Directory structure reflects hierarchy** - Folders organize content by domain and type
- **Scoped keys in references** - Cross-references use minimal keys scoped to the current context for human readability
- **Minimal required fields** - Only essential fields are required; everything else is optional
- **Separate complex concerns** - State machines, actions, and queries are separate files since they will become complex

## Directory Structure

```
models_json/
└── {model_name}/
    ├── model.json                    # Model metadata
    ├── invariants/                   # Model-level invariants (logic constraints)
    │   ├── 001.invariant.json
    │   ├── 002.invariant.json
    │   └── 003.invariant.json
    ├── actors/
    │   ├── customer.actor.json
    │   ├── manager.actor.json
    │   └── publisher.actor.json
    ├── actor_generalizations/
    │   └── user_type.agen.json
    ├── global_functions/
    │   ├── _max.json
    │   └── _set_of_values.json
    ├── class_associations/           # Model-level associations (cross-domain)
    │   └── orders.default.book_order_line--inventory.default.inventory_item--order_inventory.assoc.json
    ├── domain_associations/          # Domain constraint relationships
    │   └── orders.inventory.domain_assoc.json
    └── domains/
        └── {domain_name}/
            ├── domain.json           # Domain metadata
            ├── class_associations/   # Domain-level associations (cross-subdomain)
            │   └── orders.book_order--shipping.shipment--order_shipment.assoc.json
            └── subdomains/
                └── {subdomain_name}/
                    ├── subdomain.json    # Subdomain metadata
                    ├── class_associations/     # Subdomain-level associations (within subdomain)
                    │   ├── book_order--book_order_line--order_lines.assoc.json
                    │   └── book_order--customer--customer_orders.assoc.json
                    ├── class_generalizations/
                    │   └── medium.cgen.json
                    ├── use_case_generalizations/
                    │   └── order_management.ucgen.json
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
                    └── use_cases/
                        └── {use_case_name}/
                            ├── use_case.json     # Use case metadata
                            └── scenarios/
                                └── happy_path.scenario.json
```

## Filename Conventions

Every entity type has a specific filename pattern. These patterns are enforced by the parser.

| Entity Type | Filename Pattern | Directory | Key Derivation |
|---|---|---|---|
| Model | `model.json` | root | Directory name |
| Invariant | `NNN.invariant.json` | `invariants/` | Sequential number (001, 002, ...) |
| Actor | `{key}.actor.json` | `actors/` | Strip `.actor.json` |
| Actor Generalization | `{key}.agen.json` | `actor_generalizations/` | Strip `.agen.json` |
| Global Function | `{key}.json` | `global_functions/` | Strip `.json` |
| Domain | `domain.json` | `domains/{key}/` | Directory name |
| Subdomain | `subdomain.json` | `domains/{d}/subdomains/{key}/` | Directory name |
| Class | `class.json` | `.../classes/{key}/` | Directory name |
| State Machine | `state_machine.json` | `.../classes/{key}/` | Fixed filename |
| Action | `{key}.json` | `.../classes/{c}/actions/` | Strip `.json` |
| Query | `{key}.json` | `.../classes/{c}/queries/` | Strip `.json` |
| Class Generalization | `{key}.cgen.json` | `.../subdomains/{s}/class_generalizations/` | Strip `.cgen.json` |
| Class Association | `{compound}.assoc.json` | `class_associations/` at model/domain/subdomain level | Compound key (see below) |
| Domain Association | `{key}.domain_assoc.json` | `domain_associations/` | Strip `.domain_assoc.json` |
| Use Case | `use_case.json` | `.../use_cases/{key}/` | Directory name |
| Use Case Generalization | `{key}.ucgen.json` | `.../subdomains/{s}/use_case_generalizations/` | Strip `.ucgen.json` |
| Scenario | `{key}.scenario.json` | `.../use_cases/{uc}/scenarios/` | Strip `.scenario.json` |

## Key Naming Rules

All keys derived from filenames and directory names must follow `snake_case` format:

**Pattern**: `^[a-z][a-z0-9]*(_[a-z0-9]+)*$`

Valid: `order`, `book_order`, `order_line_item`, `customer2`, `v2_order`

Invalid: `BookOrder` (uppercase), `book-order` (hyphen), `2order` (starts with number), `_order` (starts with underscore), `order_` (trailing underscore), `order__line` (consecutive underscores)

**Exception — Global function keys**: Must start with an underscore followed by valid snake_case.

**Pattern**: `^_[a-z][a-z0-9]*(_[a-z0-9]+)*$`

Valid: `_max`, `_set_of_values`, `_calculate_total`

**Exception — Domain association keys**: Two snake_case components separated by a dot.

**Pattern**: `{domain1}.{domain2}` (e.g., `orders.inventory`)

## File Formats

### model.json

```json
{
  "name": "Web Books",
  "details": "An online bookstore application."
}
```

**Fields:**
- `name` (required): Display name of the model
- `details` (optional): Markdown description

### invariants/NNN.invariant.json

Invariants are model-level logical constraints that must always hold true. They are numbered sequentially. Each invariant is a **Logic** object (see [Logic Objects](#logic-objects)).

```json
{
  "description": "An order cannot have a total less than zero",
  "notation": "OCL",
  "specification": "context Order inv: self.total >= 0"
}
```

### actors/{actor_key}.actor.json

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

### actor_generalizations/{key}.agen.json

Actor generalizations define super/sub-type hierarchies between actors.

```json
{
  "name": "User Type",
  "details": "Different types of users in the system",
  "superclass_key": "user",
  "subclass_keys": ["customer", "admin", "manager"],
  "is_complete": true,
  "is_static": true,
  "uml_comment": ""
}
```

**Fields:**
- `name` (required): Display name of the generalization
- `superclass_key` (required): Actor key for the superclass (scoped to model actors)
- `subclass_keys` (required): Array of actor keys for the subclasses (at least one required)
- `details` (optional): Markdown description
- `is_complete` (optional, default false): Are the specializations exhaustive
- `is_static` (optional, default false): Are the specializations unchanging at runtime
- `uml_comment` (optional): Comment for UML diagrams

### global_functions/{key}.json

Global functions are reusable definitions referenced from logic expressions throughout the model. Function names **must start with an underscore**.

```json
{
  "name": "_Max",
  "parameters": ["a", "b"],
  "logic": {
    "description": "Returns the larger of two values",
    "notation": "OCL",
    "specification": "if a > b then a else b endif"
  }
}
```

**Fields:**
- `name` (required): Display name of the function (must start with `_`)
- `parameters` (optional): Array of parameter name strings (each must be non-empty)
- `logic` (required): A Logic object describing the function's behavior (see [Logic Objects](#logic-objects))

### domain_associations/{key}.domain_assoc.json

Domain associations describe constraint relationships between domains (a problem domain enforces requirements on a solution domain).

```json
{
  "problem_domain_key": "orders",
  "solution_domain_key": "inventory",
  "uml_comment": ""
}
```

**Fields:**
- `problem_domain_key` (required): Key of the domain that defines constraints
- `solution_domain_key` (required): Key of the domain that must satisfy constraints
- `uml_comment` (optional): Comment for UML diagrams

### domains/{domain}/domain.json

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

### subdomains/{subdomain}/subdomain.json

```json
{
  "name": "Order Processing",
  "details": "Handles creation and management of orders.",
  "uml_comment": ""
}
```

**Fields:**
- `name` (required): Display name of the subdomain
- `details` (optional): Markdown description
- `uml_comment` (optional): Comment for UML diagrams

**Subdomain Naming Rule:**

- **Single subdomain**: When a domain has exactly one subdomain, it **must** be named `default` (directory: `subdomains/default/`).
- **Multiple subdomains**: When a domain has two or more subdomains, **none** may be named `default`. Each must have a descriptive name.

### classes/{class_name}/class.json

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
      "nullable": false
    },
    "status": {
      "name": "Status",
      "data_type_rules": "enum(pending, confirmed, shipped, delivered)",
      "details": "Current order status",
      "nullable": false
    },
    "total": {
      "name": "Total",
      "data_type_rules": "decimal",
      "details": "Order total amount",
      "derivation_policy": {
        "description": "Sum of line item prices minus discounts plus tax"
      },
      "nullable": false
    }
  },

  "indexes": [
    ["id"],
    ["status", "total"]
  ]
}
```

**Class Fields:**
- `name` (required): Display name of the class
- `details` (optional): Markdown description
- `actor_key` (optional): Actor name (scoped to model actors)
- `uml_comment` (optional): Comment for UML diagrams

**Attribute Fields:**

Attributes are stored as an object keyed by attribute key.

- `name` (required): Display name of the attribute
- `data_type_rules` (optional): Data type specification string
- `details` (optional): Markdown description
- `derivation_policy` (optional): A Logic object describing how this derived attribute is computed (see [Logic Objects](#logic-objects))
- `nullable` (optional, default false): Whether this attribute can be null
- `uml_comment` (optional): Comment for UML diagrams

**Indexes:**

- `indexes` (optional): Array of arrays of attribute keys. Each inner array defines one index.

### class_associations/{compound_key}.assoc.json

Associations exist at three levels depending on which classes they connect:

- **Subdomain-level** (`subdomains/{subdomain}/class_associations/`): Classes within the same subdomain
- **Domain-level** (`domains/{domain}/class_associations/`): Classes from different subdomains within the same domain
- **Model-level** (`class_associations/`): Classes from different domains

The filename includes enough context to be unique at that level. Separators:
- `.` separates domain from subdomain from class in the filename
- `--` separates the from-class, to-class, and distilled name

The distilled name is the full association name, lowercase with `_` between words.

| Level | Filename Pattern | Example |
|-------|-----------------|---------|
| Subdomain | `{from_class}--{to_class}--{name}.assoc.json` | `book_order--book_order_line--order_lines.assoc.json` |
| Domain | `{from_sub}.{from_class}--{to_sub}.{to_class}--{name}.assoc.json` | `orders.book_order--shipping.shipment--order_shipment.assoc.json` |
| Model | `{from_dom}.{from_sub}.{from_class}--{to_dom}.{to_sub}.{to_class}--{name}.assoc.json` | `order_fulfillment.default.book_order_line--inventory.default.inventory_item--order_inventory.assoc.json` |

**Subdomain-level association** (`subdomains/default/class_associations/book_order--customer--customer_orders.assoc.json`):

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

**Domain-level association** (`domains/order_fulfillment/class_associations/orders.book_order--shipping.shipment--order_shipment.assoc.json`):

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

**Model-level association** (`class_associations/order_fulfillment.default.book_order_line--inventory.default.inventory_item--order_inventory.assoc.json`):

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

**Association Fields:**
- `name` (required): Display name of the association
- `from_class_key` (required): Scoped key of the source class
- `from_multiplicity` (required): Multiplicity on the "from" side
- `to_class_key` (required): Scoped key of the target class
- `to_multiplicity` (required): Multiplicity on the "to" side
- `association_class_key` (optional): Scoped key of association class if any
- `details` (optional): Markdown description
- `uml_comment` (optional): Comment for UML diagrams

**Multiplicity Format:**

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

### class_generalizations/{key}.cgen.json

Class generalizations define super/sub-type hierarchies between classes within a subdomain.

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

**Fields:**
- `name` (required): Display name of the generalization
- `superclass_key` (required): Class key for the superclass (scoped to same subdomain)
- `subclass_keys` (required): Array of class keys for the subclasses (at least one required, scoped to same subdomain)
- `details` (optional): Markdown description
- `is_complete` (optional, default false): Are the specializations complete, or can an instantiation exist without a specialization
- `is_static` (optional, default false): Are the specializations static and unchanging, or can they change at runtime
- `uml_comment` (optional): Comment for UML diagrams

### classes/{class_name}/state_machine.json

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
      "details": "Order confirmed and ready for processing"
    },
    "cancelled": {
      "name": "Cancelled",
      "details": "Order was cancelled",
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
        {"name": "items", "data_type_rules": "list of cart items"},
        {"name": "shipping_address", "data_type_rules": "string"}
      ]
    },
    "confirm": {
      "name": "confirm",
      "details": "Order payment confirmed",
      "parameters": [
        {"name": "payment_id", "data_type_rules": "string"}
      ]
    },
    "cancel": {
      "name": "cancel",
      "details": "Order cancelled",
      "parameters": [
        {"name": "reason", "data_type_rules": "string"}
      ]
    }
  },

  "guards": {
    "has_items": {
      "name": "hasItems",
      "logic": {
        "description": "Order has at least one line item"
      }
    },
    "payment_valid": {
      "name": "paymentValid",
      "logic": {
        "description": "Payment has been validated",
        "notation": "OCL",
        "specification": "self.payment.isValid()"
      }
    },
    "not_shipped": {
      "name": "notShipped",
      "logic": {
        "description": "Order has not been shipped yet"
      }
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
      "action_key": "notify_warehouse"
    },
    {
      "from_state_key": "pending",
      "to_state_key": "cancelled",
      "event_key": "cancel",
      "action_key": "refund_payment"
    },
    {
      "from_state_key": "confirmed",
      "to_state_key": "cancelled",
      "event_key": "cancel",
      "guard_key": "not_shipped",
      "action_key": "refund_payment"
    }
  ]
}
```

**State Fields:**

States are stored as an object keyed by state key.

- `name` (required): Display name of the state
- `details` (optional): Markdown description
- `uml_comment` (optional): Comment for UML diagrams
- `actions` (optional): Array of state actions (entry/exit/do)

**State Action Fields:**
- `action_key` (required): Key of the action to execute
- `when` (required): When to execute - `"entry"`, `"exit"`, or `"do"`

**Event Fields:**

Events are stored as an object keyed by event key.

- `name` (required): Display name of the event
- `details` (optional): Description
- `parameters` (optional): Array of parameters (see [Parameter Objects](#parameter-objects))

**Guard Fields:**

Guards are stored as an object keyed by guard key.

- `name` (required): Simple name for internal use
- `logic` (required): A Logic object describing the guard condition (see [Logic Objects](#logic-objects))

**Transition Fields:**
- `from_state_key` (optional): State key to transition from (null for initial transitions)
- `to_state_key` (optional): State key to transition to (null for final transitions)
- `event_key` (required): Event that triggers this transition
- `guard_key` (optional): Guard condition for this transition
- `action_key` (optional): Action to execute during transition
- `uml_comment` (optional): Comment for UML diagrams

Note: At least one of `from_state_key` or `to_state_key` must be specified.

### classes/{class_name}/actions/{action_key}.json

Each action is its own file. The filename (without `.json`) becomes the action key.

```json
{
  "name": "Calculate Total",
  "details": "Sum up line item prices and apply taxes/discounts",
  "parameters": [
    {"name": "discount_code", "data_type_rules": "string"}
  ],
  "requires": [
    {"description": "Order has at least one line item"},
    {"description": "All line items have valid prices"}
  ],
  "guarantees": [
    {"description": "Order.total is set to the computed value"},
    {"description": "Total is non-negative"}
  ],
  "safety_rules": [
    {"description": "Total must not exceed maximum order limit"}
  ]
}
```

**Action Fields:**
- `name` (required): Display name of the action
- `details` (optional): Description of what the action does
- `parameters` (optional): Array of parameters (see [Parameter Objects](#parameter-objects))
- `requires` (optional): Array of Logic objects representing preconditions (see [Logic Objects](#logic-objects))
- `guarantees` (optional): Array of Logic objects representing postconditions
- `safety_rules` (optional): Array of Logic objects representing safety constraints

### classes/{class_name}/queries/{query_key}.json

Each query is its own file. The filename (without `.json`) becomes the query key.

```json
{
  "name": "Get Subtotal",
  "details": "Calculate subtotal before tax and discounts",
  "parameters": [
    {"name": "include_shipping", "data_type_rules": "boolean"}
  ],
  "requires": [
    {"description": "Order exists"}
  ],
  "guarantees": [
    {"description": "Returns sum of (line.price * line.quantity) for all lines"},
    {"description": "Return value is non-negative"}
  ]
}
```

**Query Fields:**
- `name` (required): Display name of the query
- `details` (optional): Description of what the query returns
- `parameters` (optional): Array of parameters (see [Parameter Objects](#parameter-objects))
- `requires` (optional): Array of Logic objects representing preconditions (see [Logic Objects](#logic-objects))
- `guarantees` (optional): Array of Logic objects representing postconditions/return value guarantees

### use_cases/{use_case_key}/use_case.json

Use cases describe user stories for the system at various levels.

```json
{
  "name": "Place Order",
  "details": "Customer places a new order for books",
  "level": "sea",
  "read_only": false,
  "uml_comment": "",
  "actors": {
    "customer": {
      "uml_comment": "Primary actor"
    },
    "inventory_system": {
      "uml_comment": "Supporting actor"
    }
  }
}
```

**Fields:**
- `name` (required): Display name of the use case
- `level` (required): One of `"sky"`, `"sea"`, or `"mud"`
  - `sky` — Strategic level (high-level business goals)
  - `sea` — Tactical level (user-goal level use cases)
  - `mud` — Operational level (sub-function level details)
- `details` (optional): Markdown description
- `read_only` (optional, default false): Whether this use case only reads data
- `uml_comment` (optional): Comment for UML diagrams
- `actors` (optional): Map of class keys to actor reference objects. Each actor reference can have:
  - `uml_comment` (optional): Comment for UML diagrams

### use_case_generalizations/{key}.ucgen.json

Use case generalizations define super/sub-type hierarchies between use cases within a subdomain.

```json
{
  "name": "Order Management",
  "details": "Different types of order operations",
  "superclass_key": "manage_order",
  "subclass_keys": ["place_order", "cancel_order", "modify_order"],
  "is_complete": false,
  "is_static": true,
  "uml_comment": ""
}
```

**Fields:**
- `name` (required): Display name of the generalization
- `superclass_key` (required): Use case key for the parent (scoped to same subdomain)
- `subclass_keys` (required): Array of use case keys for the children (at least one required)
- `details` (optional): Markdown description
- `is_complete` (optional, default false): Are the specializations exhaustive
- `is_static` (optional, default false): Are the specializations unchanging at runtime
- `uml_comment` (optional): Comment for UML diagrams

### scenarios/{scenario_key}.scenario.json

Scenarios document specific flows through a use case (e.g., sequence diagrams). They live inside a use case directory.

```json
{
  "name": "Happy Path",
  "details": "Customer successfully places an order",
  "objects": {
    "c1": {
      "object_number": 1,
      "name": "Alice",
      "name_style": "instance",
      "class_key": "customer",
      "multi": false,
      "uml_comment": ""
    },
    "o1": {
      "object_number": 2,
      "name_style": "anonymous",
      "class_key": "book_order",
      "multi": false
    }
  },
  "steps": {
    "step_type": "sequence",
    "statements": [
      {
        "step_type": "leaf",
        "leaf_type": "event",
        "description": "Customer places an order",
        "from_object_key": "c1",
        "to_object_key": "o1",
        "event_key": "place"
      },
      {
        "step_type": "leaf",
        "leaf_type": "query",
        "description": "Check order total",
        "from_object_key": "c1",
        "to_object_key": "o1",
        "query_key": "get_subtotal"
      }
    ]
  }
}
```

**Scenario Fields:**
- `name` (required): Display name of the scenario
- `details` (optional): Markdown description
- `objects` (optional): Map of object keys to object definitions
- `steps` (optional): Root step node (recursive AST structure)

**Object Fields:**
- `object_number` (required): Sequential number for ordering in diagrams
- `name` (optional): Instance name (e.g., "Alice")
- `name_style` (required): How the name is displayed — `"instance"`, `"anonymous"`, or `"class"`
- `class_key` (required): Key of the class this object represents (scoped to subdomain)
- `multi` (optional, default false): Whether this represents multiple objects
- `uml_comment` (optional): Comment for UML diagrams

**Step Fields (recursive AST):**

Steps form a tree structure representing the scenario flow. There are two categories:

**Container steps** (have `statements` children):
- `step_type: "sequence"` — Execute statements in order
- `step_type: "loop"` — Repeat statements while condition holds
  - `condition` (required): Loop condition
  - `statements`: Body of the loop
- `step_type: "switch"` — Branch on conditions
  - `statements`: Array of `case` steps
- `step_type: "case"` — A branch within a switch
  - `condition` (required): Case condition
  - `statements`: Body of the case

**Leaf steps** (`step_type: "leaf"`):
- `leaf_type: "event"` — An event interaction between objects
  - `from_object_key`: Sending object
  - `to_object_key`: Receiving object
  - `event_key`: Key of the event on the target object's state machine
  - `description`: What happens
- `leaf_type: "query"` — A query interaction between objects
  - `from_object_key`: Querying object
  - `to_object_key`: Queried object
  - `query_key`: Key of the query on the target object
  - `description`: What is being queried
- `leaf_type: "scenario"` — A reference to another scenario
  - `scenario_key`: Key of the referenced scenario
  - `description`: What the sub-scenario does
- `leaf_type: "delete"` — Deletion of an object
  - `from_object_key`: Object initiating deletion
  - `to_object_key`: Object being deleted
  - `description`: What is being deleted

## Shared Objects

### Logic Objects

Logic objects appear throughout the model wherever formal or informal logic needs to be expressed: invariants, action requires/guarantees/safety_rules, query requires/guarantees, guard conditions, global function definitions, and attribute derivation policies.

```json
{
  "description": "Order total must be non-negative",
  "notation": "OCL",
  "specification": "context Order inv: self.total >= 0"
}
```

**Fields:**
- `description` (required): Human-readable description of the logic
- `notation` (optional): Formal notation system (e.g., `"OCL"`, `"Z"`, `"TLA+"`)
- `specification` (optional): Formal specification in the given notation

### Parameter Objects

Parameters appear in actions, queries, and events.

```json
{
  "name": "discount_code",
  "data_type_rules": "string, max 20 characters"
}
```

**Fields:**
- `name` (required): Parameter name
- `data_type_rules` (optional): Data type specification string

## The "details" Field: Purpose and Proper Use

**IMPORTANT**: The `details` field that appears on many entities is **emphatically NOT for describing logic**. It serves a specific, limited purpose.

### What "details" Is For

The `details` field provides a **human-readable summary** — a brief, plain-language description that helps readers understand what something is or does at a high level. Think of it as a caption or tooltip.

Good examples:
- `"details": "Represents a customer's order for books."`
- `"details": "Handles the lifecycle of customer orders."`
- `"details": "Links an order to the customer who placed it"`
- `"details": "Check if order can still be cancelled"`

### What "details" Is NOT For

The `details` field should **never** contain:
- Preconditions or prerequisites
- Postconditions or guarantees
- Logical rules or business logic
- Implementation details
- Conditional behavior ("if X then Y")
- Formulas or calculations

**Wrong** (logic in details):
```json
{
  "name": "calculateTotal",
  "details": "If order has items, sum their prices. Must have at least one item. Sets Order.total to computed value which must be non-negative."
}
```

**Correct** (logic in structured fields):
```json
{
  "name": "calculateTotal",
  "details": "Sum up line item prices and apply taxes/discounts",
  "requires": [
    {"description": "Order has at least one line item"},
    {"description": "All line items have valid prices"}
  ],
  "guarantees": [
    {"description": "Order.total is set to the computed value"},
    {"description": "Total is non-negative"}
  ]
}
```

### Where Logic Belongs

Each entity type has appropriate structured fields for logic:

| Entity | Summary Field | Logic Fields |
|--------|---------------|--------------|
| Action | `details` | `requires`, `guarantees`, `safety_rules` (arrays of Logic objects) |
| Query | `details` | `requires`, `guarantees` (arrays of Logic objects) |
| Guard | — | `logic` (single Logic object) |
| Invariant | — | Is itself a Logic object (`description`, `notation`, `specification`) |
| Global Function | — | `logic` (single Logic object) |
| Attribute | `details` | `derivation_policy` (single Logic object, optional) |
| State | `details` | `actions` (entry/exit/do behaviors) |
| Event | `details` | `parameters` |
| Class | `details` | `attributes`, `indexes` |
| Association | `details` | `from_multiplicity`, `to_multiplicity` |

### Why This Matters

1. **AI Processing**: Structured Logic objects are machine-parseable. AI can extract preconditions and postconditions reliably.

2. **Code Generation**: Generators can use `requires` for validation checks and `guarantees` for assertions. Free-form text in `details` cannot be reliably transformed.

3. **Documentation**: The `details` field appears in documentation as a brief description. Stuffing logic into it makes documentation unreadable.

4. **Validation**: Structured fields can be validated (e.g., "does the action guarantee something about attributes?"). Text in `details` cannot.

5. **Consistency**: Keeping logic in structured fields ensures all actions/queries follow the same pattern.

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

1. **Subdomain size**: Subdomains should contain between 20-40 classes. When a domain grows beyond 40 classes, consider splitting it into multiple subdomains to maintain manageable, cohesive groupings. When starting a new domain, begin with a single subdomain and split when needed.

2. **Actions are independent**: Actions do not reference or call other actions.

3. **Subdomain naming**: A domain with exactly one subdomain must name it `default`. A domain with multiple subdomains cannot use `default` as any of their names.

4. **Model completeness**: A valid model must have at least one actor and at least one domain. Each domain must have at least one subdomain.

## Implementation Architecture

### Separate Go Structs for JSON Import

The JSON import package uses its own set of Go structs, separate from the `req_model` classes. This separation exists for several reasons:

1. **Optimized input shapes**: The JSON format can use structures that are easier for input (e.g., superclass/subclass defined in generalization rather than spread across class files).

2. **Distinct error handling**: Each validation error has two distinct components:
   - **Unique error number**: Every error type has its own identifier for programmatic handling and documentation reference.
   - **Detailed output with construction advice**: Error messages include specific guidance on how to correct the input, helping AI or human authors fix issues quickly.

   These error types are specific to import validation and don't belong in the main model tree.

3. **Clear separation of concerns**: Import structs handle parsing and validation of external input; `req_model` structs represent the canonical internal model.

4. **Conversion layer**: After successful validation, import structs are converted to `req_model` structs for use in the rest of the system.

### JSON Schema Validation

Each JSON file type has a corresponding JSON Schema for validation:

- `model.schema.json`
- `actor.schema.json`
- `actor_generalization.schema.json`
- `domain.schema.json`
- `subdomain.schema.json`
- `class.schema.json`
- `class_association.schema.json`
- `class_generalization.schema.json`
- `state_machine.schema.json`
- `action.schema.json`
- `query.schema.json`
- `logic.schema.json`
- `parameter.schema.json`
- `global_function.schema.json`
- `domain_association.schema.json`
- `use_case.schema.json`
- `use_case_generalization.schema.json`
- `use_case_shared.schema.json`
- `scenario.schema.json`

These schemas provide:
- Early validation before Go parsing
- Clear documentation of expected structure
- IDE support for editing JSON files
- Consistent error messages for structural issues

### Cross-Reference Validation

After parsing, the entire model tree is validated for internal consistency:

- All class `actor_key` references must point to existing actors
- All association `from_class_key`, `to_class_key`, and `association_class_key` references must point to existing classes at the appropriate scope
- All class generalization `superclass_key` and `subclass_keys` must point to existing classes in the same subdomain
- All actor generalization `superclass_key` and `subclass_keys` must point to existing actors
- All domain association `problem_domain_key` and `solution_domain_key` must point to existing domains
- State machine transition references to states, events, guards, and actions must all be valid
- Use case actor references must point to existing classes in the same subdomain
- Scenario event and query references must point to valid events/queries on the referenced objects' classes
