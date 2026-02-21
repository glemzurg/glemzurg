# Plan: Refactor Scenario Node

## 1. Analysis of the Current Node

### What It Is

The `Node` struct (`model_scenario/node.go`) is the "abstract syntax tree" of a scenario's steps. A scenario is a sequence diagram belonging to a use case. It shows interactions between **objects** (instances of classes) arranged across the top of the diagram, with annotated horizontal arrows going left and right between them.

### Current Structure

```go
type Node struct {
    Statements    []Node        // Sub-nodes (for sequence and loop)
    Cases         []Case        // Cases (for switch)
    Loop          string        // Loop description
    Description   string        // Leaf description
    FromObjectKey *identity.Key // Source object
    ToObjectKey   *identity.Key // Target object
    EventKey      *identity.Key // Event reference (state-changing)
    ScenarioKey   *identity.Key // Cross-reference to another scenario
    AttributeKey  *identity.Key // Attribute reference (non-state-changing data flow)
    IsDelete      bool          // Delete operation
}
```

### Node Types (Inferred from Fields)

| Type | Discriminator | Purpose |
|------|--------------|---------|
| **Sequence** | `len(Statements) > 0` | Multiple sequential steps |
| **Switch** | `len(Cases) > 0` | Conditional branching (Alt/Opt in UML) |
| **Loop** | `Loop != ""` | Repeated steps |
| **Leaf** | None of the above (currently `""`, will become `"leaf"`) | A single interaction arrow |

### Leaf Types (Mutually Exclusive)

| Leaf Type | Discriminator | Meaning |
|-----------|--------------|---------|
| **Event** | `EventKey != nil` | An arrow representing a state-changing interaction. The event belongs to a class and carries parameters. Drawn as a **solid arrow**. |
| **Attribute** | `AttributeKey != nil` | An arrow representing a non-state-changing data flow (reading/querying data). The attribute belongs to a class. Drawn the same as an event arrow currently. |
| **Scenario** | `ScenarioKey != nil` | An arrow referencing another scenario (sub-scenario call). |
| **Delete** | `IsDelete == true` | A self-arrow indicating object deletion. No `ToObjectKey`. |

### Database Storage

The scenario table stores the entire Node tree as a **JSONB blob** in the `steps` column:

```sql
CREATE TABLE scenario (
    model_key text NOT NULL,
    scenario_key text NOT NULL,
    name text NOT NULL,
    use_case_key text NOT NULL,
    details text DEFAULT NULL,
    steps jsonb DEFAULT NULL,  -- Entire Node tree as JSON
    PRIMARY KEY (model_key, scenario_key)
);
```

### The Problem with AttributeKey

`AttributeKey` references a class attribute (`identity.KEY_TYPE_ATTRIBUTE`), but what it actually represents is a **non-state-changing data flow** — essentially a query. The current name and reference type are misleading:

1. **Semantic mismatch**: "Attribute" suggests a data field, but in the scenario it represents a data flow interaction (an arrow between objects).
2. **Missing query model**: The codebase already has `model_state.Query` — a first-class concept representing non-state-changing reads of a class. The scenario should reference queries, not attributes.
3. **Two kinds of arrows**: The scenario diagram has two fundamentally different arrow types:
   - **Events**: State-changing interactions (already correctly modeled with `EventKey`)
   - **Queries**: Non-state-changing data flows (currently mismodeled as `AttributeKey`)

## 2. Proposed Changes

### 2.1 Replace AttributeKey with QueryKey

In the `Node` struct, replace:
```go
AttributeKey  *identity.Key  // Remove
```
with:
```go
QueryKey      *identity.Key  // Add
```

This makes the leaf types:
- **Event** (`EventKey`): State-changing interaction arrow
- **Query** (`QueryKey`): Non-state-changing data flow arrow
- **Scenario** (`ScenarioKey`): Cross-scenario reference
- **Delete** (`IsDelete`): Object deletion

### 2.2 Fold Case into Node

The current `Case` struct is structurally identical to a `Node` — it has a condition string and child statements, just like a `Loop`. Both `Loop` and `Case` are "a condition plus children." This refactoring eliminates the `Case` struct entirely by:

1. **Removing the `Cases []Case` field** from Node.
2. **Merging `Loop string` into a general `Condition string`** field used by both loop and case node types.
3. **Adding a new `NODE_TYPE_CASE` node type.** A switch node's `Statements` are case-type nodes.

After this change, a switch is just a sequence whose children are all case nodes. The structural uniformity means one struct, one database table, one key type.

**Validation rule:** A case node must be a direct child of a switch node, and a switch node's children must all be case nodes.

### 2.3 Normalize Steps into Database Tables (Adjacency List)

Replace the JSONB `steps` column with a single normalized table using foreign keys.

#### New Table: `scenario_step`

Each row is a node in the tree. The tree structure is encoded via `parent_step_key` and `step_order`.

```sql
CREATE TABLE scenario_step (
    model_key text NOT NULL,
    scenario_step_key text NOT NULL,
    scenario_key text NOT NULL,
    parent_step_key text DEFAULT NULL,       -- NULL for root nodes
    step_order int NOT NULL,                 -- Order within parent
    step_type scenario_step_type NOT NULL,   -- 'sequence', 'switch', 'case', 'loop', 'leaf'
    condition text DEFAULT NULL,             -- Used by loop and case nodes
    description text DEFAULT NULL,           -- For leaf nodes
    from_object_key text DEFAULT NULL,       -- FK to scenario_object
    to_object_key text DEFAULT NULL,         -- FK to scenario_object
    event_key text DEFAULT NULL,             -- FK to event (via class)
    query_key text DEFAULT NULL,             -- FK to query (via class)
    scenario_ref_key text DEFAULT NULL,      -- FK to scenario (cross-reference)
    is_delete boolean NOT NULL DEFAULT false,
    PRIMARY KEY (model_key, scenario_step_key),
    CONSTRAINT fk_step_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE,
    CONSTRAINT fk_step_scenario FOREIGN KEY (model_key, scenario_key) REFERENCES scenario (model_key, scenario_key) ON DELETE CASCADE,
    CONSTRAINT fk_step_parent FOREIGN KEY (model_key, parent_step_key) REFERENCES scenario_step (model_key, scenario_step_key) ON DELETE CASCADE,
    CONSTRAINT fk_step_from_object FOREIGN KEY (model_key, from_object_key) REFERENCES scenario_object (model_key, scenario_object_key) ON DELETE CASCADE,
    CONSTRAINT fk_step_to_object FOREIGN KEY (model_key, to_object_key) REFERENCES scenario_object (model_key, scenario_object_key) ON DELETE CASCADE,
    CONSTRAINT fk_step_event FOREIGN KEY (model_key, event_key) REFERENCES event (model_key, event_key) ON DELETE CASCADE,
    CONSTRAINT fk_step_query FOREIGN KEY (model_key, query_key) REFERENCES query (model_key, query_key) ON DELETE CASCADE,
    CONSTRAINT fk_step_scenario_ref FOREIGN KEY (model_key, scenario_ref_key) REFERENCES scenario (model_key, scenario_key) ON DELETE CASCADE
);

CREATE TYPE scenario_step_type AS ENUM ('sequence', 'switch', 'case', 'loop', 'leaf');
```

No separate `scenario_step_case` table is needed — cases are just nodes with `step_type = 'case'` whose parent is a switch node.

#### Tree Reconstruction

The adjacency list is read with `ORDER BY step_order` and rebuilt into a tree in Go. The reconstruction algorithm:

1. Query all steps for a scenario, ordered by `step_order`.
2. Group steps by `parent_step_key` (NULL = root level).
3. Recursively build the `Node` tree from the grouped rows.
4. Every node type uses the same grouping — switch children are case nodes, case/loop/sequence children are their statements.

#### Remove `steps` Column

```sql
ALTER TABLE scenario DROP COLUMN steps;
```

### 2.4 Identity Key for Steps

#### Design Constraints

The `identity.Key` struct has strict rules:
- Format: `parentKey/keyType/subKey` (parsed by `ParseKey` using the last two `/`-delimited segments)
- `ValidateParent` checks that `k.ParentKey == parent.String()` exactly
- Each key type declares exactly one allowed parent type
- There is no precedent for a key type parenting itself

#### Option Considered: Nested Keys (Rejected)

