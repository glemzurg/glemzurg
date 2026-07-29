# Simulator packages and types

A junior-oriented map of `apps/requirements/req/internal/simulator/` and every package under it.  
**Package** and **type** descriptions are capped at three short sentences each.

## How the pieces fit (read this first)

Think of a simulation run as four layers:

1. **What may run** — `surface` decides which classes are in scope; `schema` freezes the static model for that run.
2. **What exists right now** — `instance` holds live objects, links, and state-machine positions; it also judges violations.
3. **What expressions mean** — `object` values + `evaluator` + `state` bindings turn TLA+ into runtime results.
4. **What happens next** — `actions` mutates the world for one event/query; `engine` chooses steps and runs the loop; `trace` / `report` show results.

Static facts live in **schema**. Mutable world data and “is this illegal?” live in **instance**. Other packages ask those two rather than rebuilding their own catalogs of truth.

---

## Package: `simulator` (root)

`internal/simulator` is the package root. It wires a few shared builders used across subpackages and is not the main run loop (that is `engine`).

### Types

#### `RegistryPipeline`

Holds a definition registry and a runtime adapter so scoped TLA+ definitions can be evaluated. Optionally carries a relation context for association field navigation.

#### `AssociationConfig`

Describes one association’s endpoints and multiplicities when building evaluator relation metadata. Used by the relation context builder, not during every simulation step.

#### `RelationContextBuilder`

Builds an `evaluator.RelationContext` from association configs so expressions can walk links by relation field name. Call this during setup, not inside hot step paths unless needed.

---

## Package: `schema`

Static brain for one simulation run: model facts, run scope, and indexes. After `schema.New(fullModel, RunScope)`, the run must not treat a separate `*core.Model` as a second authority.

### Types

#### `RunScope`

Names which class keys participate in this run. Empty/all scope means every non-realized class may be simulated.

#### `Schema`

Public API over private catalogs: keyed class/association lookups, class-sim metadata, derived attributes, association graph helpers, and eval factories. Engine and actions ask Schema questions; they do not rebuild parallel static indexes.

#### `EventInfo`

Lightweight event metadata for a class (keys/names used when selecting external drivers). Used by class simulation indexes and surface reporting.

#### `ClassSimInfo`

Per-class simulation facts: events, queries, derived attributes, and related surface membership. Built once into Schema so step selection does not re-walk the full model.

#### `AssociationInfo`

Static description of one association (endpoints, multiplicities, uniqueness, host/AC role). Looked up by key via Schema during linking and checks.

#### `AssociationClassInfo`

Static facts for an association class (the reifying class that materializes a host association). Used when creation must link host endpoints.

#### `AssociationClassLinkInfo`

Host association key plus endpoint class metadata needed to materialize or report an AC link. Shared (aliased) with `actions` so creation linking stays consistent.

#### `AssociationView`

A navigable view of an association for TLA-style binding and uniqueness walks. Prefer Schema methods over hand-walking model trees.

#### `HostAssociationInfo`

Identifies the host association for an association class. Needed when an AC instance should create or destroy the underlying binary host link.

#### `UniquenessBinding`

Describes one uniqueness constraint on an association (which attributes form the unique tuple). Checkers read these via Schema when judging duplicate links.

#### `IndexDefinition`

One class index: ordered attribute keys that must be unique together. Used by index uniqueness checking and Schema attribute indexes.

#### `DerivedAttrDef`

Static definition of a derived attribute (expression source and type context). Engine evaluates these when a surface derived read is selected.

#### `AssociationBinder`

Interface for registering association binding hooks into Schema. Lets association TLA navigation stay pluggable without exporting the private catalog.

---

## Package: `instance`

Mutable world for one run: instances, links, SM positions, liveness obligations, and all violation judgment. Violations are only constructed here; outside code calls `Check*` / checkers / `Package*` helpers.

### Types

#### `ID`

