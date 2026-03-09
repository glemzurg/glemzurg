# req_check Error Examples

This document shows what each error type looks like to an AI running `req_check`.
Output is always JSON. These are real errors produced by running `req_check` against
intentionally broken model fixtures.

## Error Types

| Type | JSON `"type"` | Description |
|------|---------------|-------------|
| **Parse Error** | `"parse"` | JSON parse, schema, cross-reference, completeness, key format, and conversion errors. Has `code`, `message`, `file`, optional `field` and `hint`. |
| **Generic Error** | `"error"` | Unexpected internal errors (filesystem failures, programming bugs). Has only `message`. |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Model is valid |
| 1 | Validation errors found |
| 2 | Usage error (bad arguments) |

## Hint Format

Every parse error's `hint` field contains pipe-delimited guidance:
- **Actionable fix** — what to change (e.g., `available actors: customer`)
- **Schema errors** — `run: req_check --schema <entity>` (specific entity)
- **File structure errors** — `run: req_check --tree` (missing files, bad filenames)
- **All errors** — `run: req_check --explain E{code}` (specific code)

---

## Parse Errors

### Invalid JSON (E1003)

```json
[
  {
    "type": "parse",
    "code": "E1003",
    "message": "failed to parse model JSON: invalid character 'o' in literal null (expecting 'u')",
    "file": "model.json",
    "hint": "ensure file contains valid JSON syntax | run: req_check --explain E1003"
  }
]
```

---

### Schema Violation - Missing Required Field (E1004)

```json
[
  {
    "type": "parse",
    "code": "E1004",
    "message": "model JSON does not match schema: jsonschema: '' does not validate with file:///...model.schema.json#/required: missing properties: 'name'",
    "file": "model.json",
    "hint": "run: req_check --schema model | run: req_check --explain E1004"
  }
]
```

---

### Schema Violation - Invalid Enum Value (E2006)

```json
[
  {
    "type": "parse",
    "code": "E2006",
    "message": "actor JSON does not match schema: jsonschema: '/type' does not validate with file:///...actor.schema.json#/properties/type/enum: value must be one of \"person\", \"system\"",
    "file": "actors/bad_actor.actor.json",
    "hint": "run: req_check --schema actor | run: req_check --explain E2006"
  }
]
```

---

### Schema Violation - Empty Attribute Name (E5004)

```json
[
  {
    "type": "parse",
    "code": "E5004",
    "message": "class JSON does not match schema: jsonschema: '/attributes/id/name' does not validate with file:///...class.schema.json#/.../minLength: length must be >= 1, but got 0",
    "file": "domains/sales/subdomains/default/classes/order/class.json",
    "hint": "run: req_check --schema class | run: req_check --explain E5004"
  }
]
```

---

### Schema Violation - Unknown Field (E7002)

```json
[
  {
    "type": "parse",
    "code": "E7002",
    "message": "state machine JSON does not match schema: jsonschema: '' does not validate with file:///...state_machine.schema.json#/additionalProperties: additionalProperties 'invalid_field' not allowed",
    "file": "domains/sales/subdomains/default/classes/order/state_machine.json",
    "hint": "run: req_check --schema state_machine | run: req_check --explain E7002"
  }
]
```

---

### Key Format Error (E11026)

```json
[
  {
    "type": "parse",
    "code": "E11026",
    "message": "actor_key key 'BadKey' has invalid format - keys must be lowercase snake_case (e.g., 'order_line'); convert to lowercase",
    "file": "actors/BadKey.actor.json",
    "field": "actor_key",
    "hint": "keys must be lowercase snake_case: ^[a-z][a-z0-9]*(_[a-z0-9]+)*$ | run: req_check --tree | run: req_check --explain E11026"
  }
]
```

---

### Association Filename Error (E11027)

```json
[
  {
    "type": "parse",
    "code": "E11027",
    "message": "association filename 'badname' must have exactly 3 parts separated by '--' (from--to--name), found 1 parts",
    "file": "domains/sales/subdomains/default/class_associations/badname.assoc.json",
    "field": "filename",
    "hint": "association filenames must follow the pattern: from--to--name.assoc.json | run: req_check --tree | run: req_check --explain E11027"
  }
]
```

---

### Cross-Reference Error - Actor Not Found (E11001)

```json
[
  {
    "type": "parse",
    "code": "E11001",
    "message": "class 'order' references actor 'nonexistent_actor' which does not exist",
    "file": "domains/sales/subdomains/default/classes/order/class.json",
    "field": "actor_key",
    "hint": "available actors: customer | run: req_check --explain E11001"
  }
]
```

---

### Cross-Reference Error - State Not Found (E11008)

```json
[
  {
    "type": "parse",
    "code": "E11008",
    "message": "transition[0] from_state_key 'nonexistent' does not exist",
    "file": "domains/sales/subdomains/default/classes/order/state_machine.json",
    "field": "transitions[0].from_state_key",
    "hint": "available states: pending | run: req_check --explain E11008"
  }
]
```

