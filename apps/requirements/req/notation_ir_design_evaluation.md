# Design Evaluation: Pure IR Between Notation and Simulator

## The Proposal (Revised)

Introduce an intermediate representation (IR) — a "pure AST" — that sits between notation-specific parsing and everything downstream: simulation, validation, database storage, and code generation.

```
TLA+ text ──parse──► TLA+ AST ──lower──► Model Expression ──► simulator evaluation
                                          ▲    │
                  raise (generate TLA+) ──┘    ├──► database (adjacency list with FKs)
                                               │
AI JSON ──parser_ai──► inputExpression ────────├──► semantic validation (class/attribute references)
                                               │
                                               └──► code generator (expression → language-independent logic
                                                                               → language-dependent code)
```

Key clarifications from design discussion:

- **TLA+ is generated from expressions, not hand-authored by AI.** A raising pass (`notation/tla/raising/`) generates canonical TLA+ from expressions for the `Specification` field. Display/rendering ultimately goes into a code generator.
- **The IR enables semantic validation.** Every reference to a class, attribute, or association in an expression can be validated against the model — like a foreign key check. References use `identity.Key`, not strings.
- **Primed (`x'`) is eliminated for state changes.** The `model_logic.Logic.Target` field already identifies what is being assigned. Primed notation in the IR is only needed for safety rules, where it means "compare next-state vs prior-state."
- **ExistingValue (`@`) becomes `PriorFieldValue`.** The TLA+ `@` means "the value the field had before this update." In the IR this is made explicit as `PriorFieldValue(field)` — the current value of the field being altered in a record update context.
- **The IR is mathematical, not imperative.** It represents declarative logic suitable for code generation into any target language. It is not a sequence of statements.
- **The IR lives under `req_model`** as part of the core model, not as a separate top-level package.
- **Database storage uses the adjacency list pattern** (like `scenario_step`), not serialized JSON. This enables foreign keys from expression nodes to model entities.
- **parser_ai exposes the IR as JSON.** AI can author logic using pure IR expressions (as `inputExpression`) without knowing TLA+. The parser_ai reads expressions in, generates the equivalent TLA+ for the `Specification` field, and returns the full model. The JSON schema for expressions must be highly explanatory so AI can self-correct bad logic.
- **Logic has two construction modes.** A Logic can be constructed with either (a) a specification string and no expression, or (b) an expression and no specification — never both. Spec-in triggers parse→lower→expression. Expression-in triggers generate→TLA+ specification. If spec parsing fails (e.g., human working out typos), the Logic is still valid — it just has no expression yet.
- **Test models use expression-first construction.** Both `GetTestModel()` and `GetStrictTestModel()` build all Logic objects from expressions, not specification strings. This ensures parser_ai round-trip tests never have a mismatch on TLA+ specifications — the TLA+ is always generated deterministically from the expression.

## Verdict: Good Idea

This is a **well-established compiler design pattern** (multi-frontend / shared-backend via IR) and it fits your architecture well. The code generation use case makes it mandatory — you cannot generate code from a TLA+-specific AST without coupling every code generator to TLA+ grammar.

---

## Why It's Sound

### 1. The Pattern Is Proven

This is exactly the architecture behind LLVM (multiple frontends → LLVM IR → multiple backends), GCC (multiple frontends → GIMPLE → multiple targets), and MLIR (extensible dialect-based IR). The core insight: if you have `M` source notations and `N` backends (simulation, code generation, validation), a shared IR reduces the problem from `M × N` translators to `M + N`.

Your backends are: simulation, semantic validation, code generation (language-independent logic), and future code generation (language-dependent code). That's enough `N` to justify the IR even with `M = 1`.

### 2. Semantic Validation Requires It

The current system stores TLA+ specification strings and parses them on demand. You cannot validate that `self.status` refers to a real attribute, or that `Order!PlaceOrder(...)` refers to a real class and action, without parsing the string first. The IR makes these references first-class — they carry `identity.Key` values that can be validated against the model like foreign keys in a database.

This is a significant correctness improvement. Today a typo in a specification string is only caught at simulation time. With the IR, it's caught at parse/lower time.

### 3. Code Generation Requires It

You cannot generate Go/Java/Python from a TLA+-specific AST without every code generator knowing TLA+ grammar. The IR provides a clean, notation-independent input for code generation:

```
Expression  ──►  language-independent logic  ──►  language-dependent code
                 (code gen phase 1)               (code gen phase 2)
```

This two-phase code generation matches how real compilers work: the IR is lowered to a target-independent form, then to target-specific code.