Opaque numeric identity of one live instance in a run. Prefer creating IDs through `State`, not inventing them in callers.

#### `Instance`

One live class object: ID, class key, and attribute record. Prefer `GetID` / `GetClassKey` / `GetAttributes` and mutate attributes via `SetAttribute` or `State.UpdateInstanceField`.

#### `State`

The whole mutable world for the run, plus a pointer to immutable `schema.Schema`. Create/update/delete instances and links here; run checks and liveness against this state.

#### `AssociationLink`

One host-association materialization row (from/to instance IDs for a host association). Used when association classes reify binary links.

#### `AssociationLinkTable`

Stores association-class host links separately from plain binary peer links. Keeps AC materialization queryable without overloading the evaluator link table.

#### `BinaryLink`

Protocol-friendly view of a binary association edge between two instance IDs. Returned by iteration helpers rather than dumping internal maps.

#### `Snapshot` / `SnapshotInstance`

Read-only freeze of instances (and related snapshot data) for inspection or reporting. Use when you need a stable view without holding write locks on live state.

#### `ExpressionBindings`

Interface instance checkers use to evaluate expressions without importing the full `state` package (avoids import cycles). Implementations are built by `state.BindingsBuilder`.

#### `StructuralInvariantCheckers`

Bundle of association/multiplicity/pair checkers used during and after action execution. Wired once into the action executor.

#### `InvariantChecker`

Runs authored TLA+ invariants (model, class, attribute-level) against live state. Consults Schema for expressions and class maps; never rebuilds the model tree.

#### `DataTypeChecker`

Checks attribute values against parsed data types and type_spec rules. Reports required/nullable, span, enum, collection size, and datetime mismatches.

#### `IndexUniquenessChecker`

Ensures class indexes are unique across live instances. Uses Schema index definitions and current attribute values.

#### `MultiplicityChecker`

Validates association endpoint multiplicities (for example 1 vs 0..*). Runs against live link counts for in-scope associations.

#### `AssociationUniquenessChecker`

Enforces association uniqueness tuples (no two links with the same uniqueness key). Complements plain “duplicate endpoint pair” checks.

#### `AssociationInstancePairChecker`

Detects duplicate links between the same two instances on associations that forbid them. Focused on pair identity, not attribute uniqueness.

#### `AssociationInvariantChecker`

Evaluates association-level authored invariants over current links and endpoints. Schema supplies which associations and expressions apply.

#### `LivenessHits` / `LivenessStep`

Records which classes, attributes, events, queries, and links were exercised during the run. Compared to obligations installed at `NewState` when checking liveness.

#### `DeferredAssertion` / `AssertionKind` / `AssessedFailure`

Internal shapes for packaging failed requires/guarantees/safety assessments into violations. Callers pass assessed failures; instance turns them into `ViolationError`s.

#### `PeerEventUnavailableInput` / `SurfaceMemberKind`

Inputs for packaging “peer event cannot run” and surface-member unavailability into violations. Used when selection or cascade cannot fire an external or peer driver cleanly.

#### `ViolationType`

Enum of every violation category (invariants, data types, multiplicity, liveness, safety, and so on). Drives reporting categories and human messages.

#### `ViolationError`

One concrete violation with type, message, and source identity. Only private constructors inside this package create these; use Check/Package APIs.

#### `ViolationErrors`

Slice of `*ViolationError` with helpers to append and merge results. The normal return type for “did this check fail?” APIs.

#### `ViolationSourceIdentity`

Who/what the violation is about (class, instance, attribute, association, event, and similar). Keeps messages locatable without free-form string soup.

#### Violation parameter structs

Structs such as `MultiplicityViolationParams`, `AssociationUniquenessViolationParams`, `PeerEventUnavailableParams`, and datetime type-spec mismatch params. They are typed inputs to private packaging helpers so callers do not assemble message text themselves.

---

## Package: `engine`

