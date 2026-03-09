# req_check Error Examples

This document shows what each error type looks like to an AI running `req_check`,
in both text and JSON output modes. These are real errors produced by running
`req_check` against intentionally broken model fixtures.

## Error Types

`req_check` produces three types of errors:

| Type | Source | JSON `"type"` | Description |
|------|--------|---------------|-------------|
| **Parse Error** | `parser_ai.ParseError` | `"parse"` | JSON parse, schema, cross-reference, completeness, key format, and conversion errors. Has `code`, `message`, `file`, optional `field` and `hint`. |
| **Validation Error** | `coreerr.ValidationError` | `"validation"` | Core model validation errors. Has `code`, `message`, `field`, optional `path`, `got`, `want`. In practice these are wrapped as ParseError (E21002) during conversion. |
| **Generic Error** | `error` | `"error"` | Unexpected internal errors (filesystem failures, programming bugs). Has only `message`. |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Model is valid |
| 1 | Validation errors found |
| 2 | Usage error (bad arguments) |

---

## Parse Errors

### Invalid JSON (E1003)

A JSON file contains syntax errors.

**Text output:**

```
E1003: failed to parse model JSON: invalid character 'o' in literal null (expecting 'u')
  file: model.json
  hint: ensure file contains valid JSON syntax

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E1003",
    "message": "failed to parse model JSON: invalid character 'o' in literal null (expecting 'u')",
    "file": "model.json",
    "hint": "ensure file contains valid JSON syntax"
  }
]
```

---

### Schema Violation - Missing Required Field (E1004)

JSON is valid but missing a required field (`name`).

**Text output:**

```
E1004: model JSON does not match schema: jsonschema: '' does not validate with
  file:///...model.schema.json#/required: missing properties: 'name'
  file: model.json
  hint: run: req_check --schema model

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E1004",
    "message": "model JSON does not match schema: jsonschema: '' does not validate with file:///...model.schema.json#/required: missing properties: 'name'",
    "file": "model.json",
    "hint": "run: req_check --schema model"
  }
]
```

---

### Schema Violation - Invalid Enum Value (E2006)

An actor has `"type": "robot"` instead of `"person"` or `"system"`.

**Text output:**

```
E2006: actor JSON does not match schema: jsonschema: '/type' does not validate with
  file:///...actor.schema.json#/properties/type/enum: value must be one of "person", "system"
  file: actors/bad_actor.actor.json
  hint: run: req_check --schema actor

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E2006",
    "message": "actor JSON does not match schema: jsonschema: '/type' does not validate with file:///...actor.schema.json#/properties/type/enum: value must be one of \"person\", \"system\"",
    "file": "actors/bad_actor.actor.json",
    "hint": "run: req_check --schema actor"
  }
]
```

---

### Schema Violation - Empty Attribute Name (E5004)

A class attribute has an empty `name` field.

**Text output:**

```
E5004: class JSON does not match schema: jsonschema: '/attributes/id/name' does not validate with
  file:///...class.schema.json#/.../minLength: length must be >= 1, but got 0
  file: domains/sales/subdomains/default/classes/order/class.json
  hint: run: req_check --schema class

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E5004",
    "message": "class JSON does not match schema: jsonschema: '/attributes/id/name' does not validate with file:///...class.schema.json#/.../minLength: length must be >= 1, but got 0",
    "file": "domains/sales/subdomains/default/classes/order/class.json",
    "hint": "run: req_check --schema class"
  }
]
```

---

### Schema Violation - Unknown Field (E7002)

A state machine JSON has a field not in the schema.

**Text output:**

```
E7002: state machine JSON does not match schema: jsonschema: '' does not validate with
  file:///...state_machine.schema.json#/additionalProperties: additionalProperties 'invalid_field' not allowed
  file: domains/sales/subdomains/default/classes/order/state_machine.json
  hint: run: req_check --schema state_machine

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E7002",
    "message": "state machine JSON does not match schema: jsonschema: '' does not validate with file:///...state_machine.schema.json#/additionalProperties: additionalProperties 'invalid_field' not allowed",
    "file": "domains/sales/subdomains/default/classes/order/state_machine.json",
    "hint": "run: req_check --schema state_machine"
  }
]
```