### 4. Your Existing AST Is Already Close

Looking at `notation/tla/ast/` (current `notation/ast/`), the current AST is already fairly abstract:
- Operators are stored as semantic tokens (`"∧"`, `"∨"`), not parse artifacts
- The node types map to *semantic* concepts (quantifier, membership, set filter) not *syntactic* ones
- The evaluator dispatches on AST node types, meaning the AST is already the evaluator's "instruction set"

The gap between your current AST and a pure model expression is smaller than it might seem.

---

## Key Design Decisions

### 1. Primed Variables: Split Into Two Concepts

Currently TLA+ uses `x'` for two different things:
- **State change targets**: `self.count' = self.count + 1` — the `count'` says "this is the new value of count"
- **Safety rule comparisons**: `self.count' >= self.count` — the `count'` means "the next-state value of count"

In the model expression, these are separated:

**State changes** don't need primed at all. The `Logic.Target` field already identifies the attribute being assigned. The expression for the guarantee is just the *value expression*:
```
Logic.Target = identity.Key for "count" attribute
Expression   = Add(FieldAccess(Self, "count"), Literal(1))
```

**Safety rules** do need a way to reference both prior and next state. The expression uses explicit `NextState(ref)` nodes:
```
Expression = Gte(NextState(FieldAccess(Self, "count")), FieldAccess(Self, "count"))
```

The TLA+ lowering pass maps `x'` to `NextState(x)` in safety rule context, and extracts the target from `x' = expr` patterns in guarantee context (discarding the primed notation).

### 2. ExistingValue (`@`) Becomes `PriorFieldValue`

In TLA+, `@` appears inside `EXCEPT` expressions:
```
[record EXCEPT !.count = @ + 1]
```

The `@` means "the current value of the field being updated." This is a TLA+-specific shorthand. The model expression equivalent is `PriorFieldValue`, which makes the semantics explicit:

```
RecordUpdate(record, "count", Add(PriorFieldValue("count"), Literal(1)))
```

This is clearer for code generation — every target language has a way to express "take the old value and compute a new one" but `@` is TLA+-specific notation.

For the evaluator, `PriorFieldValue` maps to the same `existingValue` binding mechanism currently used for `@`. The concept is the same; the expression just names it explicitly.

### 3. References Use `identity.Key`, Not Strings

The model expression replaces string-based references with typed keys:

| Current (TLA+ AST) | Model Expression |
|---|---|
| `Identifier{Name: "self"}` | `SelfRef` (implicit current instance) |
| `FieldAccess{Base: self, Member: "count"}` | `AttributeRef(attributeKey)` where `attributeKey` is an `identity.Key` |
| `FunctionCall{Name: "PlaceOrder", ScopePath: ["orders", "mgmt", "order"]}` | `ActionCall(actionKey, args)` where `actionKey` is an `identity.Key` |
| `ScopedCall{Domain: "d", Subdomain: "s", Class: "c", Function: "f"}` | `ActionCall(actionKey, args)` — same normalized form |

This enables validation: when constructing the expression, the lowering pass resolves every name to a key. If the attribute or class doesn't exist, lowering fails with a clear error. In the database, these keys become foreign keys.

### 4. The Expressions Are Declarative/Mathematical

The expressions represent pure mathematical logic, not imperative sequences. This is critical for code generation — the same expression can be translated to:
- A Python expression
- A SQL predicate
- A Go function body
- A formal verification assertion

There are no side effects. State changes are expressed as "the new value of X is [expression]" not "set X to [expression]." The code generator decides how to realize the state change in the target language.

---

## Recommended Architecture

### Folder Structure

```
notation/
├── tla/                           # TLA+ notation (moved from notation/ast/; parser moved from simulator/parser/)
│   ├── ast/                       #   TLA+ specific AST (concrete syntax tree)
│   ├── parser/                    #   TLA+ text → TLA+ AST  (moved from simulator/parser/)
│   ├── lowering/                  #   TLA+ AST → model expression (NEW)
│   └── raising/                   #   model expression → TLA+ text (NEW)

req_model/
├── model_logic/                   # Existing: Logic, GlobalFunction
├── model_expression/              # NEW: notation-independent expression tree
│   ├── node.go                    #   Expression interface, node types
│   ├── ops.go                     #   Operator enums (ArithOp, LogicOp, CompareOp, etc.)
│   ├── validate.go                #   Structural validation
│   └── validate_model.go          #   Model-aware validation (key existence checks)
├── model_class/                   # Existing
├── model_domain/                  # Existing
├── model_state/                   # Existing
└── ...

simulator/
├── evaluator/                     # Works on model_expression, NOT notation/tla/ast
├── typechecker/                   # Works on model_expression
├── model_bridge/                  # model → expression (calls notation/tla/lowering)
├── registry/                      # Stores model_expression.Expression
└── ...

generate/                          # FUTURE: code generation from model expressions
└── ...
```

