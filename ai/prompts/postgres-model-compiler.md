# Postgres Model Compiler

## Purpose

Build a model compiler that takes two JSON files as input and produces a complete, from-scratch PostgreSQL schema as output.

1. **The model file** — a pure platform-independent semantic model (PIM). Classes, attributes, data types, associations, state machines, generalizations, invariants. No Postgres-specific content.
2. **The compilation spec file** — selects which model elements go into this database and provides Postgres-specific mappings (column types, defaults, ON DELETE behavior, schema name).

Only model elements mentioned in the compilation spec are included in the output. The same model file can be paired with different compilation specs to produce different databases.

The database is a storage and constraint layer. It enforces data structure, data types, foreign keys, and straightforward value constraints. Business logic — state transitions, guards, actions, derived computations — is the responsibility of the server and client that talk to the database, not the database itself.

The output is always a complete schema meant to create a database from nothing. A separate future tool will handle comparing a generated schema against an existing database to produce migration steps.

---

## File 1: The Model (PIM)

A pure semantic model. No platform-specific content. This file is shared across all compilation targets.

```json
{
  "model": {
    "name": "Order Fulfillment",
    "details": "A model for an online bookstore order processing system."
  },
  "data_types": {
    "valid_email_address": {
      "details": "An email address.",
      "constraint": { "type": "span" }
    },
    "valid_copyright_year": {
      "details": "A copyright year.",
      "constraint": { "type": "span", "min": 1780, "min_inclusive": true }
    },
    "money_amount": {
      "details": "A monetary amount, non-negative.",
      "constraint": { "type": "span", "min": 0, "min_inclusive": true }
    },
    "positive_integer": {
      "details": "An integer that is at least 1.",
      "constraint": { "type": "span", "min": 1, "min_inclusive": true }
    },
    "unconstrained_text": {
      "details": "Unconstrained text.",
      "constraint": { "type": "unconstrained" }
    }
  },
  "classes": {
    "customer": {
      "details": "A registered customer.",
      "attributes": {
        "email": { "data_type": "valid_email_address", "nullable": false, "details": "Unique email." },
        "name": { "data_type": "unconstrained_text", "nullable": false, "details": "Full name." },
        "postal_address": { "data_type": "unconstrained_text", "nullable": false, "details": "Mailing address." }
      },
      "indexes": [
        { "columns": ["email"], "unique": true }
      ]
    },
    "book_order": {
      "details": "A customer order for one or more books.",
      "attributes": {
        "id": { "data_type": "unconstrained_text", "nullable": false, "details": "Unique identifier." },
        "date_opened": { "data_type": "unconstrained_text", "nullable": false, "details": "Date created." },
        "shipping_address": { "data_type": "unconstrained_text", "nullable": true, "details": "Postal address for shipping." }
      },
      "indexes": [
        { "columns": ["id"], "unique": true }
      ],
      "state_machine": {
        "states": {
          "open": { "details": "Order created but not placed." },
          "placed": { "details": "Submitted for fulfillment." },
          "packed": { "details": "All items packed." },
          "shipped": { "details": "Shipped." },
          "completed": { "details": "Delivered." },
          "cancelled": { "details": "Cancelled." }
        },
        "initial_state": "open"
      }
    },
    "title": {
      "details": "A book title in the catalog.",
      "attributes": {
        "isbn": { "data_type": "unconstrained_text", "nullable": false, "details": "ISBN." },
        "name": { "data_type": "unconstrained_text", "nullable": false, "details": "Title name." },
        "price": { "data_type": "money_amount", "nullable": false, "details": "Retail price." },
        "stock_level": { "data_type": "positive_integer", "nullable": false, "details": "Current stock." }
      },
      "indexes": [
        { "columns": ["isbn"], "unique": true }
      ]
    },
    "book_order_line": {
      "details": "A line item on a book order.",
      "attributes": {
        "qty": { "data_type": "positive_integer", "nullable": false, "details": "Number of copies." }
      }
    }
  },
  "associations": {
    "customer_places_book_order": {
      "details": "A customer places zero or more book orders.",
      "from_class": "customer",
      "from_multiplicity": { "min": 1, "max": 1 },
      "to_class": "book_order",
      "to_multiplicity": { "min": 0, "max": null }
    },
    "book_order_has_lines": {
      "details": "A book order has one or more lines.",
      "from_class": "book_order",
      "from_multiplicity": { "min": 1, "max": 1 },
      "to_class": "book_order_line",
      "to_multiplicity": { "min": 1, "max": null }
    },
    "book_order_line_for_title": {
      "details": "Each order line references a title.",
      "from_class": "book_order_line",
      "from_multiplicity": { "min": 0, "max": null },
      "to_class": "title",
      "to_multiplicity": { "min": 1, "max": 1 }
    }
  },
  "generalizations": {
    "medium_type": {
      "details": "A title can be sold as different medium types.",
      "is_complete": true,
      "is_static": true,
      "superclass": "medium",
      "subclasses": ["print_medium", "ebook_medium"]
    }
  }
}
```