---

### Key Format Error (E11026)

A filename uses uppercase (keys must be lowercase snake_case).

**Text output:**

```
E11026: actor_key key 'BadKey' has invalid format - keys must be lowercase snake_case
  (e.g., 'order_line'); convert to lowercase
  file: actors/BadKey.actor.json
  field: actor_key
  hint: keys must be lowercase snake_case: ^[a-z][a-z0-9]*(_[a-z0-9]+)*$

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E11026",
    "message": "actor_key key 'BadKey' has invalid format - keys must be lowercase snake_case (e.g., 'order_line'); convert to lowercase",
    "file": "actors/BadKey.actor.json",
    "field": "actor_key",
    "hint": "keys must be lowercase snake_case: ^[a-z][a-z0-9]*(_[a-z0-9]+)*$"
  }
]
```

---

### Association Filename Error (E11027)

An association filename doesn't follow the `from--to--name.assoc.json` pattern.

**Text output:**

```
E11027: association filename 'badname' must have exactly 3 parts separated by '--'
  (from--to--name), found 1 parts
  file: domains/sales/subdomains/default/class_associations/badname.assoc.json
  field: filename
  hint: association filenames must follow the pattern: from--to--name.assoc.json

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E11027",
    "message": "association filename 'badname' must have exactly 3 parts separated by '--' (from--to--name), found 1 parts",
    "file": "domains/sales/subdomains/default/class_associations/badname.assoc.json",
    "field": "filename",
    "hint": "association filenames must follow the pattern: from--to--name.assoc.json"
  }
]
```

---

### Cross-Reference Error - Actor Not Found (E11001)

A class references an actor that doesn't exist.

**Text output:**

```
E11001: class 'order' references actor 'nonexistent_actor' which does not exist
  file: domains/sales/subdomains/default/classes/order/class.json
  field: actor_key
  hint: available actors: customer

2 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E11001",
    "message": "class 'order' references actor 'nonexistent_actor' which does not exist",
    "file": "domains/sales/subdomains/default/classes/order/class.json",
    "field": "actor_key",
    "hint": "available actors: customer"
  }
]
```

---

### Cross-Reference Error - State Not Found (E11008)

A transition references a state that doesn't exist in the state machine.

**Text output:**

```
E11008: transition[0] from_state_key 'nonexistent' does not exist
  file: domains/sales/subdomains/default/classes/order/state_machine.json
  field: transitions[0].from_state_key
  hint: available states: pending

2 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E11008",
    "message": "transition[0] from_state_key 'nonexistent' does not exist",
    "file": "domains/sales/subdomains/default/classes/order/state_machine.json",
    "field": "transitions[0].from_state_key",
    "hint": "available states: pending"
  }
]
```

---

### Completeness Error - No Actors (E11017)

The model has no actors defined.

**Text output:**

```
E11017: model must have at least one actor defined - actors represent the users, systems,
  or external entities that interact with your system; define actors in the 'actors/'
  directory with files like 'actors/user.actor.json'
  file: model.json
  field: actors
  hint: create actors/{key}.actor.json with {"name": ..., "type": "person|external_system|time"}

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E11017",
    "message": "model must have at least one actor defined - actors represent the users, systems, or external entities that interact with your system; define actors in the 'actors/' directory with files like 'actors/user.actor.json'",
    "file": "model.json",
    "field": "actors",
    "hint": "create actors/{key}.actor.json with {\"name\": ..., \"type\": \"person|external_system|time\"}"
  }
]
```

---

### Completeness Error - Missing State Machine (E11023)

A class has no `state_machine.json` file.

**Text output:**

