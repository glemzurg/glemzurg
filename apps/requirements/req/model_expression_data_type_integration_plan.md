# Implementation Plan: model_expression and model_data_type Integration

This plan implements the design in `model_expression_data_type_integration.md`. Each stage is a code-review-iterate cycle scoped to a specific part of the source tree. Stages will break downstream packages — that breakage is expected and addressed in later stages.

---

## Current Status (as of 2026-03-05)

| Stage | Status | Notes |
|-------|--------|-------|
| **1A** ExpressionSpec/TypeSpec | **DONE** | `model_spec/` package with constructors, parse func support |
| **1B** model_expression_type | **DONE** | Full type hierarchy with validation |
| **1C** Logic → ExpressionSpec | **DONE** | `Spec ExpressionSpec` + `TargetTypeSpec *TypeSpec` on Logic |
| **1D** TypeSpec on DataType | **DONE** | `TypeSpec *model_spec.TypeSpec` on DataType |
| **1E** NamedSet entity | **DONE** | `model_named_set/` package + `NamedSets` field on Model |
| **1F** NamedSetRef node | **DONE** | `NamedSetRef` expression node in `model_expression` |
| **1G** Attribute Invariants | **DONE** | `Invariants` field + `SetInvariants()` on Attribute |
| **1G-ii** Test model invariants | **DONE** | 3 sets of attribute invariants with let-type logic |
| **1H** Simulator alignment | **DONE** | Uses `logic.Spec.Expression` throughout |
| **1I** Test helper updates | **DONE** | Uses ExpressionSpec constructors, includes lowering |
| **1J** Stage 1 verified | **DONE** | All req_model + test_helper + simulator tests pass |
| **Stage 2** Notation parsing | **DONE** | `ConvertToExpressionType` + `RaiseType` with tests |
| **Stage 3** parser_human | **DONE** | Named sets, attribute invariants, let type, target_type_spec, lowering |
| **Stage 4** parser_ai | **DONE** | Full round-trip support: named sets, attribute invariants, let type, target_type_spec, schemas updated |
| **Stage 5** Database | **DONE** | Schema updated, named_set/attribute_invariant tables, let enum, expression lowering in ReadModel |
| **Stage 6** req_flat/generate | **DONE** | All tests pass |
| **Stage 7** Full integration | **DONE** | `go test ./...` passes |

**Additional work done beyond original plan:**
- `LogicTypeLet` ("let") — local variable definitions in requires/guarantees/safety_rules/invariants. Fully supported across all layers.
- Expression lowering consolidated into `convert.LowerAllExpressions()` — called by parser_human, parser_ai, test_helper, and database ReadModel. Eliminated three duplicate `lower_expressions.go` files.
- `ExpressionParseFunc` on `ExpressionSpec`/`TypeSpec` constructors — enables deferred parsing with context.
- Primed value documentation in schemas — safety_rules are the only context where primed values are allowed.
- **Design change:** Parsed expression trees (intermediate representation) are NOT stored in the database. The `expression_node` table was removed from the design. Instead, `ReadModel` calls `convert.LowerAllExpressions()` after tree assembly to parse all TLA+ specifications with full context.

**All stages complete.** The plan is fully implemented.

---

## Stage 1: req_model Tree — Data Structures

**Goal:** Get the model types right. This is the source of truth — everything downstream follows.

**Packages touched:** `internal/req_model/` tree + `internal/test_helper/` + `internal/simulator/` (alignment only)

**Test command:** `go test ./internal/req_model/... && go test ./internal/test_helper/... && go test ./internal/simulator/...`

### 1A: ExpressionSpec and TypeSpec Value Objects

Create reusable formal specification value objects. These carry Notation + Specification + parsed tree as a single concept.

**Decision needed:** These are used across packages (`model_logic`, `model_data_type`, `model_expression_type`). They either live in a shared package or are duplicated. Given that `model_logic` imports `model_expression` and `model_data_type` imports `model_expression_type`, a small shared package avoids circular imports.

**New package: `internal/req_model/model_spec/`**

- `expression_spec.go` — `ExpressionSpec` struct: `Notation string`, `Specification string`, `Expression model_expression.Expression`. Constructor `NewExpressionSpec(...)`, `Validate()`. Notation is validated as `oneof=tla_plus`. Specification is optional (empty string = not yet written). Expression is optional (nil = not yet parsed).
- `type_spec.go` — `TypeSpec` struct: `Notation string`, `Specification string`, `ExpressionType model_expression_type.ExpressionType`. Constructor `NewTypeSpec(...)`, `Validate()`. Same validation pattern.
- `validate.go` — shared validator instance.
- `expression_spec_test.go` — table-driven Validate tests for ExpressionSpec.
- `type_spec_test.go` — table-driven Validate tests for TypeSpec.

**Note:** `TypeSpec` imports `model_expression_type` which doesn't exist yet. Create `model_expression_type` first (step 1B), then `model_spec`, then update consumers.

### 1B: model_expression_type Package

Create the precise structural type system.

**New package: `internal/req_model/model_expression_type/`**

- `expression_type.go` — `ExpressionType` interface with `expressionType()` marker, `TypeName() string`, `Validate() error`. Type name string constants for each concrete type.
- `types.go` — All concrete type structs:
  - `BooleanType{}` — no fields
  - `IntegerType{}` — no fields
  - `RationalType{}` — no fields
  - `StringType{}` — no fields
  - `EnumType{Values []string}` — validate: required, min=1
  - `SetType{ElementType ExpressionType}` — validate: ElementType required
  - `SequenceType{ElementType ExpressionType, Unique bool}` — validate: ElementType required
  - `BagType{ElementType ExpressionType}` — validate: ElementType required
  - `TupleType{ElementTypes []ExpressionType}` — validate: required, min=1
  - `RecordType{Fields []RecordFieldType}` — validate: required, min=1
  - `FunctionType{Params []ExpressionType, Return ExpressionType}` — validate: Return required
  - `ObjectType{ClassKey identity.Key}` — validate: ClassKey required
  - `NamedSetRef{SetKey identity.Key}` — validate: SetKey required
