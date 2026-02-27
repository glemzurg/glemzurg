# Integration Assessment: model_expression and model_data_type

## Problem Statement

The requirements model has two parallel type systems that describe the same underlying reality from different perspectives:

- **model_data_type**: The stakeholder-facing view. Describes what values look like in business terms — enumerations of allowed values, numeric ranges with units, collections with cardinality constraints, records with named fields. Designed for readability by non-technical stakeholders.

- **model_expression**: The code-generation-facing view. Describes precisely how values are computed, compared, and transformed — arithmetic on integers, set membership tests, record field access, quantified predicates. Designed for unambiguous machine interpretation.

Today these two systems are structurally independent. An `AttributeRef` expression node points at an attribute by key, and that attribute has a `DataType`, but the expression system has no awareness of what that data type is. A `BuiltinCall` to `_Stack!Push` operates on what the evaluator treats as a tuple, but nothing connects it to the attribute's declared `stack` collection type. The type checker infers types through Hindley-Milner unification at runtime, but the inferred types are ephemeral — they exist only during type-checking and are never stored or compared against declared data types.

This document proposes a design for tying these two systems together.

---

## Current Architecture

### model_data_type

A `DataType` has a `CollectionType` (atomic, ordered, unordered, queue, stack, record) and, for atomic types, a `ConstraintType` (unconstrained, span, enumeration, reference, object). Collections have optional cardinality bounds and uniqueness flags. Records have named fields, each with its own nested `DataType`.

DataTypes are attached to:
- **Attributes** via `Attribute.DataType` (class-level state)
- **Parameters** via `Parameter.DataType` (action/query/event inputs)

DataTypes live in the database across five tables: `data_type`, `data_type_atomic`, `data_type_atomic_span`, `data_type_atomic_enum_value`, `data_type_field`.

### model_expression

An `Expression` is a tree of nodes. Leaf nodes are literals (`BoolLiteral`, `IntLiteral`, `SetLiteral`, etc.) or references (`AttributeRef`, `LocalVar`, `SelfRef`). Interior nodes are operations (`BinaryArith`, `Compare`, `SetOp`, `FieldAccess`, `BuiltinCall`, etc.).

Expressions are attached to `Logic` objects as an optional `Expression` field. They live in the database in the `expression_node` table (adjacency list).

### The Gap

There is no formal connection between a `DataType` and an expression's implicit type. The expression `self.amount + 10` implicitly produces a number, and the attribute `amount` might be declared as `[0 .. 1000] at 1.0 dollar`, but nothing in the model ties these together. The type checker infers "number" during simulation, but this inference is:
1. Ephemeral (not stored)
2. Coarse (no distinction between integer, rational, or bounded spans)
3. Disconnected from stakeholder-declared constraints

---

## Design Goals

### Three Valid States for Type Information

Every typed element (attribute, parameter, query result) should support exactly one of:

1. **Stakeholder type only** — DataType exists, no precise ExpressionType. The requirements author has described what the stakeholder sees but hasn't written the precise formal definition yet. This is the normal starting state.

2. **Both stakeholder and precise types** — DataType exists and ExpressionType exists. The requirements author has provided both the stakeholder view and the code-generation view. The system can validate compatibility.

3. **Precise type only** — No DataType, but ExpressionType exists. The requirements author needed a type for the formal model (e.g., a query result type, an intermediate computation type) that has no stakeholder representation. These are implementation-facing types.

### What "Precise Type" Means

The precise type is the code-generation-complete type. Where a stakeholder type might say "a set of unconstrained", the precise type says "a set of integers". Where a stakeholder type says "ordered collection", the precise type says "a sequence of records with fields {name: string, amount: integer}".

The precise type must be fully resolved — no "unconstrained" leaves. Every element type, field type, and return type must be concrete.

### Compatibility Rules

When both stakeholder and precise types exist, they must be compatible. The `CollectionUnique` flag on DataType is critical for determining the structural type:

| Stakeholder CollectionType | CollectionUnique | Compatible Precise Types | TLA+ Constructor |
|---|---|---|---|
| `atomic` | n/a | Any non-collection precise type (boolean, integer, rational, string, enum) | `Nat`, `STRING`, etc. |
| `ordered` | true | `SequenceType(T, Unique: true)` or `SetType(T)` | `_SeqUnique(T)` or `_Set(T)` / `SUBSET T` |
| `ordered` | false | `SequenceType(T, Unique: false)` | `Seq(T)` or `_Seq(T)` |
| `unordered` | true | `SetType(T)` | `SUBSET T` / `_Set(T)` |
| `unordered` | false | `BagType(T)` | `_Bag(T)` |
| `stack` | true | `SequenceType(T, Unique: true)` | `_StackUnique(T)` |
| `stack` | false | `SequenceType(T, Unique: false)` | `_Stack(T)` |
| `queue` | true | `SequenceType(T, Unique: true)` | `_QueueUnique(T)` |
| `queue` | false | `SequenceType(T, Unique: false)` | `_Queue(T)` |
| `record` | n/a | `RecordType({field: Type, ...})` where field names match | `[f1: T1, f2: T2]` |

The key insight: the `CollectionUnique` flag on DataType maps to either the inherent uniqueness of `SetType`, or the `Unique` flag on `SequenceType`. For `unordered` collections, unique maps to `SetType` (a set IS an unordered unique collection) and non-unique maps to `BagType` (a bag IS an unordered non-unique collection). For `ordered`/`stack`/`queue` collections, the uniqueness is carried as a property on `SequenceType` — the structural type is always sequence, but the `Unique` flag tells code generators to enforce distinctness on insertion. An `ordered` + `unique` stakeholder type can also map to `SetType` when the author decides the ordering is not structurally meaningful — the author chooses the precise type that best matches code generation needs.

Additional compatibility:
- Stakeholder `span` constraints define a subset of the precise numeric type
- Stakeholder `enumeration` constraints define a subset of the precise string or integer type
- Stakeholder cardinality bounds (min/max) must be satisfiable by the precise type
- Stakeholder `object` constraint means the precise type is a record matching the referenced class

### Should the Precise Type Distinguish Sequence, Stack, and Queue?

No. At the structural level, sequence, stack, and queue are all the same thing — an ordered collection of elements. The difference is the *access pattern*:

- **Sequence**: random access, append, prepend, concatenation
- **Stack**: push (prepend), pop (head) — LIFO
- **Queue**: enqueue (append), dequeue (head) — FIFO

These access patterns are enforced by which builtin operations appear in expressions (`_Stack!Push` vs `_Seq!Append` vs `_Queue!Enqueue`), not by the type itself. A stack is structurally a sequence; the "stackness" is a behavioral constraint, not a type distinction. Making them distinct precise types would:

1. Force artificial type conversions when passing a stack to a function that operates on sequences
2. Duplicate every sequence operation for stacks and queues
3. Complicate type inference — `_Seq!Len` works on all three, so what type does it expect?

The stakeholder's collection kind already captures the intended access pattern. The precise type captures the structural reality. These are complementary, not redundant.

---

## Proposed Design: ExpressionType

### New Package: `model_expression_type`

A new package `internal/req_model/model_expression_type/` defines precise types as a closed set of concrete type constructors. Unlike `model_data_type` (which is parsed from free-form text with business constraints), `ExpressionType` is a structural algebraic type system designed for machine consumption.

```
ExpressionType (interface)
├── BooleanType          — boolean values
├── IntegerType          — arbitrary-precision integers
├── RationalType         — rational numbers (numerator/denominator)
├── StringType           — character strings
├── EnumType             — finite set of named values {Values []string}
├── SetType              — unordered unique collection {ElementType ExpressionType}
├── SequenceType         — ordered collection {ElementType ExpressionType, Unique bool}
├── BagType              — multiset {ElementType ExpressionType}
├── TupleType            — fixed-length heterogeneous {ElementTypes []ExpressionType}
├── RecordType           — named fields {Fields []RecordFieldType}
├── FunctionType         — callable {Params []ExpressionType, Return ExpressionType}
└── ObjectType           — class instance reference {ClassKey identity.Key}
```

Note: `SequenceType` carries a `Unique` flag that signals whether elements must be distinct. This flag is set by the constructor used in the TLA+ specification: `Seq(S)` / `_Seq(S)` → `Unique: false`, `_SeqUnique(S)` / `_StackUnique(S)` / `_QueueUnique(S)` → `Unique: true`. Sets are inherently unique (no flag needed). Bags are inherently non-unique (no flag needed).

Each type implements:
- `expressionType()` — marker method
- `TypeName() string` — canonical name for serialization
- `Validate() error` — structural validation

### Where ExpressionType Lives

ExpressionType is attached to **DataType** via a `TypeSpec` (the reusable Notation+Specification+ParsedTree trio — see "How Expression Types Get Authored"). Attributes and parameters already link to DataType — adding a TypeSpec to DataType means they get the precise type through their existing relationship:

**DataType:**
```go
type DataType struct {
    CollectionType   string
    CollectionUnique *bool
    CollectionMin    *int
    CollectionMax    *int
    Atomic           *Atomic
    RecordFields     []RecordFieldDataType
    TypeSpec         *TypeSpec   // NEW: optional precise type (Notation + Specification + ExpressionType)
}
```

This means:
- An Attribute with a DataType automatically has access to the precise type via `attr.DataType.TypeSpec.ExpressionType`
- A Parameter with a DataType gets it the same way: `param.DataType.TypeSpec.ExpressionType`
- The three valid states still hold — DataType can exist without TypeSpec (stakeholder only), with TypeSpec (both), or TypeSpec can exist without business constraints (precise only, using a DataType with `ConstraintType: "unconstrained"` as the carrier)
- The compatibility checker validates the ExpressionType against its own DataType — the relationship is co-located, not cross-referenced

**Logic.TargetType (separate concern):**

Query guarantees and state change guarantees use `Logic.Target` as a string identifier for the output binding. A `TargetTypeSpec` declares the type of what the logic expression produces, providing a convenient gate for verifying that the expression result is well-formed before the simulator ever runs:

```go
type Logic struct {
    Key            identity.Key
    Type           string
    Description    string
    Target         string
    Spec           ExpressionSpec   // Notation + Specification + Expression (the reusable trio)
    TargetTypeSpec *TypeSpec        // Optional: type of Logic.Target (Notation + Specification + ExpressionType)
}
```

Note that Logic itself uses `ExpressionSpec` (the behavioral trio: Notation + Specification + Expression tree) while `TargetTypeSpec` uses `TypeSpec` (the type trio: Notation + Specification + ExpressionType tree). Both follow the same pattern — the only difference is what the parsed tree represents.

TargetTypeSpec is separate from DataType's TypeSpec because it describes the **output of the expression**, not the type of a stored value. The system can then verify: does the expression's result type match the TargetTypeSpec? And does the TargetTypeSpec match the target attribute's DataType.TypeSpec.ExpressionType?