Top-level simulation loop: build Schema from model + surface, pick the next external driver, execute a step, collect violations and coverage, stop on max steps / violation / deadlock.

### Types

#### `SimulationConfig`

Run knobs: max steps, random seed, stop-on-violation, and optional surface include-list. Pass this into `NewSimulationEngine`.

#### `SimulationResult`

Outcome of a full run: steps, violations, termination reason, final state, schema, and coverage tracker. Feed this to `trace` for human/JSON output.

#### `SimulationEngine`

Orchestrates selector, step executor, checkers, and state for the whole run. Prefer this as the entry point rather than hand-wiring subcomponents.

#### `SimulationStep`

One recorded step (kind, target class/instance, event or query, parameters, cascaded peer steps, violations). Built by the step executor and stored on the result.

#### `StepKind`

What kind of top-level step this was (event transition, query, derived read, state do-action, and similar). Used in traces and coverage.

#### `PendingAction`

A candidate external driver the selector might fire next (class, instance or creation, event/query/derived, parameters). Selection samples among enabled pending actions.

#### `ActionSelector`

Chooses the next surface driver from Schema + live state (and RNG). Encodes “what is actually being exercised,” not every in-scope peer class.

#### `StepExecutor` / `StepExecutorDeps` / `StepParameterGenerator`

Runs one selected step end-to-end: bind/sample parameters, call actions, cascade peers, record the `SimulationStep`. Deps and the parameter generator keep construction and sampling testable.

#### `StateActionExecutor`

Runs entry/exit/do actions attached to states when a transition lands. Thin wrapper around `actions.ActionExecutor` for SM lifecycle hooks.

#### `CreationChainHandler`

Handles object creation chains (initial transitions, AC reify, required peer creation). Keeps create paths out of the generic transition body.

#### `DerivedAttributeEvaluator`

Evaluates external derived attributes when the selector picks a derived surface read. Uses Schema derived defs plus bindings/eval context.

#### `LivenessChecker`

Thin engine-side entry that asks `instance.State` to check liveness obligations after the run (or when requested). Does not own the obligation tables.

#### `SimulationCoverageTracker`

Records which parameter simulation specs produced values during the run. Helps testers see sampling coverage beyond step traces.

#### Surface report types

`SurfaceReport`, `SurfaceClassReport`, `SurfaceEventReport`, `SurfaceStateReport`, `SurfaceQueryReport`, `SurfaceDerivedAttributeReport`, `SurfaceActionReport`, `SurfaceAssocCreateNote`, `SurfaceUnavailableMemberReport`. These describe **drivers** and unavailable members for the human tester; they are narrower than full include-list scope.

---

## Package: `actions`

Executes one action, query, or transition against live state: bind parameters, evaluate guards and guarantees, apply primed updates, fire peer association effects, then ask instance for checks. Does not own the step loop.

### Types

#### `ActionExecutor`

Main executor for actions, queries, and transitions. Holds Schema, checkers, guard evaluator, and world-state deferral depth for nested peer work.

#### `InvariantRuntimeCheckers`

Pair of `InvariantChecker` + `DataTypeChecker` passed into the executor at construction. Keeps the constructor arg list readable.

#### `ActionResult` / `QueryResult` / `TransitionResult`

Outcomes of a single action, query, or SM transition (primed assignments, peer transitions, materialization, violations, success). Engine records these into steps/traces.

#### `AssociationMaterialization`

Records that an association-class creation produced a host link between two endpoints. Used for traces and liveness of association links.

#### `AssociationClassIndex`

Interface for “is this class an AC / host association?” lookups during creation linking. Implemented by Schema-backed indexes.

#### `ExecutionContext`

Carries deferred post-conditions, peer creations/updates, and safety rules collected while an action body runs. Flushed after the body so nested effects stay ordered.

#### `DeferredPostCondition` / `DeferredPeerCreation` / `DeferredPeerUpdate` / `DeferredSafetyRule`