Key points:
- Maps keyed by name so the compilation spec can reference elements directly.
- `constraint.type`: `"span"`, `"enumeration"`, or `"unconstrained"`. Spans have optional `min`, `max`, `min_inclusive`, `max_inclusive`. Enumerations have `values` and `ordered`.
- State machines define states and an initial state. Events, guards, actions, and transitions are part of the model but are not relevant to database generation — the server handles behavioral logic.
- Associations have `from_class`, `to_class`, multiplicities (`min`/`max`, where `max: null` means unbounded), and optional `association_class`.
- Generalizations have `is_complete`, `is_static`, `superclass`, `subclasses`.

---

## File 2: The Compilation Spec

Selects which model elements to include and provides Postgres-specific mappings. No SQL expressions — the compiler generates all SQL from the model's formal definitions and these mappings.

```json
{
  "schema": "order_fulfillment",
  "details": "OLTP schema for the order fulfillment domain.",

  "data_types": {
    "valid_email_address": { "type": "text" },
    "valid_copyright_year": { "type": "integer" },
    "money_amount": { "type": "numeric(12,2)" },
    "positive_integer": { "type": "integer" },
    "unconstrained_text": { "type": "text" }
  },

  "classes": {
    "customer": {
      "attributes": {
        "email": {},
        "name": {},
        "postal_address": {}
      }
    },
    "book_order": {
      "attributes": {
        "id": {},
        "date_opened": { "default": "CURRENT_DATE" },
        "shipping_address": {}
      }
    },
    "title": {
      "attributes": {
        "isbn": {},
        "name": {},
        "price": {},
        "stock_level": {}
      }
    },
    "book_order_line": {
      "attributes": {
        "qty": {}
      }
    }
  },

  "associations": {
    "customer_places_book_order": { "on_delete": "CASCADE" },
    "book_order_has_lines": { "on_delete": "CASCADE" },
    "book_order_line_for_title": { "on_delete": "RESTRICT" }
  },

  "generalizations": {
    "medium_type": {}
  }
}
```