### Why a Separate Type System (Not Reusing DataType)

`model_data_type.DataType` is the wrong abstraction for precise types because:

1. **It conflates structure with constraints.** A span `[0..100] at 0.01 dollar` mixes the structural type (rational number) with business constraints (range, precision, units). The expression system needs the structural type without the constraints.

2. **It lacks key structural types.** DataType has no concept of boolean, integer vs. rational, tuple (fixed-length heterogeneous), bag/multiset, or function type. These are fundamental to expression evaluation.

3. **Its collection types encode access patterns, not structure.** Stack and queue are sequences with restricted operations, not distinct structural types. The expression system should use `SequenceType` for both, with the access pattern encoded elsewhere.

4. **Its parser is designed for human authoring.** The PEG grammar parses stakeholder-friendly text like `"3+ ordered of enum of small, medium, large"`. Precise types need machine-friendly construction, not text parsing.

5. **It tolerates ambiguity.** `"unconstrained"` is a valid atomic constraint — it means "we don't know yet." Precise types must be fully resolved by definition.

### Compatibility Validation

A new function validates that a DataType and ExpressionType are compatible:

```go
func CheckCompatibility(
    dt *model_data_type.DataType,
    et model_expression_type.ExpressionType,
) []CompatibilityIssue
```

This checks:
- Collection type alignment (ordered↔Sequence, unordered↔Set, etc.)
- Record field name matching
- Cardinality bounds satisfiability
- Span subset inclusion (if the span bounds are within the integer/rational range)
- Enum value matching (stakeholder enum values must be a subset of the precise enum values)
- Object class key matching

Compatibility issues are warnings, not errors. This allows incremental refinement — a stakeholder type can be less specific than the precise type.

---

## Database Design

### New Enum and Table

```sql
CREATE TYPE expression_type_kind AS ENUM (
    'boolean',
    'integer',
    'rational',
    'string',
    'enum',
    'set',
    'sequence',
    'bag',
    'tuple',
    'record',
    'function',
    'object'
);

CREATE TABLE expression_type (
    model_key            text NOT NULL,
    expression_type_key  text NOT NULL,
    parent_type_key      text DEFAULT NULL,
    sort_order           int NOT NULL,
    type_kind            expression_type_kind NOT NULL,

    -- For enum types: values stored as children with type_kind='enum'
    -- and enum_value set.
    enum_value           text DEFAULT NULL,

    -- For record types: field name stored on child rows.
    field_name           text DEFAULT NULL,

    -- For object types: class reference.
    object_class_key     text DEFAULT NULL,

    -- For sequence types: whether elements must be unique.
    element_unique       boolean DEFAULT NULL,

    PRIMARY KEY (model_key, expression_type_key),

    CONSTRAINT fk_expr_type_parent
        FOREIGN KEY (model_key, parent_type_key)
        REFERENCES expression_type (model_key, expression_type_key)
        ON DELETE CASCADE,

    CONSTRAINT fk_expr_type_model
        FOREIGN KEY (model_key)
        REFERENCES model (model_key)
        ON DELETE CASCADE,

    CONSTRAINT fk_expr_type_class
        FOREIGN KEY (model_key, object_class_key)
        REFERENCES class (model_key, class_key)
        ON DELETE CASCADE
);
```

This follows the same adjacency-list pattern as `expression_node`. Type trees are typically shallow (depth 2-3), so the pattern is efficient.

### Storage Examples

**`IntegerType`** — single row:
```
expression_type_key='attr_amount/type', type_kind='integer'
```

**`SetType{ElementType: StringType}`** — two rows:
```
expression_type_key='attr_tags/type', type_kind='set', parent=NULL
expression_type_key='attr_tags/type/0', type_kind='string', parent='attr_tags/type'
```

**`SequenceType{ElementType: IntegerType, Unique: true}`** — two rows:
```
expression_type_key='attr_ids/type', type_kind='sequence', element_unique=true, parent=NULL
expression_type_key='attr_ids/type/0', type_kind='integer', parent='attr_ids/type'
```

**`RecordType{Fields: [{name, StringType}, {age, IntegerType}]}`** — three rows:
```
expression_type_key='attr_person/type', type_kind='record', parent=NULL
expression_type_key='attr_person/type/0', type_kind='string', parent='attr_person/type', field_name='name', sort_order=0
expression_type_key='attr_person/type/1', type_kind='integer', parent='attr_person/type', field_name='age', sort_order=1
```

**`EnumType{Values: ["red", "green", "blue"]}`** — four rows (parent + one per value):
```
expression_type_key='attr_color/type', type_kind='enum', parent=NULL
expression_type_key='attr_color/type/0', type_kind='enum', parent='attr_color/type', enum_value='red', sort_order=0
expression_type_key='attr_color/type/1', type_kind='enum', parent='attr_color/type', enum_value='green', sort_order=1
expression_type_key='attr_color/type/2', type_kind='enum', parent='attr_color/type', enum_value='blue', sort_order=2
```

### Linking to DataType and Logic

ExpressionType attaches to the `data_type` table, not to `attribute` or `parameter` tables. Attributes and parameters already have FKs to `data_type` — the precise type flows through that existing relationship:

```sql
ALTER TABLE data_type
    ADD COLUMN expression_type_notation text DEFAULT NULL,       -- "tla_plus"
    ADD COLUMN expression_type_specification text DEFAULT NULL,   -- Human-authored TLA+ string
    ADD COLUMN expression_type_key text DEFAULT NULL,             -- FK to parsed type tree
    ADD CONSTRAINT fk_data_type_expression_type
        FOREIGN KEY (model_key, expression_type_key)
        REFERENCES expression_type (model_key, expression_type_key)
        ON DELETE CASCADE;
```

This means every DataType (whether owned by an attribute, parameter, or record field) can optionally carry a TypeSpec — the notation, specification string, and parsed type tree. No changes to `attribute`, `action_parameter`, `query_parameter`, or `event_parameter` tables are needed.

