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

Module-qualified TLA+ constructors:

| Form | Meaning |
| --- | --- |
| `_Seq!Seq(T)` | Sequence (duplicates allowed) |
| `_Seq!SeqUnique(T)` | Sequence with unique elements |
| `_Set!_Set(T)` | Set |
| `_Bags!_Bag(T)` | Bag / multiset |

Replace `T` with any valid type expression, including nested forms.

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
_Set!_Set([id: Int, name: STRING])
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