### Why `req_model/model_expression/`

The model expression is part of the core model — it's the structured form of what `Logic.Specification` currently holds as a string. Just as `model_class` defines classes and `model_state` defines states, `model_expression` defines the formal logic expressions attached to model elements. It belongs with the model, not as a separate top-level package.

### Why `notation/tla/`

Grouping the TLA+ AST, parser, and lowering under `notation/tla/` makes the notation-specific boundary clear. When a second notation is added (e.g., `notation/z/` or `notation/alloy/`), each notation gets its own sub-tree with the same structure: `ast/`, `parser/`, `lowering/`.

---

## Database Storage: Adjacency List

The expression tree is stored relationally using the **adjacency list** pattern — the same pattern used by the `scenario_step` table: each row has a `parent_node_key` FK pointing back to the same table, plus `sort_order` for child ordering.

### Schema

```sql
CREATE TYPE expression_node_type AS ENUM (
    -- Literals
    'bool_literal', 'int_literal', 'rational_literal', 'string_literal',
    'set_literal', 'tuple_literal', 'record_literal', 'set_constant',
    -- References
    'self_ref', 'attribute_ref', 'local_var', 'prior_field_value', 'next_state',
    -- Binary operators
    'binary_arith', 'binary_logic', 'compare', 'set_op', 'set_compare',
    'bag_op', 'bag_compare', 'membership',
    -- Unary operators
    'negate', 'not',
    -- Collections
    'field_access', 'tuple_index', 'record_update', 'field_alteration',
    'string_index', 'string_concat', 'tuple_concat',
    -- Control flow
    'if_then_else', 'case', 'case_branch',
    -- Quantifiers
    'quantifier', 'set_filter', 'set_range',
    -- Calls
    'action_call', 'global_call', 'builtin_call'
);

CREATE TABLE expression_node (
    model_key           text NOT NULL,
    expression_node_key text NOT NULL,           -- unique ID for this node
    logic_key           text NOT NULL,           -- which Logic this expression belongs to
    parent_node_key     text DEFAULT NULL,       -- NULL for root node
    sort_order          int NOT NULL,            -- ordering among siblings
    node_type           expression_node_type NOT NULL,

    -- Scalar values (used by literals and leaf nodes, NULL otherwise)
    bool_value          boolean DEFAULT NULL,
    int_value           bigint DEFAULT NULL,
    numerator           bigint DEFAULT NULL,     -- for rational_literal
    denominator         bigint DEFAULT NULL,     -- for rational_literal
    string_value        text DEFAULT NULL,       -- for string_literal, local_var, field_access.field, etc.

    -- Operator enums (used by binary/unary operator nodes)
    operator            text DEFAULT NULL,       -- 'add', 'sub', 'and', 'or', 'lt', 'eq', etc.

    -- Model references (foreign keys to other model entities)
    attribute_key       text DEFAULT NULL,       -- for attribute_ref
    action_key          text DEFAULT NULL,       -- for action_call
    global_function_key text DEFAULT NULL,       -- for global_call
    builtin_module      text DEFAULT NULL,       -- for builtin_call
    builtin_function    text DEFAULT NULL,       -- for builtin_call

    -- Quantifier metadata
    quantifier_kind     text DEFAULT NULL,       -- 'forall' or 'exists'
    variable_name       text DEFAULT NULL,       -- bound variable name

    -- Set constant kind
    set_constant_kind   text DEFAULT NULL,       -- 'nat', 'int', 'real', 'boolean'

    -- Membership negation
    negated             boolean DEFAULT NULL,    -- for membership: ∈ vs ∉

    PRIMARY KEY (model_key, expression_node_key),

    -- Tree structure
    CONSTRAINT fk_expr_parent FOREIGN KEY (model_key, parent_node_key)
        REFERENCES expression_node (model_key, expression_node_key) ON DELETE CASCADE,

    -- Link to the logic this expression belongs to
    CONSTRAINT fk_expr_logic FOREIGN KEY (model_key, logic_key)
        REFERENCES logic (model_key, logic_key) ON DELETE CASCADE,

    -- Model reference foreign keys
    CONSTRAINT fk_expr_attribute FOREIGN KEY (model_key, attribute_key)
        REFERENCES attribute (model_key, attribute_key) ON DELETE CASCADE,
    CONSTRAINT fk_expr_action FOREIGN KEY (model_key, action_key)
        REFERENCES action (model_key, action_key) ON DELETE CASCADE,
    CONSTRAINT fk_expr_global_function FOREIGN KEY (model_key, global_function_key)
        REFERENCES global_function (model_key, global_function_key) ON DELETE CASCADE
);
```