Steps could nest deeper, e.g. `.../scenario/sc/sstep/root/sstep/child/sstep/grandchild`. This would encode the tree structure into the key path itself. However:
- `ValidateParent` would need to accept both `scenario` and `sstep` as parent types — no key type currently does this
- Keys would grow extremely long for deep nesting (a 5-level tree adds 5 `sstep/X` segments on top of the already-long scenario key)
- The tree structure is already encoded in the `parent_step_key` database column — duplicating it in the key is redundant
- `ParseKey` works (it always takes the last `keyType/subKey` pair), but the resulting keys are unwieldy

#### Chosen Design: Flat Keys (Parent = Scenario)

All steps use the scenario as their key parent, with a flat incrementing `subKey`:

```go
KEY_TYPE_SCENARIO_STEP = "sstep"  // Parent: scenario
```

Example keys for a scenario with 9 nodes:
```
domain/d/subdomain/s/usecase/uc/scenario/sc/sstep/0   (root sequence)
domain/d/subdomain/s/usecase/uc/scenario/sc/sstep/1   (leaf child)
domain/d/subdomain/s/usecase/uc/scenario/sc/sstep/2   (loop child)
domain/d/subdomain/s/usecase/uc/scenario/sc/sstep/3   (leaf grandchild of loop)
...
```

This follows the same pattern as `arequire`, `aguarantee`, `asafety`, `qrequire`, `qguarantee` — flat incrementing keys under a parent, where the ordering/structure is captured in other fields (in those cases `sort_order`, here `parent_step_key` and `step_order`).

The tree parent-child relationship between steps is encoded entirely in the database column `parent_step_key`, not in the key hierarchy. The key hierarchy only says "this step belongs to this scenario."

#### Constructor and Validation

```go
func NewScenarioStepKey(scenarioKey Key, subKey string) (Key, error)
```

`ValidateParent` in `identity/key.go`: parent must be a `scenario`.

The `subKey` is an incrementing string (e.g., `"0"`, `"1"`, `"2"`) assigned during parsing or migration. The parser walks the tree in order and assigns sequential subKeys.

### 2.5 Updated Node Struct

The `Case` struct is eliminated. The `Node` struct becomes:

```go
type Node struct {
    Key           identity.Key  // New: each node gets a key
    Statements    []Node
    Condition     string        // Used by both "loop" and "case" node types (replaces Loop and Case.Condition)
    Description   string
    FromObjectKey *identity.Key
    ToObjectKey   *identity.Key
    EventKey      *identity.Key
    QueryKey      *identity.Key  // Replaces AttributeKey
    ScenarioKey   *identity.Key
    IsDelete      bool
}
```

Node types become:

| Type | Discriminator | Condition | Statements |
|------|--------------|-----------|------------|
| **Sequence** | explicit or inferred | — | Children |
| **Switch** | explicit or inferred | — | Children (must all be case nodes) |
| **Case** | explicit or inferred | Required | Children |
| **Loop** | explicit or inferred | Required | Children |
| **Leaf** | no children | — | — |

### 2.6 Updated Constants

```go
// Node types:
NODE_TYPE_SEQUENCE = "sequence"
NODE_TYPE_SWITCH   = "switch"
NODE_TYPE_CASE     = "case"      // New (was the Case struct)
NODE_TYPE_LOOP     = "loop"
NODE_TYPE_LEAF     = "leaf"

// Leaf types:
LEAF_TYPE_EVENT    = "event"
LEAF_TYPE_QUERY    = "query"     // Replaces LEAF_TYPE_ATTRIBUTE
LEAF_TYPE_SCENARIO = "scenario"
LEAF_TYPE_DELETE   = "delete"
```

## 3. Files to Modify

### Identity Layer
| File | Change |
|------|--------|
| `identity/key_type.go` | Add `KEY_TYPE_SCENARIO_STEP`, constructor |
| `identity/key.go` | Add `ValidateParent` case for new key type |
| `identity/key_type_test.go` | Add test cases for new key constructor |

### Model Layer
| File | Change |
|------|--------|
| `model_scenario/node.go` | Eliminate `Case` struct, replace `AttributeKey` with `QueryKey`, replace `Cases []Case` and `Loop string` with `Condition string`, add `Key` field, add `NODE_TYPE_CASE`, update constants, validation, JSON/YAML marshaling |
| `model_scenario/node_test.go` | Update all tests: `attrKey` → query key, cases → case nodes, add key fields, update leaf type assertions |
| `model_scenario/scenario.go` | Update `ValidateWithParentAndClasses` to validate step keys |

