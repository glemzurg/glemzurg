// Package schema is the sole home of model facts, run scope, and static simulator
// indexes for one simulation run.
//
// Data-flow gate:
//
//	full *core.Model + RunScope ──New──► *Schema ──► instance, engine, actions, …
//
// After construction, the running simulator must not carry a separate *core.Model
// or include-list as authority for the same run. The private catalog implementation
// is not exported; callers use *Schema only.
//
// Schema presents:
//
//   - Keyed lookups: Class / Association / ClassSim as (*T, inScope bool, err)
//     with out-of-scope = (nil, false, nil)
//   - Association graph: ScopedAssociations, HostAssociationForAC, uniqueness lists
//   - Class simulation metadata: ClassSim, EachInScopeClassSim, …
//   - Attribute/derived projections: ClassIndexes, DerivedAttributes, …
//   - Association TLA navigation, peer events, caller graphs, external surface
//     membership, extent names, surface-unavailable members (all via Schema methods)
//   - Factories: NewEvalContext
//
// Engine and actions ask Schema questions; they must not rebuild parallel static
// catalogs. Runtime Check* and step execution stay outside this package.
//
// [instance.State] holds *Schema for static lookups; mutable world state stays in
// instance. Do not mutate the model after New.
package schema