### Why Adjacency List Works Here

**Advantages:**
- Foreign keys from expression nodes to model entities (`attribute_key → attribute`, `action_key → action`). This is the primary motivation — the database enforces referential integrity on expression references.
- Cascading deletes work naturally: delete a Logic, all its expression nodes disappear.
- Familiar pattern in the codebase (same as `scenario_step`).
- Each node is a single row, easy to query and update individually.

**Potential concerns and mitigations:**

1. **Loading performance**: Reconstructing a tree from flat rows requires either recursive CTEs or loading all nodes for a logic and assembling in-memory. For expressions (typically 5-50 nodes per Logic), this is negligible. The `scenario_step` table handles the same pattern at similar scale.

2. **Write amplification**: Inserting an expression means one INSERT per node. For a 20-node expression, that's 20 INSERTs. This is fine — these writes happen at parse/import time, not in hot paths. And the current system already writes one row per Logic; replacing the `specification` text with 20 nodes is a modest increase.

3. **Node count**: A complex expression might have 50-100 nodes. Across a model with hundreds of Logic entries, that's ~10K-50K expression nodes. This is well within relational database comfort zones.

4. **Wide nullable columns**: The `expression_node` table has many nullable columns because different node types use different fields. This is the same trade-off as a "single-table" entity design. The alternative — one table per node type — would be normalized but require 30+ tables and complex JOINs to reconstruct a tree. The single-table approach with nullable columns is pragmatic and matches how `scenario_step` works (it has `event_key`, `query_key`, `scenario_ref_key`, etc., most of which are NULL for any given row).

5. **Ordering**: Children of a node are ordered by `sort_order` (same as `scenario_step`). For binary operators, `sort_order = 0` is left, `sort_order = 1` is right. For lists (set literal elements, function args), ordering is natural.

### Tree Reconstruction

Loading an expression tree from the database:

```go
// Load all nodes for a given logic, ordered by parent for efficient assembly.
rows := db.Query(`
    SELECT expression_node_key, parent_node_key, sort_order, node_type, ...
    FROM expression_node
    WHERE model_key = $1 AND logic_key = $2
    ORDER BY parent_node_key NULLS FIRST, sort_order
`, modelKey, logicKey)

// Build node map, then assemble tree by wiring children to parents.
```

This is the same pattern used for loading scenario step trees.

---

## What the Expression Node Set Looks Like

### Literals
- `BoolLiteral(value bool)`
- `IntLiteral(value int64)`
- `RationalLiteral(numerator, denominator int64)`
- `StringLiteral(value string)`
- `SetLiteral(elements []Expression)`
- `TupleLiteral(elements []Expression)`
- `RecordLiteral(fields map[string]Expression)`
- `SetConstant(kind)` — Nat, Int, Real, Boolean

### References (model-aware, validated, FK-backed in DB)
- `SelfRef` — the current instance
- `AttributeRef(attributeKey identity.Key)` — reference to a class attribute (FK to `attribute` table)
- `LocalVar(name string)` — quantifier-bound or parameter-bound variable
- `PriorFieldValue(field string)` — the value of a field before a record update (replaces `@`)
- `NextState(expr Expression)` — the next-state value of an expression (safety rules only)

### Arithmetic (typed enums, not strings)
- `BinaryArith(op ArithOp, left, right Expression)` — Add, Sub, Mul, Div, Mod, Pow
- `Negate(expr Expression)`

### Logic
- `BinaryLogic(op LogicOp, left, right Expression)` — And, Or, Implies, Equiv
- `Not(expr Expression)`

### Comparison
- `Compare(op CompareOp, left, right Expression)` — Lt, Gt, Lte, Gte, Eq, Neq

### Set Operations
- `SetOp(op SetOp, left, right Expression)` — Union, Intersect, Difference
- `SetCompare(op SetCompareOp, left, right Expression)` — Subset, Superset, ProperSubset, ProperSuperset
- `Membership(element, set Expression, negated bool)`
- `SetFilter(variable string, set Expression, predicate Expression)`
- `SetRange(start, end Expression)`

### Bag Operations
- `BagOp(op BagOp, left, right Expression)` — Sum, Difference
- `BagCompare(op BagCompareOp, left, right Expression)`