Items queued on `ExecutionContext` until the primary mutation finishes. Prevents checking or cascading mid-expression.

#### `GuardEvaluator`

Evaluates transition/action guards against current bindings. Returns whether the guard holds so selection and execution can skip disabled paths.

#### `ParameterBinder`

Binds event/action parameters from model types, constraints, and live object extents. Used by sampling and by explicit parameter maps.

#### `ObjectInstanceLookup`

Function type that lists live objects for a class ref during parameter binding. Lets the binder stay free of a hard dependency on full State methods.

#### `ParameterSampler`

Samples legal parameter values under requires constraints (with exhaustion errors when the domain is empty). Powers non-deterministic surface events.

#### `ParameterOwner` / `RequireAssessmentFailure` / `ParameterSampleExhaustedError` / `UnsupportedRequiresSamplingError`

Support types for “who owns this parameter,” failed requires assessments, and sampling failure modes. Keep sampling errors distinct from model invariant violations.

#### `SurfaceEventSamplingContext`

Context for sampling parameters of a top-level surface event (state, schema, RNG, named sets). Bridges engine selection to actions sampling.

#### `PeerCreationCatalog`

Interface listing peer classes/events that may be created as side effects of association guarantees. Keeps cascade discovery model-driven.

#### `PeerTransitionRecord`

Records that a peer class event was fired as part of set-add/set-map (or similar) guarantees. Engine may expand these into cascaded steps.

#### `CreationLinkSource`

Describes where endpoint IDs came from when linking during creation (parameters, self, reify). Helps materialize host associations correctly.

#### `SamplingConstraintsForTest`

Test-facing constraint bag for parameter sampling experiments. Prefer production binding APIs outside tests.

---

## Package: `evaluator`

Evaluates TLA+ / logic expressions to `object.Object` values. Knows bindings, builtins, relations, and link tables; does not own simulation policy or violation packaging.

### Types

#### `EvalContext`

Evaluation environment: registry interface, builtins, and optional relation context. Created via Schema or pipeline setup for a run.

#### `EvalResult`

Value, primed bindings (`x' = …`), and optional error object from one evaluation. Callers apply primes to instances after a successful action body.

#### `Bindings` / `BindingEntry` / `Namespace`

Lexical environment for names (`self`, parameters, extents, enclosed scopes). Built per evaluation; prefer builders in `state` for full simulation bindings.

#### `BuiltinFn`

Function signature for TLA+ builtins registered on the eval context. Keep builtins pure with respect to world mutation.

#### `RelationContext` / `RelationInfo` / `Multiplicity`

Maps relation field names to association metadata and multiplicities for navigation expressions. Filled by `RelationContextBuilder` or instance projection.

#### `AssociationHostEndpoints` / `AssociationHostMultiplicities` / `InstanceEndpoint`

Helper shapes for host-association endpoint resolution during relation eval. Used when walking association-class materializations.

#### `LinkTable` / `Link` / `AssociationKey`

Runtime binary link storage used by evaluator and mirrored in `instance.State`. Keys associations; stores from/to instance object IDs.

#### `IdentityRegistry` / `ObjectID`

Maps between `object.Record` pointers and stable numeric object IDs for link identity. Keeps sets of instances comparable inside expressions.

#### `IRRegistryInterface`

Minimal interface evaluator needs from the definition registry (look up custom defs). Avoids importing the full registry graph into every eval path.

---

## Package: `object`

Runtime values as TLA+ sees them: numbers, booleans, strings, sets, bags, tuples, records, association relations, errors. Pure data with clone/inspect; not class instances by themselves.

### Types

#### `Object` (interface) / `ObjectType`

Common interface for every runtime value (`Type`, `Inspect`, `SetValue`, `Clone`). `ObjectType` is the string tag used for hashing and type checks.

#### `Number` / `NumberKind`

