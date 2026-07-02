# Refinement Specification Plan

## Purpose

The modeler today simulates a single state-machine **requirements** model. This plan adds a second model kind, the **specification** (or "refinement spec"), which a human author writes by hand to commit the requirements to specific implementation choices.

Goals, in order of priority:

1. Stay inside what this tool is good at: state-machine modeling and simulation.
2. Allow a user to view, edit, and simulate a requirements model **without** any spec present.
3. Allow a user to author a spec model that explicitly traces back to the requirements model it refines.
4. Make traceability between spec and requirements **automatically checkable** and **visually verifiable** in the generated markdown.
5. Leave room for a future "pattern library" of joint-simulation scenarios. That work is out of scope here.

The refinement spec itself is wholly hand-designed by the user. The tool does not infer trace links, generate scenarios, or attempt automated refinement proofs in this iteration.

---

## Core idea

A specification is just **another `core.Model`**, parsed by the same two parsers, simulated by the same engine, rendered by the same templates. The only thing that distinguishes it from a requirements model is:

1. A top-level `kind: "specification"` (default for existing files stays `requirements`).
2. A top-level `refines_model: "<requirements-model-key>"`.
3. An optional, user-authored `refines: "<kind>:<key>"` field on every linkable element (class, state, event, action, query, attribute, association, actor, use case, domain, subdomain, generalization, guard).

That is the entire data-model change. Every downstream feature — markdown rendering, traceability check, future joint simulation — reads from this single `refines` graph.

A spec element without a `refines:` value is allowed and is reported as a "spec-only" implementation detail. The tool never invents a `refines:` link.

🛑 **STOP for human review.** Confirm the scope, the "spec is just another model" framing, and the choice to make `refines:` the only new authoring concept before any code is written.

---

## Step 1: Core data-model changes