### Collections
- `FieldAccess(base Expression, field string)`
- `TupleIndex(tuple, index Expression)`
- `RecordUpdate(base Expression, alterations []FieldAlteration)` — replaces EXCEPT
- `StringIndex(str, index Expression)`
- `StringConcat(operands []Expression)`
- `TupleConcat(operands []Expression)`

### Control Flow (still declarative — these are expressions, not statements)
- `IfThenElse(condition, then, else Expression)`
- `Case(branches []CaseBranch, otherwise Expression)`

### Quantifiers
- `Quantifier(kind QuantifierKind, variable string, domain Expression, predicate Expression)` — Forall, Exists

### Calls (model-aware, validated, FK-backed in DB)
- `ActionCall(actionKey identity.Key, args []Expression)` — call a class action/query (FK to `action` table)
- `GlobalCall(functionKey identity.Key, args []Expression)` — call a global function (FK to `global_function` table)
- `BuiltinCall(module, function string, args []Expression)` — call a built-in (Seq, Len, etc.)

### Absent from expression (notation-specific or superseded)
- ~~`Parenthesized`~~ — precedence is structural (tree shape)
- ~~`ExistingValue (@)`~~ — replaced by `PriorFieldValue`
- ~~`Primed`~~ — replaced by `NextState` (safety rules) or `Logic.Target` (guarantees)
- ~~`ScopedCall`~~ — replaced by normalized `ActionCall` with `identity.Key`
- ~~`Assignment`~~ — state change target is in `Logic.Target`, not in the expression
- ~~`Fraction`~~ — replaced by `RationalLiteral` (the value, not the notation)

---

## Expression Validation

The model expression enables model-level validation that is impossible with string specifications:

1. **Attribute existence**: Every `AttributeRef(key)` can be checked — does this attribute exist in the class? In the DB, the FK constraint enforces this.
2. **Type compatibility**: The type checker can verify that `AttributeRef(countKey)` is numeric before allowing `Add`.
3. **Action existence**: Every `ActionCall(key)` can be checked — does this action exist? Are the parameter counts correct? FK-enforced in DB.
4. **Cross-class references**: When an expression references another class's attributes (via associations), the expression can validate the association exists and the multiplicity allows it.
5. **Safety rule well-formedness**: The expression can verify that safety rules use `NextState` and that guarantees do not.

This makes the expression function like a "compiled" form where name resolution has already happened — similar to how a compiler resolves symbols during lowering.

---

## Two-Level IR for Code Generation

Given that the next milestone is code generation, consider two levels:

### Level 1: Model Expression (what this document describes)
- Tied to the req_model domain (uses `identity.Key`, knows about classes/attributes)
- Validated against the model
- Stored in the database with foreign keys
- Input to the simulator evaluator
- Package: `req_model/model_expression/`

### Level 2: Language-Independent Logic (future, for code gen)
- Untied from the req_model — works in terms of types, fields, and operations
- No `identity.Key` — just type names and field names
- Input to language-specific code generators
- Think of this as "typed pseudocode in tree form"

The lowering from Level 1 → Level 2 resolves model references to concrete types:
```
Level 1:  AttributeRef(identity.Key("domain/d/subdomain/s/class/order/attribute/count"))
Level 2:  FieldAccess(TypeRef("Order"), "Count")    // using generated type/field names
```

This two-level split keeps the model expression close to the domain (good for validation, simulation, and understanding) while the code gen logic is close to implementation (good for generating clean code).

**Build Level 1 first.** Level 2 emerges naturally when code gen starts.

---

## Parser AI Integration

### AI Writes Logic via IR, Not TLA+

The parser_ai package currently accepts Logic specifications as TLA+ strings in JSON. With the IR, parser_ai gains a second (and preferred) input mode: AI writes logic as structured **expression trees** in JSON, and the system generates the TLA+ specification automatically.

This is significant because:
- AI does not need to learn TLA+ syntax to author model logic.
- The JSON schema for expression nodes can be richly descriptive, teaching the AI what valid expressions look like, what operators are available, and how to reference model entities.
- The expression structure is validated structurally (JSON schema) and semantically (model-aware validation). Error codes and `.md` error files give the AI precise course-correction instructions.
- TLA+ becomes a fixed internal notation — a display/storage format that is always generated, never hand-authored by AI.

### Flow