```
E11023: class 'item' must have a state machine defined - state machines describe the
  lifecycle and behavior of a class; create a 'state_machine.json' file in the class
  directory with states, events, and transitions
  file: domains/sales/subdomains/default/classes/item/class.json
  field: state_machine
  hint: create state_machine.json with states, events, and transitions

2 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E11023",
    "message": "class 'item' must have a state machine defined - state machines describe the lifecycle and behavior of a class; create a 'state_machine.json' file in the class directory with states, events, and transitions",
    "file": "domains/sales/subdomains/default/classes/item/class.json",
    "field": "state_machine",
    "hint": "create state_machine.json with states, events, and transitions"
  }
]
```

---

### Invalid Multiplicity (E11016)

An association has a non-parseable multiplicity value.

**Text output:**

```
E11016: association 'order--item--contains' from_multiplicity 'abc' is invalid: invalid format
  file: domains/sales/subdomains/default/associations/order--item--contains.assoc.json
  field: from_multiplicity
  hint: valid multiplicities: 1, 0..1, *, 0..*, 1..*

2 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E11016",
    "message": "association 'order--item--contains' from_multiplicity 'abc' is invalid: invalid format",
    "file": "domains/sales/subdomains/default/associations/order--item--contains.assoc.json",
    "field": "from_multiplicity",
    "hint": "valid multiplicities: 1, 0..1, *, 0..*, 1..*"
  }
]
```

---

### Unreferenced Action (E11029)

An action is defined but never used in any state or transition.

**Text output:**

```
E11029: action 'create_order' in class 'order' is defined but not referenced by any state
  action or transition - every action must be used in the state machine either as a state
  entry/exit/do action or as a transition action
  file: domains/sales/subdomains/default/classes/order/actions/create_order.json
  field: action_key
  hint: reference this action in a state entry/exit/do or transition action_key

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E11029",
    "message": "action 'create_order' in class 'order' is defined but not referenced by any state action or transition - every action must be used in the state machine either as a state entry/exit/do action or as a transition action",
    "file": "domains/sales/subdomains/default/classes/order/actions/create_order.json",
    "field": "action_key",
    "hint": "reference this action in a state entry/exit/do or transition action_key"
  }
]
```

---

### Conversion Error - Model Validation Failed (E21002)

The model converts from input format but fails core validation. The wrapped core
validation error is embedded in the message as text.

**Text output:**

```
E21002: resulting model validation failed: requires 0: [LOGIC_SPEC_INVALID] logic
  "domain/sales/subdomain/default/class/item/action/update_item/arequire/0" spec:
  [EXPRSPEC_NOTATION_REQUIRED] Notation is required (field: Notation, want: one of: tla_plus)
  (field: Spec)
  file: model.json

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "parse",
    "code": "E21002",
    "message": "resulting model validation failed: requires 0: [LOGIC_SPEC_INVALID] logic \"domain/sales/subdomain/default/class/item/action/update_item/arequire/0\" spec: [EXPRSPEC_NOTATION_REQUIRED] Notation is required (field: Notation, want: one of: tla_plus) (field: Spec)",
    "file": "model.json"
  }
]
```

---

## Generic Errors

### Nonexistent Model Path

When the model directory doesn't exist.

**Text output:**

```
STOP AND REPORT THIS ERROR to the user. This is an unexpected internal error that
cannot be fixed by changing input files: open /path/to/model/model.json: no such file
or directory

1 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output:**

```json
[
  {
    "type": "error",
    "message": "STOP AND REPORT THIS ERROR to the user. This is an unexpected internal error that cannot be fixed by changing input files: open /path/to/model/model.json: no such file or directory"
  }
]
```

---

## Multiple Errors

When error accumulation is active, `req_check` reports all errors found in a single run.
Errors are separated by blank lines in text mode.

**Text output (3 errors):**

```
E11020: subdomain 'default' must have at least 2 classes defined (has 1) - a subdomain needs
  multiple classes to represent meaningful relationships; create class directories under
  'domains/sales/subdomains/default/classes/' with 'class.json' files
  file: domains/sales/subdomains/default/subdomain.json
  field: classes
  hint: create class directories under classes/ with class.json files

