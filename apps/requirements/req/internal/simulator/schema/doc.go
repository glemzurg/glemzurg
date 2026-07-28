// Package schema is the sole home of model facts, run scope, and static simulator
// catalogs for one simulation run.
//
// Data-flow gate:
//
//	full *core.Model + RunScope ──New──► *Schema (+ Catalog) ──► instance, engine, actions, …
//
// After construction, the running simulator must not carry a separate *core.Model
// or include-list as authority for the same run. Schema presents:
//
//   - Keyed lookups: Class / Association / ClassSim as (*T, inScope bool, err)
//     with out-of-scope = (nil, false, nil)
//   - Association graph: ScopedAssociations, HostAssociationForAC, uniqueness lists
//   - Class simulation metadata: ClassSim, EachInScopeClassSim, …
//   - Attribute/derived projections: ClassIndexes, DerivedAttributes, …
//   - Catalog (via Catalog()): association TLA navigation, peer events, caller graphs,
//     external surface membership, extent names, surface-unavailable members
//   - Factories: NewEvalContext
//
// Engine and actions should ask schema/Catalog questions; they must not rebuild
// parallel static catalogs. Runtime Check* and step execution stay outside this package.
//
// [instance.State] holds *Schema for static lookups; mutable world state stays in
// instance. Do not mutate the model after New.
package schema