```
AI JSON input (inputExpression)
      │
      ▼
parser_ai reads expression ──► model_expression.Expression
      │
      ▼
generate TLA+ from expression ──► Logic.Specification (deterministic TLA+ text)
      │
      ▼
model_logic.NewLogic(...) with both Expression and generated Specification
      │
      ▼
rest of system sees a complete Logic with validated expression + matching TLA+
```

### `inputExpression` Type

Following the parser_ai naming convention (`inputAction`, `inputQuery`, `inputLogic`, etc.), the expression tree is exposed as `inputExpression` — a JSON-serializable representation of a `model_expression.Expression` node.

The `inputLogic` struct evolves to:

```go
type inputLogic struct {
    Type          string           `json:"type,omitempty"`
    Description   string           `json:"description"`
    Target        string           `json:"target,omitempty"`
    Notation      string           `json:"notation,omitempty"`
    Specification string           `json:"specification,omitempty"`  // Still accepted (human-authored TLA+)
    Expression    *inputExpression `json:"expression,omitempty"`     // NEW: structured expression tree
}
```

**Exactly one** of `specification` or `expression` may be provided, never both. The parser_ai validation enforces this with an error code and descriptive `.md` file.

### JSON Schema for Expressions

The expression JSON schema (`expression.schema.json`) must be **exceptionally descriptive** — it is the primary documentation for AI authoring logic. Each node type needs:

- A clear description of what the node represents mathematically
- Examples of when to use it (e.g., "Use `attribute_ref` to reference a class attribute like `self.count`")
- Enum values with descriptions for operators (e.g., `"add"` — "Arithmetic addition of two numeric values")
- Guidance on which fields are required for each node type
- Descriptions of how `identity.Key` references work (e.g., "The attribute_key must be a valid attribute key like `domain/d/subdomain/s/class/c/attribute/a`")

This is consistent with the existing parser_ai convention: "Schema descriptions should teach an AI how to correctly fill out the data."

### Error Handling for Expressions

Expression validation errors get their own error code range in parser_ai (e.g., `17xxx`), each with an `.md` file explaining:
- What went wrong (e.g., "attribute_key references a non-existent attribute")
- How to fix it (e.g., "Check the class attributes and use one of: [list of valid attribute keys]")
- Examples of correct usage

This matches the existing parser_ai pattern where no `req_model` error messages leak through — parser_ai reports its own errors with distinct error codes.

---

## Logic Constructor: Dual-Input Mode

### Design

`model_logic.NewLogic` accepts either a specification string or an expression, never both:

```go
func NewLogic(key identity.Key, logicType, description, target, notation, specification string, expression *model_expression.Expression) (Logic, error)
```

**Construction rules:**

| Input | Behavior |
|-------|----------|
| Spec only (expression nil) | Parse spec → attempt lowering → if successful, populate Expression; if parse fails, Expression remains nil (valid — human may be editing) |
| Expression only (spec empty) | Generate TLA+ from expression → populate Specification |
| Both provided | Validation error — ambiguous source of truth |
| Neither provided | Valid — Logic with no formal specification (description-only) |

### Why Allow Spec Without Expression?

A human editing TLA+ directly may have work-in-progress syntax errors. The system should not reject the entire Logic — it stores the spec text and tries again later (or on next load). This is a graceful degradation: the Logic is valid but "uncompiled." Semantic validation and code generation require the expression, but storage and display work fine with just the spec.

### Why Generate TLA+ from Expression?

The `Specification` field serves two purposes:
1. **Human readability** — developers and reviewers can read the TLA+ in docs and reports
2. **Round-trip stability** — the parser_ai round-trip test (WriteModel → ReadModel → Compare) needs the TLA+ text to match exactly

By generating TLA+ deterministically from the expression, the round-trip is always clean. There is no "which TLA+ formatting did the human use?" ambiguity — the generated TLA+ is canonical.

### IR-to-TLA+ Generation

A new package under `notation/tla/` handles expression → TLA+ text:

```
notation/
├── tla/
│   ├── ast/         # TLA+ AST
│   ├── parser/      # TLA+ text → TLA+ AST
│   ├── lowering/    # TLA+ AST → model expression
│   └── raising/     # model expression → TLA+ text (NEW)
```

The raising pass is straightforward — the model expression nodes have a nearly 1:1 mapping to TLA+ syntax. The generated TLA+ does not need to match any human-authored input; it just needs to be semantically correct and parseable.

---

## Test Model Strategy

### Expression-First Construction

Both `GetTestModel()` and `GetStrictTestModel()` in `test_helper` currently build all 31 Logic objects with TLA+ specification strings. After the IR is implemented, these must be changed to build Logic objects from **expressions** instead.