Key points:
- Every key references a name from the model file. If a model element is not mentioned here, it is excluded from the output.
- An empty object `{}` means "include with defaults."
- `data_types`: `type` is the Postgres column type. Omit `type` for enumerations (compiler creates `CREATE TYPE ... AS ENUM` from the model's `values`). The compiler generates CHECK constraints from the model's span bounds — no SQL expressions needed here.
- `classes`: lists which attributes to include. Unlisted attributes are excluded.
- `attributes`: `default` is a raw SQL default expression. This is the one place where SQL appears — defaults are inherently platform-specific (e.g., `CURRENT_DATE`, `now()`, `gen_random_uuid()`).
- `associations`: `on_delete` is the FK `ON DELETE` behavior (`CASCADE`, `RESTRICT`, `SET NULL`, etc.).
- `generalizations`: `{}` means include with standard class-table inheritance.

---

## What the Compiler Generates

The compiler produces structure and constraints, not behavior. The database enforces what data looks like, not what the application does with it.

### What is generated

- **Tables** with typed columns, NOT NULL, and defaults.
- **Primary keys** — synthetic `pk bigint GENERATED ALWAYS AS IDENTITY` on every table.
- **Timestamps** — `created_at` and `updated_at` on every table, with an `updated_at` trigger.
- **State columns** — for classes with state machines, a column using a Postgres ENUM type, defaulting to the model's `initial_state`. The database stores the state; the server manages transitions.
- **ENUM types** — for enumeration data types and state machine states.
- **Domain types** — for span-constrained data types, with CHECK constraints generated from the model's `min`/`max`/`min_inclusive`/`max_inclusive`.
- **Foreign keys** — from associations. Multiplicity determines nullability and which side holds the FK.
- **Join tables** — for many-to-many associations.
- **Indexes** — from the model's index definitions.
- **Generalization structure** — class-table inheritance with `type` discriminator on superclass, subclass tables with FK to superclass PK.
- **Comments** — `COMMENT ON` for every generated object, from the model's `details` fields.

### What is NOT generated

- **Event procedures** — the server handles state transitions.
- **Guard logic** — the server evaluates guards before transitions.
- **Action logic** — the server executes business logic.
- **Derived attribute functions** — the server computes these.
- **State transition enforcement triggers** — the server validates transitions.
- **Constructor/destructor procedures** — the server does inserts and deletes.
- **Cross-row or cross-table invariant triggers** — the server enforces complex invariants.

---

## Output: Complete Schema

A single `.sql` file with clean `CREATE` statements. No `CREATE OR REPLACE`, no `DROP IF EXISTS`, no `IF NOT EXISTS`. Runs against an empty schema.

### Output Sections (in order)

**1. Preamble** — `BEGIN`, `CREATE SCHEMA`, `SET search_path`.

**2. ENUM types** — For enumeration data types and state machine states.

```sql
CREATE TYPE book_order_state AS ENUM ('open', 'placed', 'packed', 'shipped', 'completed', 'cancelled');
COMMENT ON TYPE book_order_state IS 'Lifecycle states for book_order.';
```

**3. Domain types** — For span-constrained data types. CHECK generated from model bounds.

```sql
CREATE DOMAIN money_amount AS numeric(12,2) CHECK (VALUE >= 0);
COMMENT ON DOMAIN money_amount IS 'A monetary amount, non-negative.';

CREATE DOMAIN positive_integer AS integer CHECK (VALUE >= 1);
COMMENT ON DOMAIN positive_integer IS 'An integer that is at least 1.';
```

**4. Tables** — Ordered: superclass before subclass, referenced before referencing where possible.

```sql
CREATE TABLE book_order (
  pk bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  id text NOT NULL,
  date_opened text NOT NULL DEFAULT CURRENT_DATE,
  shipping_address text,
  state book_order_state NOT NULL DEFAULT 'open',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE book_order IS 'A customer order for one or more books.';
COMMENT ON COLUMN book_order.pk IS 'Synthetic primary key.';
COMMENT ON COLUMN book_order.id IS 'Unique identifier.';
COMMENT ON COLUMN book_order.date_opened IS 'Date created.';
COMMENT ON COLUMN book_order.shipping_address IS 'Postal address for shipping.';
COMMENT ON COLUMN book_order.state IS 'Current lifecycle state.';
COMMENT ON COLUMN book_order.created_at IS 'Row creation timestamp.';
COMMENT ON COLUMN book_order.updated_at IS 'Row last-modified timestamp.';
```

**5. Indexes.**

```sql
CREATE UNIQUE INDEX idx_book_order_id ON book_order (id);
```

**6. Foreign keys** — FK columns and constraints from associations.

```sql
ALTER TABLE book_order ADD COLUMN customer_pk bigint NOT NULL;
ALTER TABLE book_order ADD CONSTRAINT fk_book_order_customer
  FOREIGN KEY (customer_pk) REFERENCES customer (pk) ON DELETE CASCADE;
COMMENT ON COLUMN book_order.customer_pk IS 'FK: customer_places_book_order.';
```

Join tables for many-to-many associations.

**7. Generalization structure** — `type` discriminator on superclass, subclass tables with FK to superclass PK. Completeness enforced via deferred constraint trigger if `is_complete`. Staticness enforced via before-update trigger if `is_static`.

**8. Updated-at trigger** — One shared trigger function, applied to every table.

```sql
CREATE FUNCTION trg_set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at BEFORE UPDATE ON book_order
  FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();
```

**9. Commit.**

---

## Compiler Requirements

### Input Validation

The compiler merges and validates both files:

1. Every name in the compilation spec must exist in the model file.
2. All `data_type` references used by included attributes resolve to included data types.
3. Associations reference only included classes.
4. Generalizations reference only included classes.
5. No duplicate names within scope.

### CHECK Generation from Spans

The compiler generates CHECK constraints from model span bounds:
- `min` + `min_inclusive: true` → `VALUE >= min`
- `min` + `min_inclusive: false` → `VALUE > min`
- `max` + `max_inclusive: true` → `VALUE <= max`
- `max` + `max_inclusive: false` → `VALUE < max`
- Both min and max → combined with `AND`.
- No bounds → no CHECK (unconstrained span).

### Association Implementation

- `(X, 1)` to-one: FK column on the "many" side. `NOT NULL` if min=1, nullable if min=0.
- Many-to-many (`max: null` on both sides): Join table `{from_class}_{to_class}`.
- `association_class`: The named class's table serves as the join table.

### Generalization Implementation

Class-table inheritance. Superclass table holds shared columns and PK. Each subclass table has subclass-specific columns and a FK to the superclass PK. `type` discriminator column on superclass. If `is_complete`, deferred constraint trigger ensures a subclass row exists. If `is_static`, before-update trigger prevents `type` changes.

### Output Ordering

Topologically sorted: schema, ENUMs, DOMAINs, tables (dependency order), FKs, indexes, generalization structure, triggers, commit.

### Naming

- Tables/columns: `snake_case` matching model names.
- FK columns: `{referenced_table}_pk`.
- Join tables: `{from_class}_{to_class}`.
- State enums: `{class}_state`.
- All identifiers: valid Postgres identifiers.
