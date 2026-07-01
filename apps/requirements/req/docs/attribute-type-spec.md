# Attribute `type_spec`

The `type_spec` field is an optional string on each attribute in `class.json`. It carries a **TLA+ type expression** — a formal, machine-oriented type that complements the human-readable `data_type_rules` string.

```json
"speed": {
    "name": "Speed",
    "data_type_rules": "unconstrained",
    "type_spec": "Nat"
}
```

## How `type_spec` relates to `data_type_rules`

| Field | Purpose |
| --- | --- |
| `data_type_rules` | Requirement-language DSL: spans, enums, collections, records (`unconstrained`, `enum of active, pending`, `[0 .. 100]`, and so on) |
| `type_spec` | TLA+ formal type for logic, specifications, and precise typing |

You can use one without the other. The JSON schema notes that `type_spec` is most useful when `data_type_rules` parses successfully into a `DataType`.

For the full `data_type_rules` syntax, see error **E5011** (Attribute Data Type Unparseable).

## Notation

`type_spec` values are always interpreted as **TLA+** (`tla_plus` internally). There is no separate notation field on attributes.

## Supported type expressions

The type system recognizes the following TLA+ forms.

### Scalars

| Value | Meaning |
| --- | --- |
| `BOOLEAN` | Boolean (`TRUE` / `FALSE`) |
| `Nat` | Natural numbers (0, 1, 2, …) |
| `Int` | Integers |
| `Real` | Rationals / reals |
| `STRING` | Strings |

### Finite enumerations

A set literal of string values:

```
{"active", "inactive", "pending"}
```

### Collections

Module-qualified TLA+ constructors and operators use the `_Module!Name` form in specifications (for example `_Seq!Len(seq)`). The leading underscore marks a standard-library module call, distinct from class-scoped actions.

**Stdlib module discipline.** Each `_Module` prefix names a real TLA+ standard module. Only operators that exist in that module's official definition may appear under that prefix. `_Seq`, `_Bags`, and similar names are not general-purpose namespaces for convenience helpers; they exist so specifications remain portable TLA+ that would mean the same thing in TLC or another conforming tool.

| Module prefix | Real TLA+ module | Allowed today (simulator) |
| --- | --- | --- |
| `_Seq!` | Sequences | `Head`, `Tail`, `Append`, `Len` |
| `_Bags!` | Bags | `SetToBag`, `BagToSet`, `CopiesIn`, `BagIn`, `BagCardinality` |
| `_Stack!`, `_Queue!` | (req data-type helpers) | Stack/queue ops on tuples — not TLA+ standard modules; used only for data-type sampling |

Do not add invented operators under `_Seq!` or `_Bags!` (for example a hypothetical `_FiniteSets!Sum`). When the standard libraries lack an operator you need, express the computation in plain TLA+ (conditionals, `LET`, `CHOOSE`, quantifiers, recursion) or as a model global function whose body is valid TLA+.

| Form | Meaning |
| --- | --- |
| `_Seq!Seq(T)` | Sequence (duplicates allowed) |
| `_Seq!SeqUnique(T)` | Sequence with unique elements |

Replace `T` with any valid type expression, including nested forms.

Sets and bags have no standard TLA+ type constructor. Use `data_type_rules` for unordered collections and multisets; in specifications use set literals, `\in`, and real `_Bags!` operators such as `SetToBag` and `BagCardinality` — not invented `_Set!_Set` or `_Bags!_Bag` forms.

### Records

Bracket syntax with named fields:

```
[name: STRING, age: Int]
```

### Tuples (Cartesian product)

```
Int \X STRING
```

When raised or printed, tuples may appear as `Int × STRING`.

### Nested examples

```
_Seq!Seq([id: Int, name: STRING])
_Seq!Seq(Int \X BOOLEAN)
```

### Other values seen in fixtures

`SUBSET STRING` and `SUBSET Int` appear in named-set definitions and database comments. These are stored as text but are **not** parsed by the TLA+ type converter today.

## What does not belong in `type_spec`

- Arbitrary English descriptions (use `data_type_rules` for those)
- Function types as type expressions
- Random identifiers (only the built-in scalars, enum literals, constructors, records, and tuples listed above)

## Practical notes

1. **Optional** — omit `type_spec` when informal typing via `data_type_rules` is enough.
2. **No syntax check at JSON load** — `parser_ai` stores `type_spec` as a string without validating TLA+ syntax when `class.json` is read. Malformed values do not fail class parsing.
3. **Common values in the repo** — `"Nat"`, `"STRING"`, `"Int"`, and `"SUBSET STRING"` (named sets).

## When to use it

Use `data_type_rules` alone for most attributes. Add `type_spec` when you need a precise TLA+ type for formal logic or specifications alongside a parseable `data_type_rules` value.

## Related documentation

- [JSON AI model format](../internal/parser_ai/docs/JSON_AI_MODEL_FORMAT.md) — `class.json` structure and attribute fields
- Error **E5011** — `data_type_rules` syntax reference
- Error **E11022** — class attribute completeness (actor-backed classes may have no attributes)