**Why:** The parser_ai round-trip test writes the model to JSON and reads it back. If the test model's Logic objects are built from spec strings, the round-trip must preserve the exact TLA+ formatting — any whitespace or formatting difference causes a mismatch. If instead the test model builds from expressions, the TLA+ is generated deterministically both times (initial construction and post-round-trip reconstruction), guaranteeing an exact match.

**Migration:**

```go
// Before (spec-first):
helper.Must(model_logic.NewLogic(key, "assessment", "Order exists", "", "tla_plus", "order \\in Orders"))

// After (expression-first):
helper.Must(model_logic.NewLogic(key, "assessment", "Order exists", "", "tla_plus", "",
    &model_expression.Membership{
        Element: &model_expression.LocalVar{Name: "order"},
        Set:     &model_expression.AttributeRef{AttributeKey: ordersKey},
    },
))
```

The generated `Specification` will be `"order ∈ Orders"` (or equivalent canonical TLA+), and the round-trip reproduces this exactly.

### Impact on Existing Tests

- **parser_ai round-trip test**: Works better — no TLA+ formatting mismatches possible.
- **parser (human format) tests**: Uses `GetTestModel()`, which will now have generated TLA+ specs. The human parser reads specs as strings, so this is transparent — it just sees a different (canonical) TLA+ string.
- **simulator tests**: Already use constructors with `helper.Must()`. The Logic objects they build will also shift to expression-first. Since the simulator evaluator will accept model expressions directly, this is natural.
- **database tests**: Exempt from constructor rules. No impact.

---

## Migration Path

1. **Restructure notation** — move `notation/ast/` → `notation/tla/ast/`, move `simulator/parser/` → `notation/tla/parser/`
2. **Define model expression** (`req_model/model_expression/`) — node types, operator enums, validation
3. **Write TLA+ lowering** (`notation/tla/lowering/`) — `ast.Expression → model_expression.Expression`, requires model context for name resolution
4. **Write TLA+ raising** (`notation/tla/raising/`) — `model_expression.Expression → TLA+ text` (canonical generation)
5. **Update Logic constructor** — dual-input mode (spec-only or expression-only), with parse/lower for specs and raise/generate for expressions
6. **Port the evaluator** — change `evaluator.Eval` to accept `model_expression.Expression`
7. **Port the type checker** — same change
8. **Update model_bridge** — extract strings → parse → lower → register expression
9. **Add database table** — `expression_node` table with adjacency list and foreign keys
10. **Update database layer** — load/save expression trees
11. **Update parser_ai** — add `inputExpression` type, `expression.schema.json`, error codes for expression validation, convert expression to/from model
12. **Update test models** — convert all 31 Logic objects in `test_helper` from spec-first to expression-first construction
13. **Write code generator** (future) — consumes model expressions, produces code

Steps 1-2 can be done first as a pure refactoring. Steps 3-5 give the Logic constructor both directions (spec→expression and expression→spec). Steps 6-8 port the runtime. Steps 9-10 add persistence. Steps 11-12 complete the parser_ai integration and test infrastructure.

---

## `model_logic.Logic` Evolution

```go
type Logic struct {
    Key           identity.Key
    Type          string               // assessment, state_change, safety_rule, value
    Description   string
    Target        identity.Key         // Now a key, not a string. FK to attribute. Empty for non-assignment types.
    Notation      string               // "tla_plus" — which notation was used to author this
    Specification string               // TLA+ text — generated from Expression, or human-authored (may have parse errors)
    Expression    *model_expression.Expression  // The structured expression tree — source of truth when present
}
```

**Dual construction modes:**

- **Expression provided, no spec**: The constructor generates canonical TLA+ and populates `Specification`. This is the preferred mode for AI and programmatic construction. The `Expression` is the source of truth.
- **Spec provided, no expression**: The constructor parses the TLA+ and attempts lowering to populate `Expression`. If parsing fails (human WIP), `Expression` stays nil — the Logic is valid but "uncompiled." The `Specification` is the source of truth until parsing succeeds.
- **Both provided**: Validation error.
- **Neither provided**: Valid — description-only logic with no formal specification.

The `Target` becomes an `identity.Key` backed by a foreign key. The `Notation` field records which notation was used for authoring (always `"tla_plus"` for now) and determines which raising pass generates the specification text.

---

## Sources