E11001: class 'order' references actor 'nonexistent_actor' which does not exist
  file: domains/sales/subdomains/default/classes/order/class.json
  field: actor_key
  hint: available actors: customer

2 error(s) found. Use --explain E{code} for detailed remediation.
```

**JSON output (3 errors):**

```json
[
  {
    "type": "parse",
    "code": "E11020",
    "message": "subdomain 'default' must have at least 2 classes defined (has 1)...",
    "file": "domains/sales/subdomains/default/subdomain.json",
    "field": "classes",
    "hint": "create class directories under classes/ with class.json files"
  },
  {
    "type": "parse",
    "code": "E11001",
    "message": "class 'order' references actor 'nonexistent_actor' which does not exist",
    "file": "domains/sales/subdomains/default/classes/order/class.json",
    "field": "actor_key",
    "hint": "available actors: customer"
  }
]
```

---

## Error Code Ranges

| Range | Entity | Example Codes |
|-------|--------|--------------|
| 1xxx | Model | E1001 (name required), E1003 (invalid JSON), E1004 (schema violation) |
| 2xxx | Actor | E2004 (type invalid), E2006 (schema violation) |
| 3xxx | Domain | E3001 (name required), E3004 (schema violation) |
| 4xxx | Subdomain | E4001 (name required), E4004 (schema violation) |
| 5xxx | Class | E5004 (schema violation), E5008 (attribute name empty) |
| 6xxx | Association | E6004 (schema violation), E6012 (multiplicity invalid) |
| 7xxx | State Machine | E7002 (schema violation), E7020 (state not found) |
| 8xxx | Action | E8004 (schema violation) |
| 9xxx | Query | E9004 (schema violation) |
| 10xxx | Class Generalization | E10004 (schema violation) |
| 11xxx | Tree Validation | E11001-E11036 (cross-refs, completeness, keys, naming) |
| 12xxx | Actor Generalization | E12004 (schema violation) |
| 13xxx | Use Case Generalization | E13004 (schema violation) |
| 14xxx | Logic | E14001 (description required) |
| 15xxx | Parameter | E15001 (name required) |
| 16xxx | Global Function | E16001 (name required) |
| 17xxx | Domain Association | E17005 (invalid JSON) |
| 18xxx | Use Case | E18004 (schema violation) |
| 19xxx | Scenario | E19004 (schema violation) |
| 20xxx | Use Case Shared | E20004 (schema violation) |
| 21xxx | Conversion | E21001-E21008 (key construction, validation, multiplicity) |
| 22xxx | Named Set | E22004 (schema violation) |

---

## Using --explain

Use `--explain E{code}` to get full remediation documentation for any error code:

```
$ req_check --explain E1003
# Model Invalid JSON (E1003)

The `model.json` file contains invalid JSON syntax and cannot be parsed.
...
```

---

## Observations

1. **All errors seen by the AI are ParseError type.** Core `ValidationError` objects are
   wrapped as E21002 (`ParseError`) during the conversion step inside `ReadModel`. The
   `"validation"` JSON type exists in the output format but is not produced in practice.

2. **Generic errors indicate internal failures.** They always start with
   `"STOP AND REPORT THIS ERROR"` and mean the model directory is missing or there's
   a bug in `req_check` itself.

3. **Error accumulation is active.** Multiple errors are reported in a single run rather
   than stopping at the first error. This lets the AI fix several issues per iteration.

4. **Hints include available values.** Cross-reference errors like E11001 and E11008
   list the valid options in the hint (e.g., `"available actors: customer"`), enabling
   the AI to self-correct without additional lookups.

5. **Schema violation messages embed the JSON path.** The jsonschema library reports the
   exact path to the failing field (e.g., `'/attributes/id/name'`), helping identify
   which nested field is wrong.

6. **The `--tree` output shows `__` separators for association filenames, but the parser
   expects `--` separators.** This is a discrepancy that could confuse an AI following
   `--tree` output. Association filenames must use `--` (e.g., `order--item--contains.assoc.json`).