### Database Layer
| File | Change |
|------|--------|
| `database/sql/schema.sql` | Add `scenario_step_type` enum, `scenario_step` table, drop `steps` column from `scenario` |
| `database/scenario.go.disable` → `scenario.go` | Remove JSON serialization, enable file |
| `database/scenario_object.go.disable` → `scenario_object.go` | Enable file |
| New: `database/scenario_step.go` | CRUD for scenario steps (all node types including case) |
| New: `database/scenario_step_test.go` | Tests |
| `database/top_level_requirements.go` | Add scenario steps to WriteModel/ReadModel |
| `database/top_level_requirements_test.go` | Add test data for scenario steps |

### Parser Layer
| File | Change |
|------|--------|
| `parser/use_case_scope_object_keys.go` | Replace `attribute_key` handling with `query_key`, add `expandQueryKey` function |
| `parser_json/scenario_node.go` | Replace `AttributeKey` with `QueryKey` in `nodeInOut` |

### Generator Layer
| File | Change |
|------|--------|
| `generate/scenario.go` | Replace `stmt.AttributeKey` with `stmt.QueryKey`, use query lookup instead of implicit attribute arrow |

## 4. Tree Encoding in the Adjacency List

### Example

Given this scenario tree (all nodes are the same `Node` struct):
```
sequence
  ├── leaf (event: order.place)          from: customer → to: system
  ├── loop [condition: "for each item"]
  │   └── leaf (query: inventory.check)  from: system → to: warehouse
  └── switch
      ├── case [condition: "items available"]
      │   └── leaf (event: order.confirm) from: system → to: customer
      └── case [condition: "items unavailable"]
          └── leaf (event: order.reject)  from: system → to: customer
```

### Database Rows

**scenario_step** (single table for all node types):

| scenario_step_key | parent_step_key | step_order | step_type | condition | event_key | query_key | ... |
|---|---|---|---|---|---|---|---|
| step_root | NULL | 0 | sequence | NULL | NULL | NULL | |
| step_1 | step_root | 0 | leaf | NULL | order.place | NULL | |
| step_2 | step_root | 1 | loop | for each item | NULL | NULL | |
| step_2_1 | step_2 | 0 | leaf | NULL | NULL | inventory.check | |
| step_3 | step_root | 2 | switch | NULL | NULL | NULL | |
| step_3_1 | step_3 | 0 | case | items available | NULL | NULL | |
| step_3_1_1 | step_3_1 | 0 | leaf | NULL | order.confirm | NULL | |
| step_3_2 | step_3 | 1 | case | items unavailable | NULL | NULL | |
| step_3_2_1 | step_3_2 | 0 | leaf | NULL | order.reject | NULL | |

Note: Case nodes are children of switch nodes. The case's own children (leaf steps) are children of the case node. No separate case table needed.

### Reconstruction Algorithm

```go
func buildTree(steps []ScenarioStep) *Node {
    // 1. Index steps by key
    // 2. Group child steps by parent_step_key (NULL = root)
    // 3. Starting from root steps, recursively build:
    //    - For sequence/loop/case: attach children as Statements
    //    - For switch: attach children as Statements (all must be case nodes)
    //    - For leaf: populate event/query/scenario/delete fields
}
```

## 5. Migration Considerations

- The JSON blob format must be parseable during migration to populate the new table.
- A migration script should read existing `steps` JSONB, assign keys to each node (including case nodes), and insert into `scenario_step`.
- After migration, drop the `steps` column.
- The parser must be updated to generate keys for nodes during parsing (before they had none).

## 6. Benefits

1. **Foreign key integrity**: References to events, queries, objects, and scenarios are enforced by the database.
2. **Queryability**: Individual steps can be queried, joined, and analyzed without parsing JSON.
3. **Semantic clarity**: `QueryKey` correctly represents non-state-changing data flow, matching the existing `model_state.Query` concept.
4. **Consistency**: All model data follows the same normalized pattern — no more JSON blobs.
5. **Structural simplicity**: One `Node` struct and one `scenario_step` table represent all node types (sequence, switch, case, loop, leaf) uniformly.