- `record_field_type.go` — `RecordFieldType{Name string, Type ExpressionType}`.
- `validate.go` — shared validator instance.
- `expression_type_test.go` — table-driven Validate tests for each type (valid + error cases).

Each concrete type implements `expressionType()`, `TypeName()`, `Validate()`. Validate recurses into children.

### 1C: Refactor Logic to Use ExpressionSpec

Update `model_logic.Logic` to embed `ExpressionSpec` instead of holding Notation, Specification, and Expression as separate fields. Add TargetTypeSpec.

**Modified file: `internal/req_model/model_logic/logic.go`**

Current:
```go
type Logic struct {
    Key           identity.Key
    Type          string
    Description   string
    Target        string
    Notation      string
    Specification string
    Expression    model_expression.Expression
}
```

New:
```go
type Logic struct {
    Key            identity.Key
    Type           string
    Description    string
    Target         string
    Spec           model_spec.ExpressionSpec    // Notation + Specification + Expression
    TargetTypeSpec *model_spec.TypeSpec          // Optional: declared result type
}
```

**Constructor change:** `NewLogic(key, logicType, description, target string, spec model_spec.ExpressionSpec, targetTypeSpec *model_spec.TypeSpec) (Logic, error)`

**Blast radius (within req_model + test_helper + simulator):**
- `model_logic/logic_test.go` — update test cases
- `model_logic/global_function.go` — constructor takes Logic, check field access
- `model_logic/global_function_test.go` — update
- `test_helper/test_model.go` — ~31 NewLogic calls, all need updated signature
- All simulator files that create Logic objects (test files) — pass `model_spec.ExpressionSpec{Notation: "tla_plus", ...}` and `nil` for TargetTypeSpec

**Important:** The `Notation`, `Specification`, `Expression` fields move inside `Spec`. Every caller that accessed `logic.Notation` now accesses `logic.Spec.Notation`. Every caller that accessed `logic.Expression` now accesses `logic.Spec.Expression`.

### 1D: Add TypeSpec to DataType

Add an optional `TypeSpec` field to `model_data_type.DataType`.

**Modified file: `internal/req_model/model_data_type/data_type.go`**

New field:
```go
type DataType struct {
    // ...existing fields...
    TypeSpec *model_spec.TypeSpec  // Optional precise type specification
}
```

**Constructor change:** `New(key, text string, typeSpec *model_spec.TypeSpec) (*DataType, error)` — adds TypeSpec parameter. All existing callers pass `nil`.

**Validate change:** If TypeSpec is non-nil, call `TypeSpec.Validate()`.

**Blast radius:**
- `data_type_test.go` — update constructor calls
- `model_class/attribute.go` — `NewAttribute` calls `model_data_type.New()`, add `nil` for TypeSpec
- `model_state/parameter.go` — `NewParameter` calls `model_data_type.New()`, add `nil` for TypeSpec
- `test_helper/test_model.go` — all DataType constructions pass `nil`
- `parser/class.go` — will break (fixed in Stage 3)
- `parser_ai/convert_to_model.go` — will break (fixed in Stage 4)
- `database/` — will break (fixed in Stage 5)

### 1E: Add NamedSet Model Entity

Create the named set entity and add it to the Model.

**New package: `internal/req_model/model_named_set/`**

- `named_set.go` — `NamedSet` struct with Key (`identity.Key`), Name, Description, Spec (`ExpressionSpec`), TypeSpec (`*TypeSpec`). Constructor `NewNamedSet(...)`, `Validate()`, `ValidateWithParent()`.
- `validate.go` — shared validator.
- `named_set_test.go` — table-driven tests.

**New key type: `internal/identity/key_type.go`**

Add `KEY_TYPE_NAMED_SET = "nset"` with constructor `NewNamedSetKey(subKey string)`.

**Modified file: `internal/req_model/model.go`**

Add field:
```go
type Model struct {
    // ...existing fields...
    NamedSets map[identity.Key]model_named_set.NamedSet
}
```

Update constructor: `NewModel(key, name, details string, invariants []model_logic.Logic, globalFunctions map[identity.Key]model_logic.GlobalFunction, namedSets map[identity.Key]model_named_set.NamedSet) (Model, error)`

Update `Model.Validate()` to validate NamedSets map.

**Blast radius:**
- `model_test.go` — update constructor calls
- `test_helper/test_model.go` — add `nil` for namedSets initially
- All callers of `NewModel()` need the new parameter

### 1F: Add NamedSetRef Expression Node

Add a new expression node for referencing named sets in behavioral logic.

**Modified file: `internal/req_model/model_expression/nodes.go`**

```go
// NamedSetRef references a model-level named set by key.
type NamedSetRef struct {
    SetKey identity.Key `validate:"required"`
}

func (n *NamedSetRef) expressionNode()    {}
func (n *NamedSetRef) NodeType() string   { return NodeNamedSetRef }
```

**Modified file: `internal/req_model/model_expression/expression.go`**

Add constant: `NodeNamedSetRef = "named_set_ref"`

**Modified file: `internal/req_model/model_expression/validate.go`**

Add Validate() for NamedSetRef.

### 1G: Add Attribute Invariants

Attributes currently have no invariant support. Add optional invariants to Attribute, following the same pattern as Class.Invariants and Model.Invariants.

**New key type: `internal/identity/key_type.go`**