---

### Completeness Error - No Actors (E11017)

```json
[
  {
    "type": "parse",
    "code": "E11017",
    "message": "model must have at least one actor defined - actors represent the users, systems, or external entities that interact with your system; define actors in the 'actors/' directory with files like 'actors/user.actor.json'",
    "file": "model.json",
    "field": "actors",
    "hint": "create actors/{key}.actor.json with {\"name\": ..., \"type\": \"person|external_system|time\"} | run: req_check --tree | run: req_check --explain E11017"
  }
]
```

---

### Completeness Error - Missing State Machine (E11023)

```json
[
  {
    "type": "parse",
    "code": "E11023",
    "message": "class 'item' must have a state machine defined - state machines describe the lifecycle and behavior of a class; create a 'state_machine.json' file in the class directory with states, events, and transitions",
    "file": "domains/sales/subdomains/default/classes/item/class.json",
    "field": "state_machine",
    "hint": "create state_machine.json with states, events, and transitions | run: req_check --tree | run: req_check --explain E11023"
  }
]
```

---

### Invalid Multiplicity (E11016)

```json
[
  {
    "type": "parse",
    "code": "E11016",
    "message": "association 'order--item--contains' from_multiplicity 'abc' is invalid: invalid format",
    "file": "domains/sales/subdomains/default/associations/order--item--contains.assoc.json",
    "field": "from_multiplicity",
    "hint": "valid multiplicities: 1, 0..1, *, 0..*, 1..* | run: req_check --explain E11016"
  }
]
```

---

### Unreferenced Action (E11029)

```json
[
  {
    "type": "parse",
    "code": "E11029",
    "message": "action 'create_order' in class 'order' is defined but not referenced by any state action or transition - every action must be used in the state machine either as a state entry/exit/do action or as a transition action",
    "file": "domains/sales/subdomains/default/classes/order/actions/create_order.json",
    "field": "action_key",
    "hint": "reference this action in a state entry/exit/do or transition action_key | run: req_check --explain E11029"
  }
]
```

---

### Conversion Error - Model Validation Failed (E21002)

```json
[
  {
    "type": "parse",
    "code": "E21002",
    "message": "resulting model validation failed: requires 0: [LOGIC_SPEC_INVALID] logic \"domain/sales/subdomain/default/class/item/action/update_item/arequire/0\" spec: [EXPRSPEC_NOTATION_REQUIRED] Notation is required (field: Notation, want: one of: tla_plus) (field: Spec)",
    "file": "model.json",
    "hint": "run: req_check --explain E21002"
  }
]
```

---

## Generic Errors

### Nonexistent Model Path

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

Error accumulation reports all errors in a single run:

```json
[
  {
    "type": "parse",
    "code": "E11020",
    "message": "subdomain 'default' must have at least 2 classes defined (has 1)...",
    "file": "domains/sales/subdomains/default/subdomain.json",
    "field": "classes",
    "hint": "create class directories under classes/ with class.json files | run: req_check --tree | run: req_check --explain E11020"
  },
  {
    "type": "parse",
    "code": "E11001",
    "message": "class 'order' references actor 'nonexistent_actor' which does not exist",
    "file": "domains/sales/subdomains/default/classes/order/class.json",
    "field": "actor_key",
    "hint": "available actors: customer | run: req_check --explain E11001"
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

## Available Commands

| Command | Purpose |
|---------|---------|
| `req_check <path>` | Validate model, JSON error output |
| `req_check --explain E{code}` | Full remediation docs for a specific error |
| `req_check --schema <entity>` | JSON schema for an entity type |
| `req_check --tree` | Expected directory tree structure |

---

## Observations

1. **All errors seen by the AI are ParseError type.** Core `ValidationError` objects are
   wrapped as E21002 (`ParseError`) during the conversion step inside `ReadModel`.

2. **Generic errors indicate internal failures.** They always start with
   `"STOP AND REPORT THIS ERROR"` and mean the model directory is missing or there's
   a bug in `req_check` itself.

3. **Error accumulation is active.** Multiple errors are reported in a single run rather
   than stopping at the first error. This lets the AI fix several issues per iteration.

4. **Hints include available values.** Cross-reference errors like E11001 and E11008
   list the valid options in the hint (e.g., `"available actors: customer"`), enabling
   the AI to self-correct without additional lookups.

5. **Every error hint includes its specific `--explain` code.** The AI can call
   `req_check --explain E{code}` for detailed remediation of any specific error.

6. **Schema errors point to `--schema`.** Schema violation hints direct the AI to
   `req_check --schema <entity>` for the full JSON schema.

7. **File structure errors include `--tree`.** Errors about misnamed files or unexpected
   file locations (e.g., bad key format, bad association filename) include `--tree` in
   the hint so the AI can see the expected directory layout.