Numeric values (natural, integer, rational, float kinds). Use constructors `NewNatural`, `NewInteger`, `NewRational`, `NewFloat`; prefer coerce helpers when mixing model types.

#### `Boolean` / `String`

Simple scalar wrappers around Go `bool` and `string`. Construct with `NewBoolean` / `NewString`.

#### `Set` / `Bag` / `BagEntry`

Unordered unique collections (`Set`) and multisets (`Bag` with counts). Elements are other `Object`s hashed for membership.

#### `Tuple` / `Record`

Positional sequence (`Tuple`) and named fields (`Record`). Instance attributes are usually a `Record`.

#### `AssociationRelation`

Runtime value representing a relation/association navigation result for expressions. Distinct from link-table storage.

#### `Error`

Evaluated error value when expression evaluation fails. Surfaces as `EvalResult.Error` rather than a Go `error` alone.

---

## Package: `state`

Adapts `instance.State` into evaluator bindings and TLA extents. Thin bridge so `instance` stays free of full evaluator construction cycles.

### Types

#### Aliases

`InstanceID`, `ClassInstance`, `SimulationState`, `AssociationLink`, `AssociationLinkTable` re-export instance types for older call sites. Prefer importing `instance` directly in new code.

#### `BindingsBuilder`

Builds `evaluator.Bindings` (and related extent/relation data) from live state for a given `self` instance and parameters. Main entry for action and guard evaluation setup.

#### `DerivedAttributeResolver`

Interface to resolve derived attribute values when building bindings. Lets the builder request derived reads without owning evaluation policy.

---

## Package: `surface`

Include-list scope for a run and reports of what is loaded vs what is driven. Resolves domain/subdomain/class includes and finds members that cannot be surface-driven.

### Types

#### `SurfaceSpecification`

Include/exclude lists for domains, subdomains, and classes. Empty means “simulate everything”; otherwise only included classes are in run scope.

#### `ResolvedSurface`

Concrete set of class keys (and related resolution data) after applying the specification to a model. Fed into `schema.RunScope`.

#### `ScopeKind` / `ScopeEntry`

Human-facing scope summary: whole subdomain path when fully included, or individual class paths when partial. Shown to testers separately from surface drivers.

#### `MemberKind` / `UnavailableMember`

A class member (event, query, derived attribute, …) that cannot be used as a top-level driver, with reason. Schema stores these for surface reports.

#### `Diagnostic` / `CallerData`

Diagnostics about surface resolution and call graphs (who could call whom). Useful when a member is off-surface because dependencies are out of scope.

---

## Package: `model_bridge`

Loads and compiles TLA+ expressions from the core model into forms the registry/evaluator can run. Setup path before or while Schema/registry are populated.

### Types

#### `Loader` / `LoadResult`

Walks model logic and loads extractable expressions into intermediate results. Entry point for “make expressions evaluable.”

#### `DefinitionBuilder` / `BuildResult` / `GuaranteeKind`

Turns extracted expressions into registry `Definition`s (invariants, requires, guarantees, and similar kinds). Bridge between model AST and `registry`.

#### `ExpressionSource` / `ExtractedExpression`

Where an expression came from in the model and the extracted payload (source text, parsed form, location). Used for diagnostics when evaluation fails.

---

## Package: `registry`

Scoped store of custom TLA+ definitions (operators, invariants, guarantees) with invalidation and rebuild. Evaluator resolves names through a runtime adapter.

### Types

#### `Registry`

Map of definitions by key, with registration and lookup. Owns the definition graph for the run’s custom logic.

#### `Definition` / `DefinitionKey` / `DefinitionKind` / `Parameter` / `ScopePath`

One registered definition and its identity, kind, parameters, and scope path (global vs class-scoped). Keys must stay stable for rebuild/invalidation.

#### `ScopeContext`

Resolution scope when looking up names (global or a specific class path). Passed into evaluation and rebuild validation.