Add `KEY_TYPE_ATTRIBUTE_INVARIANT = "ainvariant"` with constructor `NewAttributeInvariantKey(attributeKey Key, subKey string)`.

**Modified file: `internal/req_model/model_class/attribute.go`**

Add field:
```go
type Attribute struct {
    // ...existing fields...
    Invariants []model_logic.Logic   // NEW: attribute-level invariants
}
```

Attribute invariants are `LogicTypeAssessment` — boolean predicates that must hold for the attribute's value. In TLA+ specifications, the binding `attribute` refers to the attribute's current value (analogous to how class invariants use `self.` for the instance).

**Constructor change:** Add `invariants []model_logic.Logic` parameter to `NewAttribute()`.

**Validation:**
- Each invariant validated with `inv.ValidateWithParent(&a.Key)`
- Each must be `LogicTypeAssessment`

Add a `SetInvariants()` method on `Attribute`, following the pattern of `Class.SetInvariants()`:
```go
func (a *Attribute) SetInvariants(invariants []model_logic.Logic) {
    a.Invariants = invariants
}
```

This allows the constructor to accept `nil` initially and invariants to be set after construction (same pattern as `Class`).

**Modified file: `internal/req_model/model_class/attribute_test.go`**

Add test cases:
- Attribute with nil invariants — valid
- Attribute with valid assessment invariants — valid
- Attribute with wrong logic type (e.g., state_change) — error
- Attribute invariant with wrong parent key — error
- Attribute with multiple invariants — valid

**Blast radius:**
- `attribute_test.go` — update test cases
- `test_helper/test_model.go` — all `NewAttribute()` calls pass `nil` for invariants initially; then real invariants added (see 1G-ii)
- `parser/class.go` — will break (fixed in Stage 3)
- `parser_ai/class.go` — will break (fixed in Stage 4)
- `database/attribute.go` — will break (fixed in Stage 5)

### 1G-ii: Add Attribute Invariants to Test Models

Following the same pattern used for class invariants and model invariants, add real attribute invariant Logic objects to the test models.

**Modified file: `internal/identity/test_keys.go`** (or wherever test keys are built)

Add attribute invariant keys:
```go
k.attrInv1, err = identity.NewAttributeInvariantKey(k.orderTotal, "0")    // Order.total >= 0
k.attrInv2, err = identity.NewAttributeInvariantKey(k.orderTotal, "1")    // Order.total <= max
k.attrInv3, err = identity.NewAttributeInvariantKey(k.productPrice, "0")  // Product.price > 0
```

**Modified file: `internal/test_helper/test_model.go`**

In the logic-building section (near where class invariants are built, ~line 1083):
1. Create attribute invariant keys in `buildKeys()`
2. Create attribute invariant Logic objects in `buildLogic()`:
   ```go
   attrInv1, err := model_logic.NewLogic(
       k.attrInv1,
       model_logic.LogicTypeAssessment,
       "Order total must be non-negative",
       "",
       spec,  // ExpressionSpec with "attribute >= 0"
       nil,
   )
   ```
3. Store in `testLogic` struct:
   ```go
   attrInvariantsOrderTotal []model_logic.Logic  // Order.total (2)
   attrInvariantsProductPrice []model_logic.Logic // Product.price (1)
   ```
4. Set on attributes using `SetInvariants()`:
   ```go
   attrOrderTotal.SetInvariants(l.attrInvariantsOrderTotal)
   attrProductPrice.SetInvariants(l.attrInvariantsProductPrice)
   ```

This ensures the test models exercise attribute invariants end-to-end (test_helper → database → parser_ai round-trip).

### 1H: Simulator Alignment

Update simulator files to compile with the new Logic and Attribute structures. This is mechanical — change field access patterns but no behavioral changes.

