package model_bridge

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
)

// ExtractFromModel extracts all TLA+ expressions from the model.
// Returns a slice of ExtractedExpression containing:
//   - Model invariants (TlaInvariants)
//   - Global function definitions (GlobalFunctions)
//   - Action requires/guarantees (TlaRequires, TlaGuarantees)
//   - Query requires/guarantees (TlaRequires, TlaGuarantees)
//   - Guard conditions (TlaGuard)
func ExtractFromModel(model *req_model.Model) []ExtractedExpression {
	var expressions []ExtractedExpression

	// Extract model-level invariants (global scope, no key)
	expressions = append(expressions, extractModelInvariants(model)...)

	// Extract global function definitions
	expressions = append(expressions, extractGlobalFunctions(model)...)

	// Extract from all domains -> subdomains -> classes
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				// Extract from actions
				for _, action := range class.Actions {
					expressions = append(expressions, extractActionExpressions(&action)...)
				}
				// Extract from queries
				for _, query := range class.Queries {
					expressions = append(expressions, extractQueryExpressions(&query)...)
				}
				// Extract from guards
				for _, guard := range class.Guards {
					expressions = append(expressions, extractGuardExpressions(&guard)...)
				}
			}
		}
	}

	return expressions
}

// extractModelInvariants extracts TLA+ invariants from the model level.
func extractModelInvariants(model *req_model.Model) []ExtractedExpression {
	expressions := make([]ExtractedExpression, 0, len(model.TlaInvariants))

	for i, tla := range model.TlaInvariants {
		expressions = append(expressions, ExtractedExpression{
			Source:     SourceModelInvariant,
			Expression: tla,
			ScopeKey:   nil, // Model invariants have global scope
			Name:       "",
			Index:      i,
		})
	}

	return expressions
}

// extractGlobalFunctions extracts global function definitions from the model.
func extractGlobalFunctions(model *req_model.Model) []ExtractedExpression {
	expressions := make([]ExtractedExpression, 0, len(model.GlobalFunctions))

	for _, gf := range model.GlobalFunctions {
		expressions = append(expressions, ExtractedExpression{
			Source:     SourceTlaDefinition,
			Expression: gf.Specification.Specification,
			ScopeKey:   nil, // Global functions have global scope
			Name:       gf.Name,
			Parameters: gf.Parameters,
			Index:      0,
		})
	}

	return expressions
}

// extractActionExpressions extracts TLA+ requires and guarantees from an action.
func extractActionExpressions(action *model_state.Action) []ExtractedExpression {
	expressions := make([]ExtractedExpression, 0, len(action.TlaRequires)+len(action.TlaGuarantees))

	// Extract TlaRequires
	for i, tla := range action.TlaRequires {
		key := action.Key // Copy to get addressable value
		expressions = append(expressions, ExtractedExpression{
			Source:     SourceActionRequires,
			Expression: tla,
			ScopeKey:   &key,
			Name:       action.Name,
			Index:      i,
		})
	}

	// Extract TlaGuarantees
	for i, tla := range action.TlaGuarantees {
		key := action.Key // Copy to get addressable value
		expressions = append(expressions, ExtractedExpression{
			Source:     SourceActionGuarantees,
			Expression: tla,
			ScopeKey:   &key,
			Name:       action.Name,
			Index:      i,
		})
	}

	return expressions
}

// extractQueryExpressions extracts TLA+ requires and guarantees from a query.
func extractQueryExpressions(query *model_state.Query) []ExtractedExpression {
	expressions := make([]ExtractedExpression, 0, len(query.TlaRequires)+len(query.TlaGuarantees))

	// Extract TlaRequires
	for i, tla := range query.TlaRequires {
		key := query.Key // Copy to get addressable value
		expressions = append(expressions, ExtractedExpression{
			Source:     SourceQueryRequires,
			Expression: tla,
			ScopeKey:   &key,
			Name:       query.Name,
			Index:      i,
		})
	}

	// Extract TlaGuarantees
	for i, tla := range query.TlaGuarantees {
		key := query.Key // Copy to get addressable value
		expressions = append(expressions, ExtractedExpression{
			Source:     SourceQueryGuarantees,
			Expression: tla,
			ScopeKey:   &key,
			Name:       query.Name,
			Index:      i,
		})
	}

	return expressions
}

// extractGuardExpressions extracts TLA+ guard conditions from a guard.
func extractGuardExpressions(guard *model_state.Guard) []ExtractedExpression {
	expressions := make([]ExtractedExpression, 0, len(guard.TlaGuard))

	for i, tla := range guard.TlaGuard {
		key := guard.Key // Copy to get addressable value
		expressions = append(expressions, ExtractedExpression{
			Source:     SourceGuardCondition,
			Expression: tla,
			ScopeKey:   &key,
			Name:       guard.Name,
			Index:      i,
		})
	}

	return expressions
}