#### `ScopeLevel`

How deep a scope is (global vs class). Affects which definitions are visible.

#### `RuntimeAdapter`

Adapts `Registry` to the evaluator’s `IRRegistryInterface`. Keeps evaluator independent of registry internals.

#### `InvalidationSet`

Tracks which definitions must be rebuilt after model or dependency changes. Used by rebuild strategies during load.

#### `RebuildStrategy` / `ValidateFunc` / `RebuildError` / `DefinitionError`

Controls and errors for recompiling definitions after invalidation. Failures here are load-time problems, not simulation invariant violations.

---

## Package: `report`

Formats and categorizes `instance.ViolationErrors` for human-readable output. Does not detect violations; it only presents them.

### Types

#### `ViolationReport`

Top-level report structure: categories of entries ready for CLI or JSON. Built from a slice of `ViolationError`.

#### `ViolationCategory`

One bucket of violations (for example multiplicity vs liveness). Groups related entries under a heading.

#### `ViolationEntry`

One display row: message and identity fields for a single violation. Mirrors `ViolationError` without exposing constructors.

---

## Package: `trace`

Turns `engine.SimulationResult` into serializable step traces and final-state snapshots for text/JSON. View-model only; no simulation logic.

### Types

#### `SimulationTrace`

Top-level trace: steps taken, termination reason, step list, optional final state. Built with `FromResult`.

#### `TraceStep`

One step’s serializable fields (kind, class, event/query/derived, parameters, assignments, cascades, violations). Nested `CascadedSteps` hold peer effects.

#### `AssociationMaterializationTrace`

Serializable host-link endpoints created via an association class. Appears on create steps and sometimes on final instance endpoints.

#### `FinalState` / `InstanceState`

End-of-run instance dump: counts, attribute maps, optional association endpoints. For testers inspecting the last world snapshot.

---

## Package: `types`

TLA+ **type system** values (not runtime `object` values): Booleans, Numbers, Sets, type variables, schemes. Used for typechecking and type_spec reasoning before or beside evaluation.

### Types

#### `Type` (interface)

Any TLA+ type with `String`, `Equals`, free-variable collection, and substitution. The root of the type algebra in this package.

#### `Substitution`

Map from type-variable IDs to types for Hindley–Milner style inference. Apply/compose when unifying types.

#### Concrete types

`Boolean`, `Number`, `String`, `Set`, `Tuple`, `Record`, `Bag`, `Function`, `Any` — structural TLA+ types. Distinct from `object.*` runtime wrappers with the same English names.

#### `TypeVar` / `Scheme`

Type variables and polymorphic schemes (`∀a. T`) for let-polymorphism. Used by inference, not by the step loop.

---

## Quick “where do I change X?”

| Goal | Start here |
| --- | --- |
| Include only some classes | `surface.SurfaceSpecification` → Schema `RunScope` |
| Look up model facts for the run | `schema.Schema` |
| Create/update instances or links | `instance.State` |
| Detect / package a violation | `instance` Check* / checkers (not `report`) |
| Run one event/action/query | `actions.ActionExecutor` |
| Choose next external driver / run N steps | `engine.SimulationEngine` |
| Evaluate a TLA expression | `evaluator` + `state.BindingsBuilder` + `object` |
| Pretty-print the run | `trace` then CLI formatting; violations via `report` |

---

## Mental model reminders

- **Scope** = classes loaded for the run. **Surface drivers** = external hooks the selector may fire. Peers and association classes can be in scope without being listed as drivers.
- **Schema is static; instance is mutable.** If you need “what does the model say?” ask Schema. If you need “what is true right now?” ask State.
- **Never assemble `ViolationError` outside `instance`.** Call a Check or Package API and return what it gives you.
- **Production code is model-agnostic.** Use keys, names, and multiplicities from Schema/model data, not hard-coded domain strings.
)