**Key changes in simulator:**
- `model_bridge/extractor.go` — change `logic.Specification` to `logic.Spec.Specification`, `logic.Notation` to `logic.Spec.Notation`, `logic.Expression` to `logic.Spec.Expression`
- `model_bridge/` test files — update Logic construction
- `invariants/data_type_checker.go` — no change yet (accesses `attr.DataType` which still exists). Attribute invariant evaluation is future work — the `DataTypeChecker` would extend to bind `attribute` and evaluate each invariant Logic.
- `typechecker/` — no change (doesn't access Logic fields directly)
- `evaluator/` — no change
- `registry/` — update if it accesses Logic fields
- Test files throughout simulator that construct Logic or Attribute objects — update to new constructor signatures

**No behavioral changes to the simulator.** TypeSpec, TargetTypeSpec, NamedSet, attribute invariants are all nil/empty/unused at this point. The simulator continues to work exactly as before.

### 1I: Test Helper Updates

Update `test_helper/test_model.go` to construct model objects with new signatures.

- All `NewLogic()` calls: wrap Notation/Specification/Expression into `model_spec.ExpressionSpec{}`, pass `nil` for TargetTypeSpec
- All `NewModel()` calls: pass `nil` for namedSets
- All `NewAttribute()` calls: pass `nil` for invariants (real invariants are set via `SetInvariants()` — see step 1G-ii)
- All `model_data_type.New()` calls (if any direct): pass `nil` for TypeSpec
- Add attribute invariant keys, Logic objects, and `SetInvariants()` calls as described in step 1G-ii

### 1J: Verify Stage 1

```bash
cd /workspaces/glemzurg/apps/requirements/req
go test ./internal/req_model/...
go test ./internal/test_helper/...
go test ./internal/simulator/...
```

All three must pass. The `parser`, `parser_ai`, `database`, `req_flat`, and `generate` trees will have compile errors — that is expected.

---

## Stage 2: notation Tree — Parsing

**Goal:** Enable parsing of TLA+ type expressions into ExpressionType trees.

**Packages touched:** `internal/notation/`

**Test command:** `go test ./internal/notation/...`

### 2A: Assess Grammar Needs

The TLA+ expression parser already handles most type expression constructs as regular expressions:
- `Seq(Int)` → parsed as `FunctionCall` with ScopePath
- `_Set(STRING)` → parsed as `FunctionCall` with underscore module prefix
- `SUBSET S` → may or may not exist in grammar
- `{"a", "b", "c"}` → parsed as `SetLiteral`
- `Nat`, `Int`, `Real`, `BOOLEAN`, `STRING` → parsed as `Identifier` or `SetConstant`

**What needs grammar work:**
- `[name: STRING, age: Int]` — record TYPE syntax with colon (currently only `[name |-> value]` record VALUE syntax exists). The parser needs to distinguish `[f: T]` (type) from `[f |-> v]` (value). These are distinct TLA+ constructs.
- `Nat \X STRING` — Cartesian product. Check if `\X` or `\times` is in the grammar.

### 2B: Extend PEG Grammar (if needed)

**Modified file: `internal/notation/tla_plus/parser/peg/tla_expression.peg`**

- Add record type syntax rule: `RecordTypeExpr <- '[' FieldTypeBinding (',' FieldTypeBinding)* ']'` where `FieldTypeBinding <- Identifier ':' Expression`. This is distinct from `RecordInstance` which uses `|->`.
- Add Cartesian product if not present: `CartesianExpr <- Expression '\X' Expression` (or `\times`).
- Regenerate: run pigeon to produce `tla_parser.generated.go`.

### 2C: AST Node for Record Type (if new)

If record type syntax produces a new AST node:

**New file or modified: `internal/notation/tla_plus/ast/record_type.go`**

```go
type RecordType struct {
    Fields []RecordTypeField
}
type RecordTypeField struct {
    Name *Identifier
    Type Expression
}
```f

Implements `Expression` interface.

### 2D: Type Expression Converter

Create a function that converts a TLA+ AST (from the parser) into an ExpressionType tree. This is the "type interpretation" pass.

**New file: `internal/notation/tla_plus/ast/type_convert.go`** (or a new sub-package)

```go
func ConvertToExpressionType(expr ast.Expression) (model_expression_type.ExpressionType, error)
```

Mapping:
| AST Node | ExpressionType |
|---|---|
| `SetConstant("BOOLEAN")` | `BooleanType{}` |
| `SetConstant("NAT")` / `SetConstant("INT")` | `IntegerType{}` |
| `SetConstant("REAL")` | `RationalType{}` |
| `Identifier("STRING")` | `StringType{}` |
| `SetLiteral{StringLiterals...}` | `EnumType{Values: [...]}` |
| `FunctionCall{ScopePath:["_Seq"], Name:"Seq", Args:[X]}` | `SequenceType{ElementType: convert(X), Unique: false}` |
| `FunctionCall{ScopePath:["_Seq"], Name:"_SeqUnique", Args:[X]}` | `SequenceType{ElementType: convert(X), Unique: true}` |
| `FunctionCall{ScopePath:["_Set"], Name:"_Set", Args:[X]}` | `SetType{ElementType: convert(X)}` |
| `FunctionCall{...Stack/Queue variants}` | `SequenceType{...}` with appropriate Unique |
| `FunctionCall{ScopePath:["_Bags"], Name:"_Bag", Args:[X]}` | `BagType{ElementType: convert(X)}` |
| `RecordType{Fields:[...]}` | `RecordType{Fields: [...]}` |
| `CartesianProduct{...}` or `\X` | `TupleType{ElementTypes: [...]}` |
| `Identifier(name)` where name is a known NamedSet | `NamedSetRef{SetKey: ...}` |
| Anything else | error: "not a valid type expression" |

**Test file: `internal/notation/tla_plus/ast/type_convert_test.go`**

Table-driven tests: parse TLA+ string → convert → assert ExpressionType matches expected.

### 2E: Verify Stage 2

```bash
cd /workspaces/glemzurg/apps/requirements/req
go test ./internal/notation/...
```

---

## Stage 3: parser Tree — Human Input Flow

**Goal:** Enable the human parser to read `precise_type:` from YAML and populate `DataType.TypeSpec`.

**Packages touched:** `internal/parser_human/`

**Test command:** `go test ./internal/parser_human/...`

### 3A: Fix Compile Errors from Stage 1

The human parser calls `NewLogic()`, `NewModel()`, and `model_data_type.New()` — all signatures changed. Fix these first.

**Modified file: `internal/parser_human/class.go`**
- All `NewLogic()` calls: wrap Notation/Specification/nil into `model_spec.ExpressionSpec{Notation: "tla_plus", Specification: spec}`, pass `nil` for TargetTypeSpec
- All `model_data_type.New()` calls: pass `nil` for TypeSpec (initially — step 3B adds real parsing)
- All struct literal Logic construction: update to use new fields (`Spec: model_spec.ExpressionSpec{...}`)

**Modified file: `internal/parser_human/model.go`** (or wherever NewModel is called)
- Add `nil` for namedSets parameter

### 3B: Parse `precise_type:` Field on Attributes

**Modified file: `internal/parser_human/class.go`**

In `attributeFromYamlData()`:
- Read `"precise_type"` from YAML map (optional field)
- If present, it's a TLA+ string. Store it as the DataType's TypeSpec:
  ```go
  if preciseTypeStr != "" {
      ts, err := model_spec.NewTypeSpec("tla_plus", preciseTypeStr, nil)
      // Pass ts to DataType construction
  }
  ```
- Parse the TLA+ string into an ExpressionType (using Stage 2's converter):
  ```go
  ast, err := parser.ParseExpression(preciseTypeStr)
  exprType, err := notation_ast.ConvertToExpressionType(ast)
  ts.ExpressionType = exprType
  ```
- Handle parse errors gracefully (same pattern as DataTypeRules — store string, nil ExpressionType on failure)

### 3C: Parse `precise_type:` Field on Parameters

Same pattern for parameters in actions, queries, events. Parameters read `"rules"` for DataTypeRules — they would also read `"precise_type"` if present.

### 3D: Parse Named Sets

**Modified file: `internal/parser_human/model.go`** (or new `internal/parser_human/named_set.go`)

Read `"named_sets"` from model-level YAML:
```yaml
named_sets:
  IsoStateAbbr:
    description: ...
    specification: '{"AL", "AK", ...}'
```

Parse each into `model_named_set.NamedSet`:
- Name from the YAML key
- Description from `"description"`
- Specification from `"specification"`
- Parse the specification TLA+ string into Expression tree
- Optionally derive TypeSpec from the Expression

### 3E: Parse Attribute Invariants

**Modified file: `internal/parser_human/class.go`**

In `attributeFromYamlData()`:
- Read `"invariants"` from YAML map (optional field, list of objects)
- Each invariant has `description` and `specification` fields
- Parse each into a `model_logic.Logic` with type `LogicTypeAssessment`:
  ```go
  invariantKey := identity.NewAttributeInvariantKey(attributeKey, subKey)
  spec := model_spec.ExpressionSpec{Notation: "tla_plus", Specification: invSpec}
  logic, err := model_logic.NewLogic(invariantKey, model_logic.LogicTypeAssessment, description, "", spec, nil)
  ```
- Pass the resulting `[]model_logic.Logic` to `NewAttribute()`

Example YAML:
```yaml
attributes:
  balance:
    name: Account Balance
    rules: [0 .. 1000000] at 0.01 dollar
    precise_type: Nat
    invariants:
      - description: Balance must be non-negative
        specification: "attribute >= 0"
```

### 3F: Write-Back

When the human parser writes model back to files, TypeSpec, NamedSet, and attribute invariants need to be serialized. The `precise_type:` field is written to attribute YAML. Named sets are written to model-level YAML. Attribute invariants are written under `invariants:` in attribute YAML.

### 3G: Verify Stage 3

```bash
cd /workspaces/glemzurg/apps/requirements/req
go test ./internal/parser_human/...
```

---

## Stage 4: parser_ai Tree — AI Input Flow

**Goal:** Enable the AI parser to accept and produce ExpressionType and NamedSet data.

**Packages touched:** `internal/parser_ai/`

**Test command:** `go test ./internal/parser_ai/...`

### 4A: Fix Compile Errors from Stage 1

Same pattern as Stage 3 — update all `NewLogic()`, `NewModel()`, `model_data_type.New()` calls.

**Modified files:**
- `convert_to_model.go` — update Logic construction to use ExpressionSpec, add nil for TypeSpec/TargetTypeSpec/NamedSets
- `convert_from_model.go` — update Logic deconstruction to read from Spec fields

### 4B: Add inputTypeSpec Type

**New or modified file: `internal/parser_ai/type_spec.go`**

```go
type inputTypeSpec struct {
    Notation      string `json:"notation,omitempty"`
    Specification string `json:"specification,omitempty"`
}
```

### 4C: Update inputAttribute and inputParameter

**Modified file: `internal/parser_ai/class.go`**

```go
type inputAttribute struct {
    Name             string         `json:"name"`
    DataTypeRules    string         `json:"data_type_rules,omitempty"`
    PreciseType      *inputTypeSpec `json:"precise_type,omitempty"`  // NEW
    Details          string         `json:"details,omitempty"`
    DerivationPolicy *inputLogic   `json:"derivation_policy,omitempty"`
    Nullable         bool           `json:"nullable,omitempty"`
    UMLComment       string         `json:"uml_comment,omitempty"`
    Invariants       []inputLogic   `json:"invariants,omitempty"`    // NEW: attribute invariants
}
```

Each invariant in `Invariants` is a `LogicTypeAssessment` logic. The AI parser should enforce that all attribute invariants have `type: "assessment"` (or auto-set it since no other type is valid for attribute invariants).

**Modified file: `internal/parser_ai/parameter.go`**

```go
type inputParameter struct {
    Name          string         `json:"name"`
    DataTypeRules string         `json:"data_type_rules,omitempty"`
    PreciseType   *inputTypeSpec `json:"precise_type,omitempty"` // NEW
}
```

### 4D: Update inputLogic

**Modified file: `internal/parser_ai/logic.go`**

Add target type spec:
```go
type inputLogic struct {
    Type              string         `json:"type,omitempty"`
    Description       string         `json:"description"`
    Target            string         `json:"target,omitempty"`
    Notation          string         `json:"notation,omitempty"`
    Specification     string         `json:"specification,omitempty"`
    TargetTypeSpec    *inputTypeSpec `json:"target_type_spec,omitempty"` // NEW
}
```

### 4E: Add inputNamedSet Type

**New file: `internal/parser_ai/named_set.go`**

```go
type inputNamedSet struct {
    Name          string `json:"name"`
    Description   string `json:"description,omitempty"`
    Notation      string `json:"notation,omitempty"`
    Specification string `json:"specification,omitempty"`
}
```

### 4F: Update JSON Schemas

**Modified schemas:**
- `attribute.schema.json` — add `precise_type` object with `notation`, `specification` fields; add `invariants` array of logic objects (assessment type only)
- `parameter.schema.json` — add `precise_type` object
- `logic.schema.json` — add `target_type_spec` object
- **New schema:** `named_set.schema.json` — schema for named set input
- **New schema:** `type_spec.schema.json` — shared schema for type spec objects
- `model.schema.json` (or equivalent) — add `named_sets` map

All schema descriptions should teach the AI how to correctly fill out the data, including examples of TLA+ type expressions.

The `attribute.schema.json` invariants description should explain:
- Each invariant is a boolean predicate (assessment type) about the attribute's value
- In TLA+ specifications, `attribute` is the binding that refers to the attribute's current value
- Example: `"attribute >= 0"` for a non-negative constraint, `"attribute \\in {\"active\", \"inactive\"}"` for enum membership

### 4G: Update Conversion Functions

**Modified file: `convert_to_model.go`**
- `convertAttributeToModel()` — parse inputTypeSpec → TypeSpec → pass to DataType; convert `Invariants []inputLogic` → `[]model_logic.Logic` with `LogicTypeAssessment` type and attribute invariant keys
- `convertParametersToModel()` — same pattern for TypeSpec (no invariants on parameters)
- `convertLogicToModel()` / `convertLogicsToModel()` — construct ExpressionSpec, handle TargetTypeSpec
- New: `convertNamedSetToModel()` — inputNamedSet → model_named_set.NamedSet

**Modified file: `convert_from_model.go`**
- `convertAttributeFromModel()` — extract TypeSpec from DataType → inputTypeSpec; convert `Invariants []model_logic.Logic` → `[]inputLogic`
- `convertParametersFromModel()` — same for TypeSpec
- `convertLogicFromModel()` — extract from Spec fields, handle TargetTypeSpec
- New: `convertNamedSetFromModel()` — NamedSet → inputNamedSet

### 4H: Add Error Codes

New error code range for type spec validation (22xxx or next available):
- Errors for invalid notation
- Errors for unparseable TLA+ type specification
- Errors for type spec on non-existent attribute
- Errors for invalid named set definition
- Errors for attribute invariant with wrong logic type (must be assessment)
- Errors for invalid attribute invariant specification (unparseable TLA+)
- Errors for attribute invariant missing description

Each error code gets a `.md` file in `errors/` that instructs the AI how to correct the error.

### 4I: Update Round-Trip Test

**Modified file: `test_helper/` (strict model)**

Add named sets and type specs to `GetStrictTestModel()` so the AI parser round-trip test exercises the new fields. The attribute invariants added in step 1G-ii are already in the flexible test model (which the strict model inherits), so the round-trip test will automatically exercise attribute invariant serialization/deserialization.

### 4J: Verify Stage 4

```bash
cd /workspaces/glemzurg/apps/requirements/req
go test ./internal/parser_ai/...
```

---

## Stage 5: database Tree — Postgres Mapping

**Goal:** Store the vetted model structure in the database.

**Packages touched:** `internal/database/`

**Test command:** `go test ./internal/database/... -dbtests`

### 5A: Schema Changes — DONE

**Modified file: `internal/database/sql/schema.sql`**

**Design change:** The `expression_node` and `expression_type` tables were removed from the design. Parsed expression trees (intermediate representation) are NOT stored in the database. Instead, specification text is stored as notation+specification string pairs, and expressions are parsed at load time via `convert.LowerAllExpressions()`.

What was actually implemented:

1. Added `let` to `logic_type` enum:
```sql
CREATE TYPE logic_type AS ENUM ('assessment', 'state_change', 'query', 'safety_rule', 'value', 'let');
```

2. Added `named_set` table with `type_spec_notation` and `type_spec_specification` columns.

3. Added `attribute_invariant` join table (same pattern as `class_invariant`).

4. Added `target_type_notation` and `target_type_specification` columns to `logic` table.

5. Added `type_spec_notation` and `type_spec_specification` columns to `data_type` table.

6. NOT implemented: `expression_type` table, `expression_node` changes (removed from design).

### 5B: New Data Access File — named_set.go — DONE

**New file: `internal/database/named_set.go`** — `scanNamedSet`, `LoadNamedSet`, `AddNamedSet`/`AddNamedSets`, `QueryNamedSets`, `RemoveNamedSet`. Handles ExpressionSpec and optional TypeSpec reconstruction via constructors with nil parseFunc.

**New file: `internal/database/named_set_test.go`** — round-trip tests.

### 5C: New Data Access File — attribute_invariant.go — DONE

**New file: `internal/database/attribute_invariant.go`** — `LoadAttributeInvariant`, `AddAttributeInvariant`/`AddAttributeInvariants`, `QueryAttributeInvariants` (grouped by attribute key), `RemoveAttributeInvariant`.

**New file: `internal/database/attribute_invariant_test.go`** — round-trip tests.

### 5D: Update data_type.go — DONE

Added `type_spec_notation` and `type_spec_specification` nullable columns to scan/insert/update. TypeSpec reconstructed via `NewTypeSpec()` constructor.

### 5E: Update logic.go — DONE

Added `target_type_notation` and `target_type_specification` nullable columns to scan/insert/update. TargetTypeSpec reconstructed via `NewTypeSpec()` constructor.

### 5F: Update top_level_requirements.go — DONE

**WriteModel:** Collects and inserts named_set rows, attribute invariant logic entries + join rows.

**ReadModel:** Loads named_sets, loads attribute_invariant join rows, stitches invariant Logic slices onto Attributes. After full tree assembly, calls `convert.LowerAllExpressions(&model)` to parse all TLA+ expressions with full context.

### 5G: Expression Lowering Consolidation — DONE (beyond original plan)

Three nearly-identical `lower_expressions.go` files (in parser_human, parser_ai, test_helper) were consolidated into a single shared function:

**New file: `internal/notation/tla_plus/convert/lower_all_expressions.go`** — `LowerAllExpressions(model *req_model.Model) error`

All four consumers (parser_human, parser_ai, test_helper, database) now call `convert.LowerAllExpressions()`. The three duplicate files and the `StripExpressions` workaround were deleted (~730 lines removed).

### 5H: Regenerate Database Docs — DONE

### 5I: Verify Stage 5 — DONE

```bash
cd /workspaces/glemzurg/apps/requirements/req
go test ./internal/database/... -dbtests  # PASS
go test ./...                              # PASS
```

---

## Stage 6: Remaining Trees — req_flat, generate

**Goal:** Fix compile errors and pass through new fields.

**Packages touched:** `internal/req_flat/`, `internal/generate/`

### 6A: req_flat

Check if `req_flat` accesses Logic.Notation, Logic.Specification, or Logic.Expression directly. If so, update to `logic.Spec.Notation`, etc.

Check if it exposes DataType in lookups — if TypeSpec needs to be available in templates, add it.

Likely minimal changes since TypeSpec is optional and not rendered by existing templates.

### 6B: generate

Templates consume DataType via `data_type_rules`. No changes needed unless templates want to render precise types — that's future work.

Fix any compile errors from changed constructors.

### 6C: Verify Stage 6

```bash
cd /workspaces/glemzurg/apps/requirements/req
go test ./internal/req_flat/...
go test ./internal/generate/...
```

---

## Stage 7: Full Integration Verification

```bash
cd /workspaces/glemzurg/apps/requirements/req
go test ./...
```

All tests pass. The system now has:
- ExpressionType hierarchy in `model_expression_type`
- Reusable ExpressionSpec/TypeSpec value objects in `model_spec`
- TypeSpec on DataType for precise type declarations
- TargetTypeSpec on Logic for result type declarations
- NamedSet model entity for shared set definitions
- NamedSetRef expression node for behavioral logic
- Attribute invariants (assessment logic with `attribute` binding)
- TLA+ type expression parsing (notation layer)
- Human parser reads `precise_type:` and attribute `invariants:` fields
- AI parser accepts/produces type specs, named sets, and attribute invariants
- Database stores and retrieves all new data (including attribute_invariant join table)

---

## What This Plan Does NOT Include

These are deferred to future sessions:

1. **Compatibility checker** — `CheckCompatibility(DataType, ExpressionType)`. Important but not structural — it's a validation function that can be added after the data structures exist.

2. **Simulator type checker integration** — Using declared ExpressionTypes as type hints in the HM inference engine. Requires the full pipeline to exist first.

3. ~~**LET/IN shared local variables**~~ — DONE. `LogicTypeLet` implemented as a new logic type. Let bindings can appear in requires, guarantees, safety_rules, and invariants. Fully supported in model, test helper, parser_human, parser_ai (including JSON schemas). NOT yet in database `logic_type` enum (deferred to Stage 5).

4. **ObjectClassKey migration** — Changing `Atomic.ObjectClassKey` from `*string` to `*identity.Key`. Small cleanup.

5. **notation/ast → notation/tla_plus/ast restructuring** — DONE. Moved to `notation/tla_plus/ast/` and `notation/tla_plus/parser/`.

6. ~~**TLA+ lowering/raising passes**~~ — DONE. The full lowering pipeline (`ast.Expression` → `model_expression.Expression`) is implemented in `notation/tla_plus/convert/`. See [Completed: TLA+ Lowering Pass](#completed-tla-lowering-pass) below for details.

7. **Test model enrichment** — PARTIALLY DONE. TargetTypeSpecs and attribute invariants (with let types) are in the test models. Named sets are NOT yet in the test model (`NamedSets: nil`). Precise types on DataType (`DataType.TypeSpec`) are NOT yet in the test models.

8. **Enumeration ordering** (design doc Gap 6) — `DataType.EnumOrdered *bool` should control whether comparison operators are valid on enum values. The expression type checker should enforce this.

9. **Nullable attributes** (design doc Gap 7) — Design decision that `Attribute.Nullable bool` is a DataType constraint, not a structural type concern. ExpressionType remains non-nullable.

10. **Derivation policy type checking** (design doc Gap 8) — `Attribute.DerivationPolicy` expression's result type must match the attribute's declared ExpressionType.

11. **Association class attributes** (design doc Gap 9) — How to access attributes of association classes during traversal, using LET/IN mechanism.

12. **Generation templates consume ExpressionType** (design doc Gap 10) — Templates may need to render ExpressionType alongside DataTypeRules. The `req_flat` layer may need to expose ExpressionType.

13. **Parameter type enforcement** (design doc Gap 11) — Action/query parameters have DataTypes that are never type-checked during expression evaluation. When ExpressionType exists on a parameter's DataType, the type checker should use it.

14. **No function values at runtime** (design doc Gap 12) — The simulator's `object` package has no `Function` type for runtime values, limiting higher-order patterns.

15. **Scenario test data typing** (design doc Gap 13) — When scenarios need concrete test data, values must conform to both DataType and ExpressionType.

16. **Class associations and object type traversal** (design doc section) — How `FieldAccess` does double duty for attribute access and association traversal. `TypeContext` with `AssociationTypes` for type-checking association navigation. Multiplicity-based result type mapping.

17. **Builtin call type signatures** (design doc section) — Defining formal type signatures for builtin calls (`_Stack!Push`, `_FiniteSet!Cardinality`, etc.) connecting the expression layer to the precise type system.

18. **NamedSetRef.ResolveType()** — Function to resolve `NamedSetRef` type nodes to the actual structural type at type-checking time.

19. ~~**expression_node.named_set_key column**~~ — Addressed: added to Stage 5A (items 7 and 8).

---

## Completed: TLA+ Lowering Pass

The full pipeline for converting `notation/tla_plus/ast.Expression` → `model_expression.Expression` is implemented and the simulator is rewired to consume `model_expression` nodes instead of AST nodes. This work was done on the `feature/intermediate-representation` branch across multiple sessions.

### What Was Built

**Stage 0: `model_expression` uses `math/big`**
- `IntLiteral.Value` changed from `int64` to `*big.Int`
- `RationalLiteral` changed from `Numerator/Denominator int64` to `Value *big.Rat`
- Aligns with the simulator's runtime `Number` type which uses `*big.Rat`

**Stage 1: Lowering pass — `internal/notation/tla_plus/convert/`**

New files:
- `convert/lower.go` — `Lower(expr ast.Expression, ctx *LowerContext) (model_expression.Expression, error)` — single-pass converter with class context for semantic resolution
- `convert/lower_test.go` — comprehensive test suite

`LowerContext` carries:
- `ClassKey`, `AttributeNames` (name → `identity.Key`), `ActionNames`, `QueryNames` — for resolving identifiers to `AttributeRef`/`ActionCall`/etc.
- `ParameterNames` — for resolving action/query parameters to `LocalVar`
- `GlobalFunctions`, `NamedSets` — for model-level resolution
- `AllActions` — for cross-class action call resolution (fully-qualified TLA+ names)

Key behaviors:
- `Identifier("self")` → `SelfRef{}`
- `Identifier(attrName)` → `AttributeRef{Key}` (when in class context and not quantifier-bound)
- `Identifier(paramName)` → `LocalVar{Name}` (when in action/query parameter scope)
- Quantifier/SetFilter variable bindings shadow attribute names during recursion
- `Parenthesized` nodes are unwrapped (transparent)
- `NumberLiteral` → `IntLiteral{*big.Int}` or `RationalLiteral{*big.Rat}`
- All operator strings (Unicode `∧∨⇒≡<>≤≥=≠∈∉∪∩⊆⊂⊇⊃⊕⊖⊏⊑⊐⊒`) mapped to typed enums
- `FunctionCall` routing: builtin → `BuiltinCall`, global → `GlobalCall`, same-class → `ActionCall`, cross-class → `ActionCall`
- Unresolved identifiers produce errors (strict resolution)

**Stage 2: Model population — `convert/lower_model.go`**

New files:
- `convert/lower_model.go` — `LowerModel(model *req_model.Model) error`
- `convert/lower_model_test.go` — tests using `test_helper.GetTestModel()`

`LowerModel()` walks the entire model tree and populates every `ExpressionSpec.Expression` field:
1. Parses `ExpressionSpec.Specification` via `parser.ParseExpression()`
2. Builds a `LowerContext` from the class hierarchy (attribute names, actions, queries, parameters)
3. Calls `Lower(parsedAST, ctx)` to get the `model_expression.Expression`
4. Sets `ExpressionSpec.Expression = result`

Walk order covers: Model.Invariants, Model.GlobalFunctions, Class.Invariants, Attribute.DerivationPolicy, Action.Requires/Guarantees/SafetyRules, Guard.Logic, Query.Requires/Guarantees.

Empty specifications (`Specification == ""`) are silently skipped. Errors are collected with the Logic.Key identifying which expression failed.

**Stage 3: Simulator rewiring**

Modified files:
- `evaluator/eval.go` — new `Eval()` dispatches on `model_expression.Expression` (38 node types); old AST-based evaluator renamed to `EvalAST()`
- `evaluator/eval_me.go` — new file with all `model_expression` evaluation handlers
- `actions/executor.go` — uses pre-lowered `logic.Spec.Expression` instead of parsing per-execution
- `actions/guard_evaluator.go` — uses `guard.Logic.Spec.Expression`
- `invariants/invariant_checker.go` — stores `model_expression.Expression` from `inv.Spec.Expression`
- `engine/derived_evaluator.go` — uses `attr.DerivationPolicy.Spec.Expression`
- `model_bridge/definition_builder.go` — registers `model_expression.Expression` in registry; new `ContainsAnyPrimedME()` walks model_expression tree
- `registry/` — `Definition.Body` type changed from `ast.Expression` to `model_expression.Expression`
- All evaluator test files rewritten to construct `model_expression` nodes

Key simplifications from rewiring:
- No operator string matching at eval time (enums replace Unicode strings)
- No identifier resolution at eval time (already resolved to `AttributeRef`/`SelfRef`/`LocalVar`)
- No quantifier decomposition at eval time (Variable+Domain already separated by lowering)
- No EXCEPT field tracking at eval time (`PriorFieldValue` has field name from lowering)

### Impact on This Plan

- **Stage 1H (Simulator Alignment)** will need to account for the existing `model_expression`-based evaluator when changing Logic field access patterns (e.g., `logic.Spec.Notation` instead of `logic.Notation`)
- **`ExpressionSpec.Expression` is now populated** — after `LowerModel()` runs, all Logic objects have their Expression field set. Stage 1C's refactoring of Logic to use ExpressionSpec is already partially realized (Logic already has a `Spec` field with `ExpressionSpec`)
- **The evaluator no longer imports `notation/tla_plus/ast`** — it only depends on `model_expression`. This clean separation means notation changes (Stage 2) won't affect the evaluator
- **`LowerModel()` must be called** before any simulator operation — this is the new initialization step that populates all expression trees from their TLA+ specification strings

---

## Risk Notes

- **Stage 1 has the largest blast radius.** Changing Logic's constructor touches ~50+ files across req_model, test_helper, and simulator. This is mechanical but extensive.
- **Stage 2 may require PEG grammar expertise.** Record type syntax (`[f: T]`) vs record value syntax (`[f |-> v]`) disambiguation in the parser could be tricky.
- **Stage 5 depends on all model types being finalized.** Database schema is the last to change because it reflects the vetted Go structures.
- **Each stage breaks downstream packages.** This is by design — the review cycle for each stage focuses on getting that stage's tree correct before moving on.