- [Intermediate Representation — Wikipedia](https://en.wikipedia.org/wiki/Intermediate_representation)
- [Intermediate Representation — Communications of the ACM](https://cacm.acm.org/practice/intermediate-representation/)
- [Intermediate Representation — ACM Queue](https://queue.acm.org/detail.cfm?id=2544374)
- [Why Do We Use Intermediate Representations? — Musing Mortoray](https://mortoray.com/why-we-use-intermediate-representations/)
- [Design Guidelines for Domain Specific Languages — SE-RWTH](https://www.se-rwth.de/publications/Design-Guidelines-for-Domain-Specific-Languages.pdf)
- [When and How to Develop Domain-Specific Languages — Computing Surveys](https://inkytonik.github.io/assets/papers/compsurv05.pdf)
- [The Complete Guide to Domain Specific Languages — Strumenta](https://tomassetti.me/domain-specific-languages/)

-----------------------------

The general steps to follow are these to minimize rewrites...

-----------------------------

1. The Model (do first)

The heart of apps/requirements/req is the package apps/requirements/req/internal/req_model. Any code to be made that would be in that file tree should be done first. It is the source-of-truth for the data model in the system. Work confirmed with `go test ./internal/req_model/...`

A defined model is the full object tree that is passed around the system to be be used in different ways.

Then the test model needs to be updated to work. It is the input to the tests in other packages. Work confirmed with `go test ./internal/test_helper/...`

(All tests in this document run from `cd /workspaces/glemzurg/apps/requirements/req`)

-----------------------------

2. The Database (do second)

The database is the first grindy confirmation of the model. It forces the exactness of hte model to be pushed into a relational SQL shape completely. 

The apps/requirements/req/internal/database/sql/schema.sql is the schema, and any change to it should have comments that fit the pattern of the rest of the file. If there is an enum in the req_model, then it should be enumeration type in the schema, just like the rest of the model shows.

The database layer code should match the pattern of the rest of the data access calls and then there is a top-level round trip test to confirm the test model works. The database must be tested with a flag: `go test ./internal/database/... -dbtests`

There are some specific design choices in the database code:

- The INSERT in teh load never uses string placehodlers as parameter, only literally written strings.
- Each table has its own golang code file and unit test. The golang code files should not call other table's files. It will just be the top-level test and code that stitch values together from the dabase into objects.

After database work is done the documenation should be regenerated.

- Remove all the files in apps/requirements/req/docs/dbdoc
- Run apps/requirements/req/doc.sh

-----------------------------

3. The AI Parser (do third)

The AI parser (apps/requirements/req/internal/parser_ai) has more strict data requirements thatn teh rest of the system, so it uses the strict test model instead of the test mode. If the struct test model needs more objects defined to pass the ai parser round-trip then it should be updated to do so. Work confirmed with `go test ./internal/parser_ai/...`

The AI parser has a few design choices:

- All data from the req_model needs to be in the objects that are written to or read from json. This is enforced by the round trip test working.
- No error message from req_model should ever be eported by the parser_ai code. Instead parser_ai should find that error and report it itself, ensuring it has a distinct error code and a error md that instructs a calling ai how to correct the error.
- As much validation information should be put into the json schemas as possible, since it is a clean form of documentation for AI tooling.
- Every bit of description in the schemas should be helping instruct and teach and ai how to correctly fill out the data.

Before considering the parser_ai complete for a task, do an examination that every possible error that can be reported under a req_model class will be reported with an appropriate error code and md file from parser_ai.

-----------------------------

4. The human parser (do fourth)

The human parser (apps/requirements/req/internal/parser) is a custom markdown yaml data format. Ensure all the data that goes in and out is explored in the tests for a class. And then the round-trip test will confirm that everything works well together. It uses the test_helper.GetTestModel(), not the strict model, but that model should *not* be updated to fix a bug in the parser project. That model is meant to explore the constraints and the flexibility of the req_model tree. 

-----------------------------

5. The flattened requirements (do fifth)

The flattened requirements (apps/requirements/req/internal/req_flat) prepares lookups for the template generation. Add any new looksups needed. Confirm work with `go test ./internal/req_flat/...`

-----------------------------

6. The generation code (do sixth)

The generation code (apps/requirements/req/internal/generate) should be updated to include any new objects or data, fitting into the patterns that already exist. New lookups will likely need to be added to apps/requirements/req/internal/generate/template.go. The markdown can be generated for testing from apps/requirements/req/internal/generate/dump_test_model_test.go (if the Skip is temporarily disabled). Work here should be confirmed with `go test ./internal/generate/...`

-----------------------------

7. The simulator (do seventh)

The simulator (apps/requirements/req/internal/simulator) should be updated and confirmed with testing: `go test ./internal/simulator/...`


-----------------------------

Lastly, do a final check: `go test ./...`