For Logic.TargetTypeSpec (the declared type of a logic expression's result):

```sql
ALTER TABLE logic
    ADD COLUMN target_type_notation text DEFAULT NULL,
    ADD COLUMN target_type_specification text DEFAULT NULL,
    ADD COLUMN target_type_key text DEFAULT NULL,
    ADD CONSTRAINT fk_logic_target_type
        FOREIGN KEY (model_key, target_type_key)
        REFERENCES expression_type (model_key, expression_type_key)
        ON DELETE CASCADE;
```

### Relationship Diagram

```
                    ┌──────────────┐
                    │    model     │
                    └──────┬───────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
   ┌──────┴───────┐ ┌─────┴──────┐  ┌──────┴──────────┐
   │  data_type   │ │   logic    │  │ expression_type  │
   │  (stakeholder│ │            │  │  (precise types) │
   │   types)     │ │            │  │                  │
   └──────┬───────┘ └─────┬──────┘  └──────┬───────────┘
          │               │                │
          │  expression   │                │
          │  _type_key ───┼────────────────┘
          │  (optional)   │         target_type_key
          │               │─────────────────┘
          │         ┌─────┴──────┐   (optional)
          │         │ expression │
          │         │   _node    │
          │         └────────────┘
          │
   ┌──────┴───────────────┐
   │      attribute        │
   │  data_type_key ──►    │
   │      data_type        │
   └───────────────────────┘
```

The `expression_type_key` on `data_type` is nullable. The three valid states:
- **Stakeholder only**: DataType exists, `expression_type_key` NULL — the normal starting state
- **Both**: DataType exists with `expression_type_key` set — fully specified
- **Precise only**: DataType exists (as structural carrier) with `expression_type_key` set but minimal business constraints

`Logic.target_type_key` is independently nullable — it declares the expected result type of the expression.

---

## Type Mapping Between Systems

### DataType CollectionType → ExpressionType

| DataType.CollectionType | CollectionUnique | ExpressionType | TLA+ Constructor | Notes |
|---|---|---|---|---|
| `atomic` + `unconstrained` | n/a | Any leaf type | varies | No constraint on precise type |
| `atomic` + `span` | n/a | `IntegerType` or `RationalType` | `Nat`/`Int`/`Real` | Integer if precision=1.0 and no denominators; rational otherwise |
| `atomic` + `enumeration` | n/a | `EnumType` | `{"a","b","c"}` | Precise enum values must be superset of stakeholder enum values |
| `atomic` + `reference` | n/a | Any type (determined by NamedSet) | varies | See note below |
| `atomic` + `object` | n/a | `ObjectType{ClassKey}` | n/a | Class keys must match |
| `ordered` | true | `SequenceType{Unique: true}` or `SetType` | `_SeqUnique(T)` or `_Set(T)` / `SUBSET T` | Author chooses based on code gen needs |
| `ordered` | false | `SequenceType{Unique: false}` | `Seq(T)` / `_Seq(T)` | Standard ordered collection |
| `unordered` | true | `SetType{ElementType}` | `_Set(T)` / `SUBSET T` | Unique unordered → set |
| `unordered` | false | `BagType{ElementType}` | `_Bag(T)` | Non-unique unordered → bag/multiset |
| `stack` | true | `SequenceType{Unique: true}` | `_StackUnique(T)` | LIFO with uniqueness |
| `stack` | false | `SequenceType{Unique: false}` | `_Stack(T)` | Standard LIFO stack |
| `queue` | true | `SequenceType{Unique: true}` | `_QueueUnique(T)` | FIFO with uniqueness |
| `queue` | false | `SequenceType{Unique: false}` | `_Queue(T)` | Standard FIFO queue |
| `record` | n/a | `RecordType{Fields}` | `[f1: T1, f2: T2]` | Field names must match; field types recursively mapped |

**Note on `reference` types:** The stakeholder sees a reference as a citation to external documentation (e.g., `"ref from ISO 3166-2 US state abbreviation codes"`). But the precise type depends on what the referenced data actually looks like. A set of state codes might be a flat set of strings (`SetType{StringType}`), but it could equally be a set of records (`SetType{RecordType{[code: StringType, name: StringType, region: StringType]}}`). The ExpressionType is determined by the NamedSet definition that formalizes the reference — the reference DataType only tells us "there's external data here," not what shape it takes.

### Builtin Calls → ExpressionType Signatures

The precise type system enables type signatures for builtin calls:

| Builtin | Signature |
|---|---|
| `_Seq!Head` | `SequenceType(T) → T` |
| `_Seq!Tail` | `SequenceType(T) → SequenceType(T)` |
| `_Seq!Append` | `(SequenceType(T), T) → SequenceType(T)` |
| `_Seq!Len` | `SequenceType(T) → IntegerType` |
| `_Stack!Push` | `(SequenceType(T), T) → SequenceType(T)` |
| `_Stack!Pop` | `SequenceType(T) → T` |
| `_Queue!Enqueue` | `(SequenceType(T), T) → SequenceType(T)` |
| `_Queue!Dequeue` | `SequenceType(T) → T` |
| `_FiniteSet!Cardinality` | `SetType(T) → IntegerType` |
| `_Bags!SetToBag` | `SetType(T) → BagType(T)` |
| `_Bags!BagToSet` | `BagType(T) → SetType(T)` |
| `_Bags!CopiesIn` | `(T, BagType(T)) → IntegerType` |

These signatures connect the expression layer's builtin calls directly to the precise type system.

---

## Integration with Expression Validation

### Type-Aware Expression Validation

Currently, `expression.Validate()` checks structural correctness (required fields present, children non-nil). With ExpressionType, a second validation pass can check type correctness:

```go
func TypeCheck(
    expr model_expression.Expression,
    context TypeContext,
) (model_expression_type.ExpressionType, []TypeError)
```

Where `TypeContext` provides:
- Attribute types: `AttributeKey → ExpressionType`
- Parameter types: `ParameterName → ExpressionType`
- Global function signatures: `FunctionKey → FunctionType`
- Local variable bindings: `VarName → ExpressionType` (from quantifiers)

This replaces the current Hindley-Milner type checker with a simpler, deterministic type checker that uses declared types rather than inferring them. The current type checker remains useful as a fallback when precise types are not yet declared (incremental adoption).

### Inference Direction

When a precise type exists for an attribute, the type checker uses it directly. When no precise type exists, the type checker can infer the type from the expression and offer to store it as the precise type. This enables a workflow:

1. Author writes stakeholder DataType (`"ordered of unconstrained"`)
2. Author writes Logic expression (`_Seq!Append(self.items, param.new_item)`)
3. Type checker infers: items is `SequenceType(T)`, new_item is `T`
4. If author provides new_item's type as `StringType`, then items resolves to `SequenceType(StringType)`
5. System stores `SequenceType(StringType)` as the precise type for the `items` attribute
6. Compatibility check confirms: `ordered` ↔ `SequenceType` is valid

---

## How Expression Types Get Authored

A precise ExpressionType doesn't appear out of thin air. Someone — either a human author or an AI parser — must write the formal definition that resolves to that type. Today, formal definitions live in `Logic` objects, which have a `Notation` (currently always TLA+), a `Specification` string, and an optional `Expression` tree. The question is: where does the formal definition of a *type* live?

### The Authoring Pattern

Currently, when a human wants to express something precise in the model, they write TLA+:

- **Guard logic**: `self.amount > 0` (TLA+ expression → Logic.Specification → Logic.Expression)
- **Action guarantee**: `self.amount + param.delta` (TLA+ expression → Logic.Specification → Logic.Expression)

The authoring flow is always: **human writes TLA+ string → parser produces structured tree → tree stored in model**.

DataType should follow the same pattern. Just as Logic has `Specification` (human-written TLA+) and `Expression` (parsed tree), DataType needs `ExpressionTypeSpecification` (human-written TLA+ type expression) and `ExpressionType` (parsed type tree).

### The Reusable Trio: Notation + Specification + Structured Tree

The pattern of `Notation` + `Specification` + parsed structured tree appears in multiple places:
- **Logic**: `Notation` + `Specification` + `Expression`
- **DataType**: needs `Notation` + `Specification` + `ExpressionType` (for precise types)
- **NamedSet**: needs `Notation` + `Specification` + `Expression` (for set definitions)

These three fields always travel together — you never have a specification without a notation, and the parsed tree is always derived from the specification string. This is a reusable value object:

```go
// FormalSpec carries a formal specification in a given notation,
// along with its parsed structured representation.
type FormalSpec[T any] struct {
    Notation      string  `validate:"required,oneof=tla_plus"` // The notation language.
    Specification string  // The human-authored source text.
    Parsed        T       // The parsed structured tree (nil/zero = not yet parsed).
}
```

In practice (since Go generics with interfaces need care), this could also be two concrete types:

```go
// ExpressionSpec carries a TLA+ specification and its parsed Expression tree.
type ExpressionSpec struct {
    Notation      string                        `validate:"required,oneof=tla_plus"`
    Specification string
    Expression    model_expression.Expression   // Parsed expression tree.
}

// TypeSpec carries a TLA+ type specification and its parsed ExpressionType tree.
type TypeSpec struct {
    Notation      string                                    `validate:"required,oneof=tla_plus"`
    Specification string
    ExpressionType model_expression_type.ExpressionType     // Parsed type tree.
}
```

Then the model types embed these:

```go
type Logic struct {
    Key           identity.Key
    Type          string
    Description   string
    Target        string
    Spec          ExpressionSpec   // Notation + Specification + Expression
}

type DataType struct {
    CollectionType   string
    CollectionUnique *bool
    CollectionMin    *int
    CollectionMax    *int
    Atomic           *Atomic
    RecordFields     []RecordFieldDataType
    TypeSpec         *TypeSpec    // Optional: Notation + Specification + ExpressionType
}
```

Whether to extract this into its own package or keep it as embedded structs is an implementation detail. The key insight is that the trio is a single concept — a formal specification — and should be grouped as one, not spread across three independent fields. This also makes it impossible to have a specification without a notation, or a parsed tree orphaned from its source text.

### DataType With TypeSpec

```go
type DataType struct {
    CollectionType   string
    CollectionUnique *bool
    CollectionMin    *int
    CollectionMax    *int
    Atomic           *Atomic
    RecordFields     []RecordFieldDataType
    TypeSpec         *TypeSpec    // Optional precise type (nil = stakeholder type only)
}
```

When `TypeSpec` is nil, the DataType has no precise type — stakeholder only. When present, it carries the notation, the human-authored TLA+ string, and the parsed ExpressionType tree. The compatibility checker validates the ExpressionType against the DataType's stakeholder fields.

### What the Human Writes

In the input file, the attribute's `rules` field already carries the stakeholder type. A new field carries the precise type as TLA+:

```yaml
attributes:
  items:
    name: Items
    details: The line items in this order.
    rules: ordered of unconstrained
    precise_type: Seq([name: STRING, price: Nat])
```

Or for simpler attributes:

```yaml
attributes:
  balance:
    name: Account Balance
    rules: [0 .. 1000000] at 0.01 dollar
    precise_type: Nat
  tags:
    name: Tags
    rules: unique unordered of unconstrained
    precise_type: _Set(STRING)
  status:
    name: Status
    rules: enum of open, closed, pending
    precise_type: {"open", "closed", "pending"}
```

The parser reads `precise_type` as a TLA+ string and stores it in `DataType.ExpressionTypeSpecification`. A TLA+ type parser converts the string into the ExpressionType tree. Both the raw string and the parsed tree are stored — the string for round-tripping, the tree for machine consumption.

Parameters work the same way — they already have DataTypes, and the precise type specification flows through the same DataType fields.

### TLA+ Already Has Type Notation

TLA+ expresses types as sets — there is no separate type language. This means the human writes the same kind of expressions they already know:

| Precise Type | TLA+ Notation | What It Means |
|---|---|---|
| `BooleanType` | `BOOLEAN` | The set {TRUE, FALSE} |
| `IntegerType` | `Int` | The set of all integers |
| `IntegerType` (natural) | `Nat` | The set {0, 1, 2, ...} |
| `RationalType` | `Real` | The set of all reals |
| `StringType` | `STRING` | The set of all strings |
| `EnumType{"red","green","blue"}` | `{"red", "green", "blue"}` | A finite set literal |
| `SetType(StringType)` | `SUBSET STRING` | The powerset of strings |
| `SequenceType(IntegerType)` | `Seq(Int)` | All finite sequences of integers |
| `RecordType({name: String, age: Int})` | `[name: STRING, age: Int]` | Record type constructor |
| `SetType(RecordType(...))` | `SUBSET [name: STRING, age: Int]` | Sets of records |

### Why DataType Is the Right Home

Previous analysis considered five options (Logic reuse, bare fields on Attribute, TypeDefinition wrapper, Logic+TypeExpr hybrid, fields on DataType). The decision that ExpressionType lives on DataType resolves this — the authoring fields belong on DataType too:

1. **Co-location.** The stakeholder type and precise type describe the same thing from different perspectives. Having both on DataType makes the compatibility check local — no cross-referencing between different model entities.

2. **Follows the existing pattern.** DataType already has `DataTypeRules` as its human-authored string that gets parsed into the structured fields. Adding `ExpressionTypeSpecification` as the TLA+ equivalent is the same pattern: human string → parsed structure.

3. **Works for nested types.** DataType record fields contain nested DataTypes. If a record field needs a precise type, it's right there on its own DataType — no need for external linkage.

4. **No new model entities.** No TypeDefinition struct, no new database tables for type definitions. The expression_type table stores the parsed trees, and the DataType table stores the specification strings.

5. **Parameters get it for free.** Parameters already have DataTypes. Adding the precise type specification to DataType means parameters automatically gain the authoring surface without any Parameter-specific changes.

### Types as TLA+ Sets of Valid Values

This design rests on a fundamental TLA+ principle: **every type is a set of valid values**. There is no separate "type language" in TLA+ — you express types by writing the set that contains all valid values of that type. This is why the human writes TLA+ set expressions for precise types.

The built-in TLA+ sets that serve as base types:

| TLA+ Set Expression | What It Represents | ExpressionType |
|---|---|---|
| `BOOLEAN` | `{TRUE, FALSE}` | `BooleanType` |
| `Nat` | `{0, 1, 2, ...}` | `IntegerType` (non-negative) |
| `Int` | `{..., -2, -1, 0, 1, 2, ...}` | `IntegerType` |
| `Real` | All real numbers | `RationalType` |
| `STRING` | All finite strings | `StringType` |
| `{"a", "b", "c"}` | Finite set literal | `EnumType{Values: ["a","b","c"]}` |
| `SUBSET S` | Powerset of S (all subsets of S) | `SetType{ElementType: S}` |
| `Seq(S)` | All finite sequences over S | `SequenceType{ElementType: S}` |
| `[f1: T1, f2: T2]` | Record type constructor | `RecordType{Fields: [{f1,T1},{f2,T2}]}` |
| `S1 \X S2` | Cartesian product | `TupleType{ElementTypes: [S1, S2]}` |

This system extends TLA+ with custom set constructors for collection types. Each constructor has a standard and a `Unique` variant. The `Unique` variants are essential because without them, the human would need to write awkward quantified predicates to express uniqueness — e.g., `{s \in Seq(T) : \A i, j \in DOMAIN s : i /= j => s[i] /= s[j]}` for a sequence with unique elements. The custom constructors make the intent clear in one word and give code generators an explicit uniqueness signal.

**Standard constructors** (allow duplicate elements):

| Custom Set Expression | What It Represents | ExpressionType |
|---|---|---|
| `_Set(S)` | All finite subsets of S | `SetType{ElementType: S}` |
| `_Seq(S)` | All finite sequences over S | `SequenceType{ElementType: S, Unique: false}` |
| `_Stack(S)` | Sequences over S with LIFO access | `SequenceType{ElementType: S, Unique: false}` |
| `_Queue(S)` | Sequences over S with FIFO access | `SequenceType{ElementType: S, Unique: false}` |
| `_Bag(S)` | All finite bags/multisets over S | `BagType{ElementType: S}` |

**Unique constructors** (all elements must be distinct):

| Custom Set Expression | What It Represents | ExpressionType |
|---|---|---|
| `_SeqUnique(S)` | Sequences over S with no duplicates | `SequenceType{ElementType: S, Unique: true}` |
| `_StackUnique(S)` | Unique-element stacks over S | `SequenceType{ElementType: S, Unique: true}` |
| `_QueueUnique(S)` | Unique-element queues over S | `SequenceType{ElementType: S, Unique: true}` |

Note: `_Set(S)` is a helper alias for the standard TLA+ `SUBSET S`. Both are accepted and produce the same `SetType`. Sets are inherently unique — there is no non-unique set variant (that's what `_Bag(S)` is for).

The `_Set(S)`, `_Seq(S)` constructors are helper aliases for the standard TLA+ `SUBSET S` and `Seq(S)` respectively. Both canonical and helper forms are accepted and produce the same TypeExpression. The helper forms exist for consistency across the full constructor family — `_Set`, `_Seq`, `_Stack`, `_Queue`, `_Bag` — so that authors can use a uniform syntax. The underscore prefix signals "this is a custom system constructor." The standard TLA+ forms (`SUBSET S`, `Seq(S)`) remain valid for authors who prefer canonical TLA+.

The `_Stack` and `_Queue` constructors resolve to `SequenceType` — they exist as documentation hints for the author (and for builtin operation validation), not as distinct structural types. The `_Bag(S)` constructor is structurally distinct because bags have different semantics than sets (duplicate elements with counts).

The `Unique` flag on `SequenceType` is a property that flows to code generators. A `SequenceType{Unique: true}` tells the code generator to enforce uniqueness on insertion operations. This is the precise-type equivalent of the stakeholder's `CollectionUnique` flag on DataType — both express the same constraint, but the precise type embeds it in the structural type itself rather than as a separate flag.

This means the full TLA+ type vocabulary for a human author is:

```
# Primitives
precise_type: BOOLEAN
precise_type: Nat
precise_type: Int
precise_type: Real
precise_type: STRING
precise_type: {"open", "closed", "pending"}

# Sets (inherently unique)
precise_type: SUBSET STRING                    # set of strings (standard TLA+)
precise_type: _Set(STRING)                     # same as SUBSET STRING, helper alias
precise_type: _Set([name: STRING, age: Nat])   # set of records

# Sequences (allow duplicates)
precise_type: Seq(Nat)                         # sequence of integers (standard TLA+)
precise_type: _Seq(Nat)                        # same as Seq(Nat), helper alias
precise_type: _Seq([name: STRING, price: Nat]) # sequence of records

# Sequences (unique elements)
precise_type: _SeqUnique(Nat)                  # sequence of distinct integers
precise_type: _SeqUnique(STRING)               # sequence of distinct strings

# Stacks and Queues (allow duplicates)
precise_type: _Stack(STRING)                   # LIFO stack of strings
precise_type: _Queue(Nat)                      # FIFO queue of integers

# Stacks and Queues (unique elements)
precise_type: _StackUnique(STRING)             # LIFO stack, no duplicate strings
precise_type: _QueueUnique(Nat)                # FIFO queue, no duplicate integers

# Bags/Multisets
precise_type: _Bag(STRING)                     # bag/multiset of strings

# Records
precise_type: [name: STRING, age: Nat, active: BOOLEAN]

# Tuples (fixed-length heterogeneous)
precise_type: Nat \X STRING                    # (integer, string) pair
precise_type: Nat \X STRING \X BOOLEAN         # triple

# Nested
precise_type: SUBSET _Seq([item: STRING, qty: Nat])  # set of sequences of records
precise_type: _Bag([name: STRING, score: Nat])        # bag of records
```

All of these are valid TLA+ expressions (with the custom constructors registered as builtin calls following the `_Module(args)` pattern already used for `_Seq`, `_Stack`, `_Queue`, `_Bags`). The parsed AST is then converted to a TypeExpression tree by a straightforward mapping.

### TypeSpec — The Formal Specification Object

As described in "The Reusable Trio" above, the precise type authoring fields are grouped into a `TypeSpec` value object that lives on DataType. This replaces the earlier TypeDefinition concept — there is no separate entity. The TypeSpec is a property of DataType, not a standalone model object with its own key.

### TLA+ AST to TypeExpression Conversion

The conversion from TLA+ AST nodes to TypeExpression nodes is a straightforward mapping:

| TLA+ Source | AST Node | TypeExpression |
|---|---|---|
| `BOOLEAN` | `SetConstant("BOOLEAN")` | `BooleanType{}` |
| `Nat`, `Int` | `SetConstant("Nat")` / `SetConstant("Int")` | `IntegerType{}` |
| `Real` | `SetConstant("Real")` | `RationalType{}` |
| `STRING` | `BuiltinCall("STRING")` or identifier | `StringType{}` |
| `{"a","b","c"}` | `SetLiteral{Elements: [StringLiteral...]}` | `EnumType{Values: [...]}` |
| `SUBSET X` | `Powerset{Set: X}` | `SetType{ElementType: convert(X)}` |
| `_Set(X)` | `BuiltinCall("_Set", "_Set", [X])` | `SetType{ElementType: convert(X)}` |
| `Seq(X)` | `BuiltinCall("_Seq", "Seq", [X])` | `SequenceType{ElementType: convert(X), Unique: false}` |
| `_Seq(X)` | `BuiltinCall("_Seq", "_Seq", [X])` | `SequenceType{ElementType: convert(X), Unique: false}` |
| `_SeqUnique(X)` | `BuiltinCall("_Seq", "_SeqUnique", [X])` | `SequenceType{ElementType: convert(X), Unique: true}` |
| `_Stack(X)` | `BuiltinCall("_Stack", "_Stack", [X])` | `SequenceType{ElementType: convert(X), Unique: false}` |
| `_StackUnique(X)` | `BuiltinCall("_Stack", "_StackUnique", [X])` | `SequenceType{ElementType: convert(X), Unique: true}` |
| `_Queue(X)` | `BuiltinCall("_Queue", "_Queue", [X])` | `SequenceType{ElementType: convert(X), Unique: false}` |
| `_QueueUnique(X)` | `BuiltinCall("_Queue", "_QueueUnique", [X])` | `SequenceType{ElementType: convert(X), Unique: true}` |
| `_Bag(X)` | `BuiltinCall("_Bags", "_Bag", [X])` | `BagType{ElementType: convert(X)}` |
| `[f1: T1, f2: T2]` | `RecordConstructor{Fields: [...]}` | `RecordType{Fields: [...]}` |
| `T1 \X T2` | `CartesianProduct{Operands: [...]}` | `TupleType{ElementTypes: [...]}` |

AST nodes that don't correspond to types (BinaryArith, IfThenElse, Quantifier, etc.) produce a conversion error — the human wrote a computation where a type was expected.

Note: The access-pattern distinction (Stack vs Queue vs Seq) is lost in the TypeExpression — all three produce `SequenceType`. The access pattern is a stakeholder concern captured by the DataType's `CollectionType` field, not a structural concern. The `Unique` flag, however, is preserved because it is a structural property that affects code generation (insertion operations must enforce uniqueness).

---

## ExpressionType Is Structure, DataType Has Constraints — Both Are Needed

A code generator or simulator needs both ExpressionType and DataType. They carry different information:

| Information | Where It Lives | Example |
|---|---|---|
| What data structure to use | ExpressionType | `SequenceType{ElementType: IntegerType, Unique: true}` |
| Minimum cardinality | DataType.CollectionMin | `3` (at least 3 elements) |
| Maximum cardinality | DataType.CollectionMax | `10` (at most 10 elements) |
| Numeric range bounds | DataType.Atomic.Span | `[0 .. 1000] at 0.01 dollar` |
| Allowed enum values | DataType.Atomic.Enums | `["open", "closed", "pending"]` |
| Units of measure | DataType.Atomic.Span.Units | `"dollar"` |
| Precision | DataType.Atomic.Span.Precision | `0.01` |

ExpressionType does NOT carry constraints like min/max counts, span bounds, units, or precision. These are business rules that belong on DataType. ExpressionType answers "what shape is this value?" while DataType constraints answer "what values are valid?"

### Why Not Put Constraints on ExpressionType?

Constraints are business rules, not structural types. Consider:

- Two attributes might have the same ExpressionType (`SequenceType{IntegerType}`) but different cardinality constraints (one has min=1, the other has min=3, max=10). If constraints were on the ExpressionType, these would be different types — which breaks type compatibility. You couldn't pass one where the other is expected.

- Constraints change independently of structure. A stakeholder might tighten the min from 1 to 3 without changing the structural type. If constraints were on ExpressionType, this would require updating every expression that references the type.

- Constraints are inherently approximate — they describe the stakeholder's understanding of valid ranges. The precise type is exact — it describes what the machine needs to generate. Mixing approximate constraints into the exact type muddies both.

### How Consumers Use Both

**Simulator:**
- Uses ExpressionType for type-checking expressions (does `self.items` have the right type for `_Seq!Append`?)
- Derives invariant rules from DataType constraints:
  - `Len(self.items) >= 3` from `CollectionMin = 3`
  - `Len(self.items) <= 10` from `CollectionMax = 10`
  - `self.amount >= 0 /\ self.amount <= 1000` from span bounds
- These derived invariants are checked after every state transition, using the same invariant-checking infrastructure that already exists

**Code Generator:**
- Uses ExpressionType to choose data structures (`SequenceType{Unique: true}` → use a linked hash set or ordered set)
- Uses DataType constraints to generate validation logic:
  - `if len(items) < 3 { return ValidationError("items must have at least 3 elements") }`
  - `if amount < 0 || amount > 1000 { return ValidationError("amount out of range") }`
- Uses DataType.CollectionType to determine API shape:
  - `stack` → expose `Push()`/`Pop()` methods only
  - `queue` → expose `Enqueue()`/`Dequeue()` methods only
  - `ordered` → expose full sequence API

**AI Parser:**
- Produces DataType with both stakeholder constraints and optional TypeSpec (precise TLA+ spec)
- Can validate compatibility between stakeholder constraints and ExpressionType

**Human Parser:**
- Reads `rules:` field → DataType stakeholder constraints
- Reads `precise_type:` field → DataType.TypeSpec
- Both are optional and independent — TypeSpec is nil when `precise_type` is absent

### The Full Picture for a Single Attribute

```
Attribute: items
└── DataType
    ├── CollectionType: "ordered"
    ├── CollectionUnique: false
    ├── CollectionMin: 1
    ├── CollectionMax: 50
    ├── Atomic: {ConstraintType: "unconstrained"}
    │
    ├── TypeSpec (precise view, optional)
    │   ├── Notation: "tla_plus"
    │   ├── Specification: "_Seq([name: STRING, price: Nat])"
    │   └── ExpressionType: SequenceType{
    │         ElementType: RecordType{Fields: [
    │           {Name: "name", Type: StringType{}},
    │           {Name: "price", Type: IntegerType{}}
    │         ]},
    │         Unique: false
    │       }
    │
    └── Derived Invariants (from stakeholder constraints)
        ├── Len(self.items) >= 1    (from CollectionMin)
        └── Len(self.items) <= 50   (from CollectionMax)
```

The TypeSpec within DataType is optional. When present, compatibility is checked (ordered ↔ SequenceType, field names match if both specify records, etc.). The derived invariants come from the stakeholder constraint fields — they don't depend on TypeSpec existing.

---

## Reference Types and Named Set Definitions

### The Problem with Reference Atomic Types

The `reference` constraint type in DataType is special. Unlike other atomic types, it describes a human-readable business rule — a citation to external documentation. For example:

```yaml
attributes:
  state_code:
    name: State Code
    rules: ref from ISO 3166-2 US state abbreviation codes
```

This produces `Atomic{ConstraintType: "reference", Reference: "ISO 3166-2 US state abbreviation codes"}`. The stakeholder sees a clear description of what values are valid. But the TLA+ model needs something precise — a finite set of allowed values that the simulator can check membership against and the code generator can use for validation.

The previous design intent was that global functions would serve double duty: a global function like `_IsoStateAbbr` would define the set of valid values, and attribute type expressions would reference it. But semantically, global functions are computations — they take parameters and produce values through logic. A set definition is not a function — it's a named constant (a set literal or a defined set expression). Reusing `GlobalFunction` for this purpose is a category error.

### What's Needed

A named set definition needs:
1. **A name** — usable in TLA+ expressions (e.g., `IsoStateAbbr`)
2. **A TLA+ specification** — the formal definition of the set (e.g., `{"AL", "AK", "AZ", ...}` or a more elaborate construction)
3. **A description** — human-readable explanation of what the set represents
4. **Reusability** — multiple attributes and parameters can reference the same set definition
5. **Model-level scope** — like global functions, these are root-level entities not tied to any class

The ExpressionType for an attribute using a named set would be something like `NamedSetRef{SetKey: ...}` or would resolve to the actual type of the set's elements.

### Approaches

#### Approach 1: Named Set Definitions — New Model Entity

Add a new root-level model entity specifically for named set definitions:

```go
// NamedSet defines a reusable set of valid values that can be referenced
// in type expressions throughout the model.
type NamedSet struct {
    Key           identity.Key      // KEY_TYPE_NAMED_SET, root-level
    Name          string            // "IsoStateAbbr" — used in TLA+ text
    Description   string            // "ISO 3166-2 US state abbreviation codes"
    Spec          ExpressionSpec    // Notation + Specification + Expression (the reusable trio)
    TypeSpec      *TypeSpec         // Optional: what type the elements are (e.g., SetType{StringType})
}
```

Note that NamedSet uses both trios: `ExpressionSpec` for the set definition itself (a TLA+ expression that evaluates to a set), and optionally `TypeSpec` for declaring the type of the set (e.g., `SetType{RecordType{...}}`). The ExpressionSpec holds the computational definition, the TypeSpec holds the structural type declaration.

The model gains a new top-level field:

```go
type Model struct {
    // ...existing...
    GlobalFunctions map[identity.Key]model_logic.GlobalFunction
    NamedSets       map[identity.Key]NamedSet            // NEW
}
```

In TLA+ type expressions, named sets are referenced by name:

```yaml
attributes:
  state_code:
    rules: ref from ISO 3166-2 US state abbreviation codes
    precise_type: IsoStateAbbr     # references the named set
  billing_state:
    rules: ref from ISO 3166-2 US state abbreviation codes
    precise_type: IsoStateAbbr     # same set, different attribute
```

In the ExpressionType tree, a reference to a named set could be:

```go
// NamedSetRef references a model-level named set definition.
type NamedSetRef struct {
    SetKey identity.Key   // FK to NamedSet
}
```

Or alternatively, the named set reference resolves to the set's element type during type resolution, and `NamedSetRef` is only an intermediate form.

**In the simulator:** The named set is a constant — its Expression tree is evaluated once and the resulting value (a TLA+ set) is available in scope. Any expression like `state_code \in IsoStateAbbr` resolves to membership in that set. This is semantically identical to how TLA+ CONSTANT declarations work.

**In behavioral logic:** Named sets can appear directly in Logic specifications:

```
\* Guard: state must be a valid ISO state
param.state \in IsoStateAbbr
```

The expression tree for this guard would contain a `NamedSetRef` node instead of a `GlobalCall` node.

**Database:**

```sql
CREATE TYPE named_set_key_type AS ENUM ('named_set');
COMMENT ON TYPE named_set_key_type IS 'Key type for named set definitions.';

CREATE TABLE named_set (
    model_key     text NOT NULL,
    set_key       text NOT NULL,
    name          text NOT NULL,
    description   text NOT NULL DEFAULT '',
    notation      text NOT NULL DEFAULT 'tla_plus',
    specification text NOT NULL DEFAULT '',
    PRIMARY KEY (model_key, set_key),
    UNIQUE (model_key, name),
    CONSTRAINT fk_named_set_model FOREIGN KEY (model_key) REFERENCES model (key) ON DELETE CASCADE
);
COMMENT ON TABLE named_set IS 'A named set definition that can be referenced in type expressions and logic.';
COMMENT ON COLUMN named_set.model_key IS 'The model this named set belongs to.';
COMMENT ON COLUMN named_set.set_key IS 'The unique key for this named set.';
COMMENT ON COLUMN named_set.name IS 'The name usable in TLA+ expressions (e.g., IsoStateAbbr).';
COMMENT ON COLUMN named_set.description IS 'Human-readable description of what this set represents.';
COMMENT ON COLUMN named_set.notation IS 'The notation used for the specification (tla_plus).';
COMMENT ON COLUMN named_set.specification IS 'The TLA+ definition of the set.';
```

The expression_node table and expression_type table would gain a `named_set_key` column for nodes/types that reference a named set.

**Pros:**
- Clear semantic distinction: named sets are constants, global functions are computations
- First-class model entity with its own identity key, validation, and database table
- Referenced by name in TLA+ — natural TLA+ usage (`x \in IsoStateAbbr`)
- Reusable across unlimited attributes and parameters
- The name convention doesn't require underscore prefix (not a function), making it clearer that it's a set, not a function
- Carries its own Expression for the simulator to evaluate once
- Carries its own TypeExpr for type-checking to know the element type

**Cons:**
- New model entity: new key type, new constructor, new validation, new database table, new parser support, new test helper entries — full build-order pass
- One more thing for the AI parser to learn and generate
- Adds a new node type to both Expression trees (`NamedSetRef` in model_expression) and TypeExpression trees (`NamedSetRef` in model_expression_type)

#### Approach 2: Extend GlobalFunction with a `Kind` Field

Keep global functions but distinguish between computational functions and set definitions:

```go
const (
    GlobalFunctionKindFunction = "function"  // A computation: takes args, returns value
    GlobalFunctionKindSetDef   = "set_def"   // A named set definition: no args, value is a set
)

type GlobalFunction struct {
    Key        identity.Key
    Name       string
    Kind       string             // NEW: "function" or "set_def"
    Parameters []string           // Empty for set_def
    Logic      Logic              // LogicTypeValue for both
}
```

Usage in TLA+ is the same — `IsoStateAbbr` or `_IsoStateAbbr` appears in expressions. The distinction is purely internal for validation and tooling.

**Pros:**
- Minimal change — extends existing infrastructure rather than creating new
- Same database table, same parser pipeline, same expression node type (GlobalCall)
- Reuses existing test helpers, validators, and lifecycle

**Cons:**
- Semantically muddy — "global function" now means two different things
- The underscore naming convention (`_Max`, `_Identity`) was designed for functions. Set definitions might not want the underscore prefix
- `Parameters` must be empty for set_def but the field still exists
- `Logic.Type` must be `LogicTypeValue` for both, but a set definition's "value" is specifically a set — this isn't enforced
- Downstream consumers (simulator, code gen) must check `Kind` to know how to treat the entity
- Conflates two concepts that have different lifecycles — function definitions change when computation logic changes; set definitions change when business rules change

#### Approach 3: Named Sets as Model-Level Constants (CONSTANT in TLA+ Terminology)

TLA+ has a `CONSTANT` declaration mechanism — named values that are fixed for a given model instantiation. Named sets fit this pattern precisely. This approach introduces a `ModelConstant` entity:

```go
// ModelConstant represents a named constant value at the model level.
// In TLA+ terms, this is a CONSTANT whose value is provided by the specification.
type ModelConstant struct {
    Key           identity.Key                           // KEY_TYPE_MODEL_CONSTANT
    Name          string                                 // "IsoStateAbbr"
    Description   string                                 // Human-readable
    Notation      string                                 // "tla_plus"
    Specification string                                 // TLA+ definition
    Expression    model_expression.Expression             // Parsed expression tree
    TypeExpr      model_expression_type.ExpressionType   // Resolved type
}
```

This is more general than Approach 1 — it allows named constants of any type, not just sets. A named integer constant, a named record constant, etc. are all possible.

**Pros:**
- Aligns with TLA+ semantics — CONSTANT is a first-class TLA+ concept
- More general than named sets — supports any constant value (though sets are the primary use case)
- Clean semantics: a constant is neither a function nor a set definition — it's a named value
- In TLA+, referencing a CONSTANT by name is the standard pattern

**Cons:**
- Same cost as Approach 1 (new entity, full build-order pass)
- Generality may be premature — the immediate need is specifically for sets of allowed values
- "ModelConstant" is less self-documenting than "NamedSet" for the primary use case
- Naming: TLA+ constants are typically ALL_CAPS by convention, but the model uses CamelCase elsewhere

#### Approach 4: TypeDefinition-Level Named Sets (No Model Entity)

Instead of a model-level entity, put the set definition directly on the TypeDefinition where it's first used, and allow TypeDefinitions to reference each other:

```go
type TypeDefinition struct {
    Key           identity.Key
    ExportedName  string                              // Optional: if set, this type is available by name
    Description   string
    Notation      string
    Specification string
    TypeExpr      model_expression_type.ExpressionType
}
```

The first attribute that defines `IsoStateAbbr` exports it by name. Subsequent attributes reference it. The model collects all exported names as an implicit global namespace.

**Pros:**
- No new model entity — reuses the TypeDefinition struct from the recommended Option C
- The definition lives close to its first use
- Lighter-weight than a full model-level entity

**Cons:**
- "First use exports" is fragile — which attribute owns the definition? What if it's deleted?
- The exported name is a side effect of one attribute's type definition — confusing for the reader
- Circular references are possible (TypeDefinition A references B which references A)
- Doesn't work for behavioral logic — if a guard references `IsoStateAbbr`, where is the set defined? It's on some attribute's TypeDefinition, which is awkward to discover
- The set definition is not clearly a model-level entity, making it hard to manage

### Recommendation: Approach 1 (Named Set Definitions) or Approach 3 (Model Constants)

Approaches 2 and 4 are unsatisfying — one muddies the global function concept, the other creates fragile implicit exports.

The choice between Approach 1 and 3 comes down to whether we need named constants beyond sets:

**If only sets are needed** → Approach 1 (`NamedSet`) is clearer and more self-documenting. The entity name tells you exactly what it is. The validation can enforce that the Expression must evaluate to a set.

**If other constants may be needed** → Approach 3 (`ModelConstant`) is more general. But generality means weaker validation — can't enforce "must be a set" at the type level.

For the immediate need (reference types pointing to shared set definitions), Approach 1 is the most direct fit. The entity is called `NamedSet`, the TLA+ name is used directly in type expressions and behavioral logic, and the model enforces that the specification defines a set.

### How Named Sets Connect to ExpressionType

With Approach 1, the ExpressionType hierarchy gains a new leaf type:

```go
// NamedSetRef references a model-level named set definition.
// In a type context, this represents "the type of elements in the named set."
// For example, if IsoStateAbbr = {"AL", "AK", ...}, then NamedSetRef{SetKey: isoStateAbbrKey}
// resolves to EnumType{Values: ["AL", "AK", ...]} during type resolution.
type NamedSetRef struct {
    SetKey identity.Key  // FK to NamedSet
}
```

In practice, consumers would resolve the reference:

```go
// During type checking, resolve NamedSetRef to the actual type:
func ResolveType(t ExpressionType, namedSets map[identity.Key]NamedSet) ExpressionType {
    if ref, ok := t.(*NamedSetRef); ok {
        return namedSets[ref.SetKey].TypeExpr
    }
    return t
}
```

The model_expression package also gains a new node type:

```go
// NamedSetRef references a model-level named set by key.
// Used in behavioral logic: "param.state \in IsoStateAbbr"
type NamedSetRef struct {
    SetKey identity.Key  // FK to NamedSet
}
```

This is different from `GlobalCall` — it doesn't take arguments, it's not a function call, it's a reference to a constant set.

### Compatibility Between Reference DataType and NamedSet

When an attribute has both a reference DataType and a TypeDefinition that uses a NamedSet, the compatibility check verifies:

1. The DataType is `reference` type → compatible with any NamedSet (the reference is a business description, the NamedSet is the formal definition)
2. The NamedSet's TypeExpr tells the type checker what type of values the attribute holds (e.g., `StringType` for state abbreviation codes, `RecordType` for more complex references)

### The Full Picture for a Reference Attribute

```
Attribute: state_code
├── DataType (stakeholder view)
│   ├── CollectionType: "atomic"
│   └── Atomic: {ConstraintType: "reference", Reference: "ISO 3166-2 US state abbreviation codes"}
│
├── TypeDefinition (precise view)
│   ├── Description: "Two-letter US state codes per ISO 3166-2."
│   ├── Notation: "tla_plus"
│   ├── Specification: "IsoStateAbbr"
│   └── TypeExpr: NamedSetRef{SetKey: key("isostateabbr")}
│
└── Model-Level NamedSet: IsoStateAbbr
    ├── Name: "IsoStateAbbr"
    ├── Description: "The set of valid US state abbreviation codes per ISO 3166-2."
    ├── Specification: '{"AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", ...}'
    ├── Expression: SetLiteral{Elements: [StringLiteral("AL"), StringLiteral("AK"), ...]}
    └── TypeExpr: EnumType{Values: ["AL", "AK", "AZ", ...]}
```

Multiple attributes/parameters can reference the same `IsoStateAbbr` named set. The set definition lives in one place, and changes propagate to all references.

### Named Sets in Behavioral Logic

Named sets are usable in all Logic types, not just TypeDefinitions:

```
\* Guard: validate state is in the allowed set
param.state \in IsoStateAbbr

\* Action guarantee: assign a valid state
self.billing_state' = param.state    \* type checker knows both are IsoStateAbbr elements

\* Safety rule: state must always be valid
self.state \in IsoStateAbbr

\* Query: filter by valid states
{order \in self.orders : order.state \in IsoStateAbbr}
```

The expression parser recognizes `IsoStateAbbr` as a named set reference (rather than a local variable or global function call) by looking it up in the model's named sets registry.

### Named Sets in the Human Parser

In the input file, named sets would appear at the model level, similar to global functions:

```yaml
named_sets:
  IsoStateAbbr:
    description: The set of valid US state abbreviation codes per ISO 3166-2.
    specification: '{"AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA", "HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD", "MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ", "NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC", "SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY"}'
  ProductCategory:
    description: Valid product categories in the catalog system.
    specification: '{"electronics", "clothing", "food", "furniture", "toys"}'
  PriorityLevel:
    description: Priority levels for task ordering.
    specification: '{1, 2, 3, 4, 5}'
```

These are conceptually different from global functions — they have no parameters, no `_` prefix convention, and their Logic (or equivalent) is always just a set expression. The name is used directly in TLA+ without any call syntax.

---

## Class Associations and Object Type Traversal

### The Scenario

Consider a class `Order` with an attribute `assigned_track` of type `object` pointing to class `Track`. An event triggers an action with a parameter `incoming_track` also of type `object` pointing to `Track`. A named association "Tracks" connects `Order` to `Track` (one-to-many). The action's TLA+ logic:

1. Uses `incoming_track` (the parameter) to filter over the "Tracks" association
2. Finds the matching Track instance by comparing a field
3. Reads values from the matched Track
4. Filters over "Tracks" again to find another match
5. Stores the result into the `assigned_track` attribute

In TLA+ this might look like:

```
LET matched == CHOOSE t \in self.Tracks : t.track_id = param.incoming_track.track_id
IN  LET target == CHOOSE t \in self.Tracks : t.region = matched.region /\ t.priority > matched.priority
    IN  self.assigned_track' = target
```

### What Exists Today in model_expression

The existing expression nodes handle this flow completely:

| Step | TLA+ Fragment | Expression Node | What It Produces |
|---|---|---|---|
| Navigate association | `self.Tracks` | `FieldAccess{Base: SelfRef{}, Field: "Tracks"}` | Set of Track instances |
| Access parameter | `param.incoming_track` | `LocalVar{Name: "incoming_track"}` | A Track instance (record) |
| Access field on instance | `t.track_id` | `FieldAccess{Base: LocalVar{Name: "t"}, Field: "track_id"}` | Field value |
| Filter set | `CHOOSE t \in S : P(t)` | `Quantifier` or `SetFilter` + selection | Single matching element |
| Compare values | `t.track_id = param.incoming_track.track_id` | `Compare{Op: "eq", Left: ..., Right: ...}` | Boolean |
| State assignment | `self.assigned_track' = target` | Logic with `Target: "assigned_track"` | State change |

The critical design: **`FieldAccess` does double duty** — it handles both attribute access and association traversal. The simulator distinguishes them at runtime by checking the `RelationContext`. When the field name matches an association name on the current class, the evaluator performs association traversal (returning a set of linked instances). When it doesn't match an association, it's a regular record field access.

This means **no new expression node types are needed** for association traversal. The `FieldAccess` node is structurally sufficient — the distinction between "field access" and "association traversal" is a semantic/type-level concern, not a structural one.

### What's Needed for ExpressionType

The type checker needs to resolve `self.Tracks` to a type. This requires knowing:

1. **That "Tracks" is an association, not an attribute** — The type checker must consult the association registry
2. **What class is on the other end** — `Track` in this case
3. **What multiplicity** — one-to-many means the result is a set; one-to-one means the result is a single instance

The result type of association traversal:

| Multiplicity on target end | Result ExpressionType |
|---|---|
| `0..1` or `1..1` (single) | `ObjectType{ClassKey: targetClassKey}` (nullable if 0..1) |
| `0..*`, `1..*`, etc. (many) | `SetType{ElementType: ObjectType{ClassKey: targetClassKey}}` |

This fits naturally into the `TypeContext` already proposed:

```go
type TypeContext struct {
    AttributeTypes    map[identity.Key]ExpressionType        // attribute key → type
    ParameterTypes    map[string]ExpressionType               // param name → type
    GlobalFunctions   map[identity.Key]FunctionType           // function key → signature
    LocalBindings     map[string]ExpressionType               // quantifier-bound variables
    AssociationTypes  map[string]AssociationTypeInfo           // NEW: association name → type info
}

type AssociationTypeInfo struct {
    TargetClassKey identity.Key
    ResultType     ExpressionType  // SetType{ObjectType{...}} or ObjectType{...}
    Multiplicity   Multiplicity
}
```

No new ExpressionType nodes are needed. `ObjectType{ClassKey}` already represents a class instance reference. `SetType{ElementType: ObjectType{ClassKey}}` represents a set of class instances. These compose with existing nodes — `FieldAccess` on an `ObjectType` resolves to the accessed attribute's type.

### How ObjectType Connects to DataType

The stakeholder DataType already has `Atomic{ConstraintType: "object", ObjectClassKey: "..."}`. The precise ExpressionType counterpart is `ObjectType{ClassKey: identity.Key}`. Compatibility checking verifies:

| DataType | ExpressionType | Compatible? |
|---|---|---|
| `object` of class A | `ObjectType{ClassKey: A}` | Yes — class keys must match |
| `object` of class A | `ObjectType{ClassKey: B}` | No — different classes |
| `unconstrained` | `ObjectType{ClassKey: A}` | Yes — unconstrained is compatible with anything |
| `span` / `enum` / `reference` | `ObjectType{ClassKey: A}` | No — non-object stakeholder type vs object precise type |

### Why No New Nodes Are Needed

The key insight is that associations are **relationships between classes**, not new data structures. In the expression tree:

- Navigating an association is structurally identical to accessing a field — it's a `FieldAccess` node. The fact that "Tracks" is an association rather than an attribute is resolved by the type checker consulting the association registry, not by a different node type.

- The result of association traversal (a set of class instances) is typed using existing type constructors: `SetType{ObjectType{...}}`. No new ExpressionType node is needed.

- Filtering, comparing, and accessing fields on the traversed instances all use existing nodes: `SetFilter`, `Compare`, `FieldAccess`.

- Storing an object reference into an object-typed attribute is a standard state_change Logic — the target is the attribute name, and the expression evaluates to an `ObjectType` value.

The only addition is that `TypeContext` must include association information so the type checker knows what `self.Tracks` returns. This is a type-checking context addition, not a model/expression change.

### Association Traversal and Reverse Traversal

The simulator supports both forward and reverse traversal:

- **Forward**: `self.Tracks` — navigates from Order to Track via the "Tracks" association
- **Reverse**: `self._Tracks` — navigates from Track back to Order (underscore prefix convention)

In the expression tree, both are `FieldAccess` nodes. The underscore prefix is part of the `Field` string. The type checker resolves `_Tracks` by looking up reverse associations in the AssociationTypeInfo.

The reverse result type follows the same multiplicity logic but uses the multiplicity from the opposite end of the association.

---

## Unit Compatibility: Documentation, Not Enforcement

Two attributes can share the same ExpressionType (`IntegerType`) but have incompatible DataType constraints — one measured in dollars, the other in meters. The question is whether the formal system should prevent expressions like `self.price + self.distance`.

**Decision: Units are stakeholder documentation, not formal enforcement.**

The `AtomicSpan.Units` field (e.g., `"dollar"`, `"meters"`) is metadata that helps stakeholders understand what a number represents. The expression type system does not track, propagate, or enforce unit compatibility. An expression like `self.price + self.distance` is:

- **Structurally valid** — both sides are `IntegerType`, the expression type checker accepts it
- **Semantically nonsensical** — caught by human review of the specification
- **Likely caught at runtime** — if the result is stored into an attribute with span constraints, the invariant checker will flag out-of-range values

This is the right trade-off because:

1. **Unit algebra is complex machinery for marginal benefit.** Determining that `price * quantity` produces "dollar-quantities" or that `price / 100` is still "dollars" requires a dimensional analysis system. The number of unit-crossing expressions in a typical model is small enough that human review catches errors.

2. **ExpressionType stays purely structural.** Adding unit tags to numeric types would blur the line between structural types (what the machine needs) and business constraints (what stakeholders care about). Units belong with DataType, which already carries the business context.

3. **The invariant checker provides a safety net.** After every state change, the `DataTypeChecker` validates that attribute values fall within their declared span bounds. An expression that combines incompatible units will typically produce a result outside the target attribute's expected range, surfacing the error during simulation.

4. **Consistency with the rest of the design.** Enumeration ordering, span bounds, and cardinality constraints are all DataType concerns that constrain what expressions are *valid for the business* but don't change the *structural type*. Units follow the same pattern.

This means two `IntegerType` values are always interchangeable from the expression type system's perspective, regardless of what units their DataTypes declare. The responsibility for ensuring semantic correctness falls to:
- The requirements author (human review)
- The AI parser (which can flag suspicious unit combinations as warnings in error `.md` files)
- The invariant checker (runtime bounds validation)

---

## Attribute Invariants

### Current State

The model already supports invariants at two levels:

- **Model-level invariants** (`Model.Invariants []Logic`) — Global assertions checked against the entire simulation state. These have no `self.` binding — they see global state only. Example: `\A order \in AllOrders : order.total >= 0`.

- **Class-level invariants** (`Class.Invariants []Logic`) — Per-instance assertions using `self.` to reference the instance's attributes. Example: `self.total = Sum({li.price : li \in self.lineItems})`. Key type: `cinvariant`. Database: `class_invariant` join table.

Both use `LogicTypeAssessment` — boolean predicates that must hold.

### What's Missing: Attribute-Level Invariants

Attributes currently have no invariant support. The only formal specification on an attribute is `DerivationPolicy` (a `LogicTypeValue` expression for computed attributes). But attributes have rich type constraints (span bounds, enumeration membership, cardinality) that are only enforced by the `DataTypeChecker` at runtime — they're not expressible as formal Logic specifications that the expression system can reason about.

With ExpressionType integration, attribute invariants become valuable: they let the author write formal constraints that tie the precise type, the stakeholder constraints, and the behavioral logic together.

### Design

Each attribute gains an optional list of invariants:

```go
type Attribute struct {
    Key              identity.Key
    Name             string
    Details          string
    DataTypeRules    string
    DataType         *model_data_type.DataType
    DerivationPolicy *model_logic.Logic
    Invariants       []model_logic.Logic          // NEW: attribute-level invariants
    Nullable         bool
    UmlComment       string
    IndexNums        []uint
}
```

Attribute invariants are `LogicTypeAssessment` — boolean predicates that must hold for the attribute's value. They use **`attribute`** as the TLA+ binding for the value being checked, analogous to how class invariants use `self.` for the instance.

The `attribute` binding resolves to the attribute's current value at evaluation time. This is simpler than `self.field_name` because:
1. The invariant is already scoped to a specific attribute — the binding name is unambiguous
2. For complex types (records, collections), `attribute` gives direct access: `attribute.name`, `Len(attribute)`, `attribute[1]`
3. It avoids repeating the attribute name in every invariant

### Examples

```yaml
attributes:
  balance:
    name: Account Balance
    rules: [0 .. 1000000] at 0.01 dollar
    precise_type: Nat
    invariants:
      - description: Balance must be non-negative
        specification: "attribute >= 0"
      - description: Balance must not exceed limit
        specification: "attribute <= 1000000"

  items:
    name: Line Items
    rules: ordered of unconstrained
    precise_type: Seq([name: STRING, price: Nat])
    invariants:
      - description: All items must have non-empty names
        specification: "\\A i \\in DOMAIN attribute : Len(attribute[i].name) > 0"
      - description: All prices must be non-negative
        specification: "\\A i \\in DOMAIN attribute : attribute[i].price >= 0"

  status:
    name: Order Status
    rules: enum of new, processing, complete
    precise_type: '{"new", "processing", "complete"}'
    invariants:
      - description: Status must be a valid value
        specification: 'attribute \in {"new", "processing", "complete"}'
```

### Identity Keys and Database

New key type: `KEY_TYPE_ATTRIBUTE_INVARIANT = "ainvariant"` — child of the attribute key.

New join table:
```sql
CREATE TABLE attribute_invariant (
    model_key     text NOT NULL,
    attribute_key text NOT NULL,
    logic_key     text NOT NULL,
    PRIMARY KEY (model_key, attribute_key, logic_key),
    CONSTRAINT fk_attr_inv_attribute
        FOREIGN KEY (model_key, attribute_key) REFERENCES attribute (model_key, attribute_key) ON DELETE CASCADE,
    CONSTRAINT fk_attr_inv_logic
        FOREIGN KEY (model_key, logic_key) REFERENCES logic (model_key, logic_key) ON DELETE CASCADE
);
```

### Simulator Integration

The invariant checker evaluates attribute invariants per attribute per instance:
1. For each class instance, for each attribute with invariants
2. Bind `attribute` to the attribute's current value
3. Evaluate each invariant Logic specification
4. Report violations

This extends the existing `DataTypeChecker.CheckInstance()` pattern — after checking DataType constraints (span bounds, enum membership), check attribute invariants.

### Relationship to DataType Constraints

Attribute invariants and DataType constraints overlap but serve different purposes:

- **DataType constraints** are stakeholder-facing — parsed from `rules:` text, stored in the DataType structure, checked by the `DataTypeChecker`. They require no TLA+ knowledge.
- **Attribute invariants** are formal specifications — TLA+ expressions with the `attribute` binding. They can express constraints that DataType cannot (cross-field dependencies, conditional constraints, quantified predicates over collection elements).

Both are checked at runtime. DataType constraints are always derived automatically from the stakeholder's `rules:` field. Attribute invariants are explicitly authored.

---

## Infrastructure Gaps and Prerequisites

A thorough analysis of the codebase reveals gaps between what the design assumes and what currently exists, plus additional concerns discovered through deep exploration of the full `apps/requirements/req/` tree.

### Gap 1: Type Expression Parsing

The design proposes that humans write TLA+ set expressions like `_Set(STRING)`, `_Seq(Nat)`, `_Bag([name: STRING])` for precise types. These use the existing module call syntax (`_Module:Function(args)` or `_Module(args)`) which the parser already handles. **Most type constructors parse as standard function calls — no new grammar is needed for them.**

However, two constructs need attention:

1. **Record type syntax** `[name: STRING, age: Nat]` uses colon syntax, while record **values** use `[name |-> "Alice"]`. The parser currently only handles the value syntax. Record types need either a grammar extension or an alternate notation.

2. **Cartesian product** `Nat \X STRING` is not in the grammar. Could use a helper like `_Tuple(Nat, STRING)` instead.

**Resolution:** Use the custom constructor pattern consistently. Record types can be written as `_Record(name: STRING, age: Nat)` or the grammar can be extended for the `[field: Type]` syntax since this is the standard TLA+ record type notation and is distinct from the `[field |-> value]` syntax. The parser already distinguishes `|->` vs `:` — it just doesn't handle the colon variant yet.

### Gap 2: Simulator Type System — Purpose and Relationship

The simulator has its own type system (`simulator/types/`) with `Boolean`, `Number`, `String`, `Set`, `Tuple`, `Record`, `Bag`, `Function`, `TypeVar`, and `Any`. This exists because the simulator uses **Hindley-Milner type inference** — a fundamentally different mechanism from declared types.

The HM type system adds three things that declared types don't provide:

1. **TypeVar (type variables)**: Enable polymorphic builtins. `_Seq:Head` has signature `∀a. Tuple[a] → a`. When called with `_Seq:Head([1,2,3])`, the type variable `a` is unified to `Number`. Without TypeVars, every builtin would need per-type overloads.

2. **Unification**: Solves type constraints. When an expression like `{x \in S : x > 0}` is type-checked, unification deduces that `S` must be `Set[Number]` because `>` requires numbers.

3. **Type schemes**: Enable let-polymorphism for definitions used in multiple contexts.

**Relationship to ExpressionType:** The simulator type system is the **runtime inference engine**. ExpressionType is the **stored declaration**. They serve different purposes:

| System | Purpose | Analogy |
|---|---|---|
| `model_data_type.DataType` | Stakeholder constraints (ranges, cardinality, enums) | "Business rules" |
| `model_expression_type.ExpressionType` | Declared structural types stored in the model | "Type annotations" |
| `simulator/types.Type` | Inferred types during expression evaluation | "Type inference engine" |

When ExpressionType is available for an attribute/parameter, the simulator can use it as a **type hint** — pre-populating the type environment with declared types rather than inferring from scratch. This makes inference faster and catches more errors, but doesn't replace inference because:
- Not all attributes will have ExpressionType immediately (incremental adoption)
- Interior expression nodes still need inferred types (declared types are only on declarations)
- Polymorphic builtins still need TypeVar + unification regardless

**The simulator type system stays.** ExpressionType feeds into it as declared type hints, not as a replacement.

### Gap 3: LET/IN — Shared Local Variables Across Logic

LET/IN is not currently implemented in parser, AST, or evaluator. But the need goes beyond the association scenario.

**The real use case:** An Action has multiple Guarantees (each a separate Logic specification), plus Requirements and SafetyRules. These separate specs often need shared intermediate values. Today, each spec is independent — if two guarantees need the same filtered set, the filtering must be duplicated.

LET/IN should be a **shared scope mechanism** at the Action/Query level, not embedded within a single Logic spec:

```yaml
actions:
  assign_track:
    local_definitions:                          # NEW: shared LET bindings
      - name: matched_track
        specification: "CHOOSE t \in self.Tracks : t.id = param.track_id"
      - name: related_tracks
        specification: "{t \in self.Tracks : t.region = matched_track.region}"
    requires:
      - description: Track must exist
        specification: "matched_track /= NULL"
    guarantees:
      - description: Assign the matched track
        target: assigned_track
        specification: "matched_track"
      - description: Update related count
        target: related_count
        specification: "Cardinality(related_tracks)"
    safety_rules:
      - description: Track must remain valid
        specification: "self.assigned_track' \in self.Tracks"
```

This means `local_definitions` is a new field on Action and Query — a list of named value definitions (Logic of type "value") that are evaluated once and available to all requires, guarantees, and safety rules. In the model, these would be `[]model_logic.Logic` with `LogicTypeValue`.

**In model_expression:** A `LocalDefinition` or `LetBinding` node type may be needed for the expression tree, or the local definitions could remain as separate Logic objects with their own expression trees, referenced by name via `LocalVar` nodes in the other specs.

**For the simulator:** Local definitions are evaluated before the requires/guarantees/safety rules, and their values are added to the bindings scope.

### Gap 4: CHOOSE — Not Needed

CHOOSE is TLA+'s mechanism for non-deterministic selection from a set satisfying a predicate. In TLA+ formal proofs, CHOOSE returns a specific but unspecified value. **This system doesn't do formal proofs — the simulator randomly selects values.**

CHOOSE is not needed because:
- The simulator already handles non-determinism by randomly picking parameter values, event orderings, and action selections
- Where a spec needs "find a matching element," the filtering logic (`{x \in S : P(x)}`) already produces the candidate set, and the simulator picks from it
- Adding CHOOSE would imply deterministic selection semantics that don't match the simulator's random exploration approach

The association scenario example should be rewritten without CHOOSE, using set filtering instead:

```
\* Instead of CHOOSE, filter to a set and let the simulator pick
LET candidates == {t \in self.Tracks : t.id = param.track_id}
IN  self.assigned_track' = candidates
```

Or the action's guarantee directly assigns from the filtered set, and the simulator's random value selection picks an element.

### Gap 5: ObjectClassKey Should Be identity.Key

`Atomic.ObjectClassKey` is currently `*string`. It should be `*identity.Key` to match the rest of the model's key system. This is a cleanup that should happen when ExpressionType is integrated — `ObjectType{ClassKey identity.Key}` already uses the correct type.

### Gap 6: Enumeration Ordering

DataType's `EnumOrdered *bool` flag determines whether `<`, `>`, `<=`, `>=` comparisons are valid on enum values. **By default, enumerations should be assumed unordered** — comparison operators beyond equality should only be allowed when `EnumOrdered` is explicitly true.

The expression type checker (when integrated) should enforce this: `Compare{Op: "lt"}` on an `EnumType` is only valid if the corresponding DataType has `EnumOrdered = true`. This is another example of constraints on DataType affecting what expressions are valid.

### Gap 7: Nullable Attributes

`Attribute.Nullable bool` indicates an attribute can be absent/null. ExpressionType should remain non-nullable — nullability is a DataType constraint, not a structural type concern. The invariant checker already validates null constraints at runtime.

### Gap 8: Derivation Policy Type Checking

`Attribute.DerivationPolicy *Logic` computes a derived attribute's value. For a code generator, this is essentially a short function definition. The type-checking approach should be the same as for any function: the expression's result type must match the attribute's declared ExpressionType (when both exist).

### Gap 9: Association Class Attributes

Associations can have an association class (`Association.AssociationClassKey *identity.Key`). When traversing an association, accessing the link data (attributes of the association class) likely needs the LET/IN mechanism from Gap 3. A local definition could bind the link data for use in subsequent logic:

```yaml
local_definitions:
  - name: link
    specification: "self.Tracks_link(param.target)"  # hypothetical accessor
```

This is a future concern — the current simulator doesn't support association class attribute access.

### Gap 10: Generation Templates Consume DataType

Templates in `internal/generate/` actively render DataType via `{{ data_type_rules .DataTypeRules .DataType }}`. When ExpressionType is added, templates may need to render precise type information alongside DataTypeRules for generated documentation. The `req_flat` layer would need to expose ExpressionType in its lookups.

### Gap 11: Parameter Type Enforcement

Action and query parameters have DataTypes but these are **never type-checked during expression evaluation**. The simulator type checker infers parameter types from usage in expressions, independent of the declared DataType. When a parameter's DataType carries an ExpressionType, the type checker should use it as a constraint — ensuring the expression treats the parameter consistently with its declared type. Since ExpressionType now lives on DataType, parameters get this automatically through `param.DataType.ExpressionType`.

### Gap 12: No Function Values at Runtime

The simulator's `object` package has no `Function` type for runtime values. The type system has `types.Function` for type-checking, but the evaluator cannot produce or store function values. This doesn't block ExpressionType integration (types are declarations, not values), but limits higher-order patterns in expressions.

### Gap 13: Scenario Test Data Typing

Scenario objects (`model_scenario/object.go`) reference classes but carry no test data values. When scenarios need concrete test data (attribute values for simulation), those values must conform to both DataType constraints and ExpressionType structure. This is a future concern — scenarios currently test behavioral flows, not data validation.

---

## Implementation Considerations

### Build Order

Following the project's standard build order (Model → test_helper → Database → Parsers → Simulator):

1. **Formal spec value objects** — Define `ExpressionSpec` and `TypeSpec` as reusable Notation+Specification+Tree trios
2. **model_expression_type package** — Type constructors, validation, TypeName serialization
3. **DataType and Logic updates** — Add optional `TypeSpec` to DataType, refactor Logic to use `ExpressionSpec`, add `TargetTypeSpec` to Logic
4. **Database schema** — expression_type table, FK on data_type and logic tables
5. **Database data access** — Flatten/rebuild cycle for expression types (same pattern as expression_node)
6. **Compatibility checker** — DataType stakeholder constraints ↔ TypeSpec.ExpressionType validation
7. **Parser updates** — AI parser gets TypeSpec in its schema; human parser reads `precise_type:` field
8. **TLA+ type expression grammar** — Extend PEG grammar with type expression rules (prerequisite for human-authored precise types)
9. **Type checker integration** — Use declared ExpressionTypes when available; bridge to simulator/types
10. **LET/IN** — Implement as shared local variable scope (prerequisite for complex expression patterns)

### Blast Radius

TypeSpec on DataType is a pointer (nil = no precise type). TargetTypeSpec on Logic is the same. Refactoring Logic to use ExpressionSpec groups existing fields — no semantic change. This means:
- All existing code continues to work unchanged — DataType objects without TypeSpec behave identically
- DataType's constructor gains an optional TypeSpec parameter (nil for all existing callers)
- Logic's constructor change is structural (grouping existing fields into ExpressionSpec) — callers change shape but not meaning
- Database migration adds nullable columns — no data migration needed
- Tests don't need updating until they exercise type features

### Alternative Considered: Reusing DataType

One alternative is to extend DataType with precise-type capabilities rather than creating a new type system. This was rejected because:

1. DataType's text parser would need fundamental changes to handle types like `function(integer, integer) → boolean` or `bag of record {name: string, count: integer}`.
2. DataType's validation assumes human-authored constraints (span precision, cardinality bounds). Precise types have no such constraints — they are structural only.
3. DataType's database schema (five tables) encodes the constraint hierarchy. Precise types need a simpler tree structure.
4. Merging the two would make DataType's API confusing — callers would need to know whether they're working with a stakeholder type or a precise type.

### Alternative Considered: Embedding DataType in Expression Nodes

Another alternative is to add a `DataType` field to expression nodes that produce typed values. This was rejected because:
1. Types belong to declarations (attributes, parameters), not to individual expression nodes.
2. Interior expression nodes have types derived from their children — storing types on every node is redundant and error-prone.
3. This would massively increase the expression_node table width and row count.

---

## Summary

| Aspect | model_data_type | model_expression_type |
|---|---|---|
| **Audience** | Stakeholders | Code generators, model compilers |
| **Authoring** | Human text parsing (PEG grammar) | Human writes TLA+ set expression, AI parser, type inference |
| **Completeness** | Can be partial ("unconstrained") | Must be fully resolved |
| **Constraints** | Business constraints (ranges, enums, cardinality) | Structural only (no business constraints) |
| **Collection model** | Access-pattern-oriented (stack, queue, ordered, unordered) | Structure-oriented (sequence, set, bag) |
| **Uniqueness** | `CollectionUnique` flag on collection | SetType (inherently unique), SequenceType.Unique flag, BagType (inherently non-unique) |
| **Record model** | Nested DataType fields | Nested ExpressionType fields |
| **Where stored** | attribute.data_type_key, parameter.data_type_key | data_type.expression_type_key (via TypeSpec on DataType), logic.target_type_key |
| **Database pattern** | 5-table specialization hierarchy | Single adjacency-list table |
| **Notation** | Has a TLA+-independent text syntax (PEG grammar) | TLA+ set expressions: `Seq(Int)`, `SUBSET STRING`, `[name: STRING]` |
| **Authoring surface** | Human markdown files (`rules:` field), AI parser | Human markdown files (`precise_type:` field), AI parser |
| **Authoring wrapper** | DataTypeRules string → parsed DataType | TypeSpec on DataType (Notation + Specification + ExpressionType) |

The two type systems are complementary, not competing. DataType answers "what does the stakeholder need to understand?" while ExpressionType answers "what does the machine need to generate code?" The compatibility checker bridges them, ensuring that the precise definition is consistent with the stakeholder's expectations.

The `CollectionUnique` flag on DataType maps to the precise type system as follows: for `unordered` collections, unique → `SetType`, non-unique → `BagType`. For `ordered`/`stack`/`queue` collections, uniqueness is carried as the `Unique` flag on `SequenceType`. The custom TLA+ constructors (`_SeqUnique`, `_StackUnique`, `_QueueUnique`) give human authors a clear way to express this without writing complex quantified predicates, and give code generators an explicit signal to enforce element distinctness.

For `reference` atomic data types, the stakeholder-facing DataType carries a human-readable citation (`ref from ISO 3166-2 US state abbreviation codes`) while the precise type references a model-level `NamedSet` — a new entity that holds a TLA+ set definition (e.g., `{"AL", "AK", ...}`). Named sets are distinct from global functions: they are named constants, not computations. They can be referenced by name in both type expressions and behavioral logic throughout the model.
