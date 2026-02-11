// Package model_bridge extracts TLA+ expressions from req_model and compiles them
// for use by the simulator.
package model_bridge

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

// ExpressionSource identifies where a TLA+ expression originates in the model.
type ExpressionSource int

const (
	// SourceModelInvariant is a model-level invariant from Model.Invariants.
	SourceModelInvariant ExpressionSource = iota
	// SourceTlaDefinition is a global TLA+ definition from Model.TlaDefinitions.
	SourceTlaDefinition
	// SourceActionRequires is an action precondition from Action.Requires.
	SourceActionRequires
	// SourceActionGuarantees is an action postcondition from Action.Guarantees.
	SourceActionGuarantees
	// SourceQueryRequires is a query precondition from Query.Requires.
	SourceQueryRequires
	// SourceQueryGuarantees is a query filtering criterion from Query.Guarantees.
	SourceQueryGuarantees
	// SourceGuardCondition is a guard condition from Guard.Logic.
	SourceGuardCondition
)

// String returns a human-readable name for the expression source.
func (s ExpressionSource) String() string {
	switch s {
	case SourceModelInvariant:
		return "model_invariant"
	case SourceTlaDefinition:
		return "tla_definition"
	case SourceActionRequires:
		return "action_requires"
	case SourceActionGuarantees:
		return "action_guarantees"
	case SourceQueryRequires:
		return "query_requires"
	case SourceQueryGuarantees:
		return "query_guarantees"
	case SourceGuardCondition:
		return "guard_condition"
	default:
		return "unknown"
	}
}

// ExtractedExpression represents a TLA+ expression extracted from the model
// along with its source location and scope information.
type ExtractedExpression struct {
	// Source indicates where this expression comes from (invariant, action, etc.)
	Source ExpressionSource

	// Expression is the raw TLA+ string.
	Expression string

	// ScopeKey is the identity.Key of the containing entity.
	// For model invariants and TLA definitions: nil (global scope)
	// For actions/queries/guards: pointer to the action/query/guard's key
	ScopeKey *identity.Key

	// Name is the name of the definition/action/query/guard.
	// For model invariants: empty string
	// For TLA definitions: the definition name (e.g., "_Max")
	// For actions: the action name
	// For queries: the query name
	// For guards: the guard name
	Name string

	// Parameters contains the parameter names for TLA definitions.
	// Only populated for SourceTlaDefinition.
	Parameters []string

	// Index is the position in the source array (e.g., which Requires entry).
	// For TLA definitions: always 0
	Index int
}