All changes in this step live inside [internal/core](file:///workspaces/glemzurg/apps/requirements/req/internal/core). No parser, generator, simulator, or CLI is touched. After Step 1 the codebase still behaves exactly as it does today for every existing model file; the new fields are simply available, default-zero, and unread by anything outside `core`.

The work is broken into six small **layers**. Each layer is a self-contained PR-sized change that compiles, passes tests, and leaves the system runnable. We implement them in order. Do not fold layers together.

### Layer 1.1 — `Refines` value object

A new tiny type that represents a single user-authored refinement link. Pure data and parsing. No host structs reference it yet.

1. New file `internal/core/refines.go` with:
   ```go
   // Refines is a single hand-authored link from a specification element
   // to the requirements element it implements. The link is to a key inside
   // the requirements model named by Model.RefinesModel; cross-model
   // resolution is performed by the trace checker, not by core.
   type Refines struct {
       Kind string // identity key-type string, e.g. "class", "state", "action".
       Key  string // The element's key string within the requirements model.
   }
   ```
2. Add `ParseRefines(s string) (Refines, error)` accepting the canonical syntax `<kind>:<key>`. Whitespace is trimmed, kind is lowercased.
3. Add `(Refines) String() string` returning the canonical form (round-trippable).
4. Add `(Refines) Validate(ctx *coreerr.ValidationContext) error` checking:
   - Kind is non-empty and is a known `identity.KEY_TYPE_*` string. The set of allowed kinds is the union of every element kind that may carry `Refines` (the leaf and container kinds enumerated in Layers 1.4 and 1.5). Models intentionally do **not** appear in this set; cross-model refinement is named only by `Model.RefinesModel`.
   - Key is non-empty.
   - Key is shaped like an `identity.Key.String()` value of the declared kind. Re-use the existing identity parsers; do not write a second parser.
5. New `coreerr` codes (added to `codes.go`):
   - `RefinesSyntaxInvalid`
   - `RefinesKindUnknown`
   - `RefinesKindKeyMismatch`
   - `RefinesKeyRequired`
6. New file `internal/core/refines_test.go` with table-driven coverage of: well-formed parse, whitespace tolerance, missing colon, empty kind, empty key, unknown kind, kind/key shape mismatch, round-trip via `String`.

Acceptance: `go test ./internal/core/...` passes. No other package compiles differently because nothing else imports `Refines` yet.

### Layer 1.2 — `Model.Kind` enum

Adds the discriminator that lets a model declare it is a specification, with no other behavioural change.

1. New constants in a new file `internal/core/model_kind.go`:
   ```go
   const (
       ModelKindRequirements  = "requirements"  // default
       ModelKindSpecification = "specification"
   )
   ```
2. Add field `Kind string` to `core.Model` in [model.go](file:///workspaces/glemzurg/apps/requirements/req/internal/core/model.go).
3. Update `NewModel` to accept `kind string`, with a single helper `NewRequirementsModel(...)` and `NewSpecificationModel(...)` that wrap `NewModel` so call sites do not need to pass the constant. Existing call sites in tests and converters move to `NewRequirementsModel` so behaviour is unchanged.
4. In `Model.Validate`:
   - If `Kind` is empty, treat it as `ModelKindRequirements` (back-compat for existing JSON / human files that predate this field).
   - If `Kind` is set, it must be one of the two constants. Otherwise emit `ModelKindInvalid`.
5. New `coreerr` code: `ModelKindInvalid`.
6. Update `model_test.go` with positive and negative cases; update existing tests only enough to use `NewRequirementsModel` where they currently call `NewModel`.

Acceptance: `go test ./...` passes. No file under `internal/parser_*`, `internal/generate`, or `internal/simulator` is modified.

### Layer 1.3 — `Model.RefinesModel`

Adds the cross-model pointer, gated on `Kind`.

1. Add field `RefinesModel string` to `core.Model`.
2. Extend `NewSpecificationModel` to require a non-empty `refinesModel` argument; `NewRequirementsModel` does not accept it.
3. In `Model.Validate`:
   - If `Kind == ModelKindSpecification`, `RefinesModel` must be non-empty. Otherwise emit `ModelRefinesModelRequired`.
   - If `Kind == ModelKindRequirements`, `RefinesModel` must be empty. Otherwise emit `ModelRefinesModelForbidden`.
   - `RefinesModel`, if non-empty, must satisfy the same normalization rules as `Model.Key` (non-empty, lowercase, trimmed). Re-use the existing model-key validation helper.
4. New `coreerr` codes: `ModelRefinesModelRequired`, `ModelRefinesModelForbidden`.
5. Tests in `model_test.go` cover all four `(Kind, RefinesModel)` combinations.

Acceptance: `go test ./...` passes. Still no consumers outside `core`.

### Layer 1.4 — `Refines` on leaf elements

Adds the `Refines *Refines` pointer to the simplest, leaf-level element structs first. We choose leaves so we can validate the pattern in a small surface before propagating.

Targeted structs, in this order:

1. `model_class.Attribute`
2. `model_state.State`
3. `model_state.Event`
4. `model_state.Guard`
5. `model_state.Action`
6. `model_state.Query`
7. `model_state.Transition`
8. `model_state.StateAction`
9. `model_logic.Logic` (covers invariants, requires, guarantees, safety rules — all of which already share this struct)

For each struct:

- Add `Refines *Refines` field. Pointer so that absence is unambiguous in marshalled output and existing zero-values are preserved.
- The constructor function (e.g. `NewAttribute`) is **not** changed; `Refines` is set via a small setter `SetRefines(*Refines)` so the existing constructor signatures stay stable. This keeps test fixtures untouched.
- The element's `Validate` method gets a new helper call `validateRefines(ctx, modelKind)` that:
  - If `modelKind == ModelKindRequirements` and `Refines != nil`, return `RefinesNotAllowedOnRequirements`.
  - If `Refines != nil`, delegate to `Refines.Validate(ctx)`.
  - Otherwise, no-op.
- Because element `Validate` methods today do not know the model's kind, propagate it down via the existing `coreerr.ValidationContext`. Add a new accessor `ctx.ModelKind()` and set it at the top of `Model.Validate`. This is a one-line plumbing change per element validator.

New `coreerr` code: `RefinesNotAllowedOnRequirements`.

Tests:

- Each element's `_test.go` gets a small block of cases covering: unset (always valid), set on a spec model (valid), set on a requirements model (rejected), invalid `Refines` content (rejected via Layer 1.1).
- One end-to-end `model_test.go` case loads a hand-built spec model with refinement links on all leaf kinds and asserts `Validate` succeeds.

Acceptance: `go test ./internal/core/...` passes. Other packages still build because `Refines` is an optional pointer added to existing structs and never read elsewhere.

### Layer 1.5 — `Refines` on container elements

Same mechanical change as Layer 1.4, applied to the structures that own children. Done in a separate layer so the leaf change can be reviewed in isolation.

Targeted structs:

1. `model_class.Class`
2. `model_class.Association`
3. `model_actor.Actor`
4. `model_actor.Generalization`
5. `model_class` generalization (`cgeneralization`)
6. `model_use_case.UseCase`
7. `model_use_case.Generalization`
8. `model_domain.Domain`
9. `model_domain.Subdomain`
10. `model_domain.Association`

Same pattern as Layer 1.4: pointer field, `SetRefines` setter, `validateRefines` call inside the existing `Validate`/`ValidateWithParent` methods, kind propagation via `ValidationContext`.

Tests mirror Layer 1.4: per-struct round-trip and negative cases plus one end-to-end test that builds a spec model with a fully populated refinement graph and runs `Validate`.

Acceptance: `go test ./internal/core/...` passes. The model now structurally supports a hand-authored refinement spec; no parser or generator yet knows it exists.

### Layer 1.6 — Refinement walker helper

A read-only iterator that the trace checker (Step 4) will consume. Building it now keeps Step 4 small and forces us to confirm we can reach every `Refines` field from one entry point.

1. New file `internal/core/refines_walk.go` exposing:
   ```go
   // RefinesEntry is one occurrence of a Refines link inside a model.
   type RefinesEntry struct {
       OwnerKind string        // identity key-type of the spec element holding the link.
       OwnerKey  identity.Key  // The spec element's own key (Model elements use a synthetic key).
       Refines   Refines       // The link itself.
   }

   // WalkRefines visits every element in the model that carries a Refines
   // pointer and yields one RefinesEntry per non-nil link, in deterministic
   // (sorted by OwnerKey.String()) order. Walks the model regardless of Kind;
   // the caller decides what to do with the result.
   func WalkRefines(m *Model) []RefinesEntry
   ```
2. Implementation walks: model invariants, model-level class associations, actors, actor generalizations, domain associations, and recursively into each domain → subdomain → (use cases, classes, generalizations, associations) → class children (attributes, states, events, guards, actions, queries, transitions, state actions, invariants).
3. Helpers `RefinesByOwnerKind(entries []RefinesEntry) map[string][]RefinesEntry` and `CountByOwnerKind(...)` for the trace checker's coverage report.
4. Tests in `refines_walk_test.go` cover: empty model, requirements model with no refinements, spec model with refinements at every element kind. The "every element kind" test is the contract test for Step 4.

Acceptance: `go test ./internal/core/...` passes. The package now exposes everything Step 4 needs without any code outside `core` changing.

### What remains explicitly out of Step 1

- No JSON schema, AI converter, markdown parser, or markdown generator changes. Those are Steps 2, 3, and 5.
- No `req_check` subcommand. That is Step 4.
- No simulator changes. The simulator is content-agnostic about `Refines`.
- No constructor-signature changes; setters are used so the diff stays small and existing test fixtures remain valid.

🛑 **STOP for human review.** After Layer 1.6 the core supports the full refinement-spec data model end-to-end and exposes a single walker for downstream consumers. Confirm the layer ordering, the choice of pointer-plus-setter over expanded constructors, the kind propagation via `ValidationContext`, and the planned `coreerr` codes before any parser work begins in Step 2.

---

## Step 2: `parser_ai` (JSON-schema authoring)

In [internal/parser_ai](file:///workspaces/glemzurg/apps/requirements/req/internal/parser_ai):

1. Update [json_schemas/model.schema.json](file:///workspaces/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas/model.schema.json):
   - Add optional `kind: "requirements" | "specification"` (default `requirements`).
   - Add optional `refines_model: string`.
2. Update every per-element schema (class, state_machine, action, query, parameter, etc.) to allow an optional `refines: string` property with pattern `^[a-z_]+:[a-z0-9_]+$`.
3. Update `convert_to_model.go` and `convert_from_model.go` to round-trip the new fields.
4. Add or extend test files under `test_files/` to cover:
   - A minimal `kind: specification` model with one class refining a requirements class.
   - A spec class with a spec-only attribute (no `refines:`).
   - Negative cases: malformed `refines:` syntax, `refines:` on a requirements-kind file.
5. Add a short authoring guide at `parser_ai/docs/refinement.md`. It is intentionally brief and rule-based, suitable for an AI prompt to cite verbatim:
   - A spec file is structurally identical to a requirements file.
   - Set `kind: specification` and `refines_model: <key>` at the top.
   - For each element you intentionally implement, add `refines: <kind>:<key>`.
   - Omit `refines:` on implementation-only elements.
   - Never edit the requirements file when authoring a spec.
   - Run `req_check trace` to validate.

🛑 **STOP for human review.** Confirm the schema additions, the regex, and the wording of the AI authoring guide before propagating to the human parser.

---

## Step 3: `parser_human` (markdown-style authoring)

In [internal/parser_human](file:///workspaces/glemzurg/apps/requirements/req/internal/parser_human):

1. Extend the trailing key/value block (where `actor_key:` etc. live today, see [01_basic.md](file:///workspaces/glemzurg/apps/requirements/req/internal/parser_human/test_files/class/01_basic.md)) to accept one new optional line per element:
   ```
   refines: class:order
   ```
2. Extend the model-file header to accept:
   ```
   kind: specification
   refines_model: ecommerce_requirements
   ```
   Both default-omitted; existing files keep parsing unchanged.
3. Update [regex.go](file:///workspaces/glemzurg/apps/requirements/req/internal/parser_human/regex.go) and the per-element parsers to read and write the new lines.
4. Add round-trip test files under `test_files/` mirroring the additions made in Step 2.
5. Update the `parser_human` README with a short "Authoring a refinement spec" section. Identical rules to the AI guide so authors learn one concept whether they are typing markdown or generating JSON.

🛑 **STOP for human review.** Confirm the markdown syntax, the round-trip, and that all existing test files still pass without modification before changing any rendering.

---

## Step 4: Automated traceability check (`req_check trace`)

In [cmd/req_check](file:///workspaces/glemzurg/apps/requirements/req/cmd/req_check):

1. Add a `trace` subcommand: `req_check trace <reqs-model> <spec-model>`.
2. Load both models with the existing loader. Verify `spec.Kind == Specification` and `spec.RefinesModel == reqs.Model.Key`.
3. Walk every element with a `Refines` value and check, in this order:
   - **Broken link.** The referenced key does not exist in the requirements model.
   - **Kind mismatch.** Spec `state` refines a requirements `event`, etc.
   - **Containment mismatch.** Spec `state` refines a requirements `state` whose owning class is not refined by this state's owning class. Catches "the element is real but lives in the wrong parent".
   - **Attribute narrowing.** When a spec attribute refines a requirements attribute, its data-type rules must be a subset/narrowing. Reuse [DataTypeChecker](file:///workspaces/glemzurg/apps/requirements/req/internal/simulator/invariants/data_type_checker.go) for the comparison.
4. Walk every spec element with no `Refines` and emit a "spec-only" entry, grouped by kind.
5. Walk every requirements element of a configurable set of kinds (default: `class`, `state`, `action`, `event`) that has no spec element refining it and emit an "uncovered" entry.
6. Output formats:
   - `text` — concise, CI-friendly.
   - `json` — for tooling.
   - `markdown` — written next to the generated docs as `traceability_report.md`, so the same artifact a human reads is the one CI checks.
7. Exit codes:
   - Non-zero on any broken link, kind mismatch, containment mismatch, or narrowing failure.
   - Configurable (`--strict-coverage`) whether "spec-only" or "uncovered" findings also fail the run.

Tests cover each finding class with a small fixture model pair.

🛑 **STOP for human review.** Confirm the exact set of findings, the default coverage kinds, and the exit-code policy before wiring the report into the markdown generator.

---

## Step 5: Markdown rendering — three views

In [internal/generate](file:///workspaces/glemzurg/apps/requirements/req/internal/generate):

1. Add an optional `Spec *core.Model` parameter and a `View` enum to the top-level generator:
   - `requirements_only` — current behavior, default. The spec is invisible.
   - `specification_only` — render the spec model alone, with refinement badges.
   - `paired` — render the spec, with both the badges and a per-element traceability block.
2. Per-element badge, added at the top of [class.md.template](file:///workspaces/glemzurg/apps/requirements/req/internal/generate/templates/class.md.template) and the equivalent templates for actor, use_case, etc.:
   ```
   > **Refines:** [Order](../requirements/class.order.md) ✓
   ```
   Glyphs come from one helper:
   - ✓ link resolves and kinds match
   - ⚠ element is spec-only (no `refines:` line)
   - ✗ link is broken or kinds mismatch
3. Per-class **Traceability** sub-table listing each child element (state, action, event, attribute) and what it refines, or `— *(spec-only)* ⚠`.
4. Top-level **Refinement Coverage** table on the model index page: every requirements element of the coverage-kinds set, and which spec element (if any) refines it. This is the primary visual-check artifact.
5. Helper functions go in [template.go](file:///workspaces/glemzurg/apps/requirements/req/internal/generate/template.go) so the templates stay declarative and small.
6. The data feeding the badges, the per-class tables, and the coverage table is the **same data structure** the trace checker produces in Step 4. The visual artifact and the CI exit code can never disagree.

Tests:

- Golden-file tests for each view on the same fixture pair used in Step 4.
- A test asserting the requirements-only view of a requirements model is **byte-identical** to today's output.

🛑 **STOP for human review.** Confirm the three views, the badge style, and the placement of the coverage table before considering this iteration complete.

---

## Out of scope for this plan

The following are deliberately deferred so this iteration stays small:

- **Joint simulation.** Spec and requirements simulators continue to run independently. No paired/refinement engine.
- **Automated scenarios / pattern library.** The user authors all trace links by hand; the tool offers no scenario suggestions.
- **Fault injection, environment actors, refinement mappings, TLC export of the spec.** All future work that will consume the same `refines:` field this plan introduces.
- **Editing the requirements model from inside the spec workflow.** Specs are read-only consumers of requirements.

🛑 **STOP for human review.** Confirm the deferral list before any of the deferred items are picked up by a later plan.

---

## Suggested rollout order

1. Step 1 — `core` fields and validator.
2. Step 2 — `parser_ai` schemas, conversions, tests, AI guide.
3. Step 3 — `parser_human` regex, parsers, tests, human guide.
4. Step 4 — `req_check trace` command.
5. Step 5 — markdown views and helpers.

Each step is independently shippable. After each, the requirements-only path remains untouched, so a human can always fall back to today's behavior while reviewing the new spec workflow.

🛑 **STOP for human review** at the end of every step listed above before starting the next.
