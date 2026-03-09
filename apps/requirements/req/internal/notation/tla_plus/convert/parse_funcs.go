package convert

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
)

// NewExpressionParseFunc creates an ExpressionParseFunc that uses the TLA+ parser
// and lowering pipeline with the given context. If ctx is nil, an empty context
// is used (suitable for context-free expressions like literals and arithmetic).
// Returns (expression, normalizedTLA) on success, (nil, "") on any failure.
func NewExpressionParseFunc(ctx *LowerContext) logic_spec.ExpressionParseFunc {
	if ctx == nil {
		ctx = &LowerContext{}
	}
	return func(specification string) (me.Expression, string) {
		// Parse TLA+ text to AST.
		astExpr, err := parser.ParseExpression(specification)
		if err != nil {
			return nil, ""
		}
		// Lower AST to model expression.
		expr, err := Lower(astExpr, ctx)
		if err != nil {
			return nil, ""
		}
		// Round-trip: raise back to TLA+ for normalized form.
		raisedAST, err := Raise(expr, raiseContextFromLower(ctx))
		if err != nil {
			// Lowering succeeded but raising failed — keep the expression
			// with the original specification text.
			return expr, ""
		}
		return expr, ast.Print(raisedAST)
	}
}

// raiseContextFromLower creates a RaiseContext by inverting the name maps
// from a LowerContext. LowerContext maps name→key, RaiseContext maps key→name.
func raiseContextFromLower(ctx *LowerContext) *RaiseContext {
	if ctx == nil {
		return &RaiseContext{}
	}
	rc := &RaiseContext{
		AttributeNames:  invertMap(ctx.AttributeNames),
		ActionNames:     invertMap(ctx.ActionNames),
		QueryNames:      invertMap(ctx.QueryNames),
		GlobalFunctions: invertMap(ctx.GlobalFunctions),
		NamedSets:       invertMap(ctx.NamedSets),
	}
	// AllActions → ActionScopePaths (scoped names, not simple names).
	if ctx.AllActions != nil {
		rc.ActionScopePaths = invertMap(ctx.AllActions)
	}
	return rc
}

// invertMap inverts a map[string]identity.Key to map[identity.Key]string.
func invertMap(m map[string]identity.Key) map[identity.Key]string {
	if m == nil {
		return nil
	}
	result := make(map[identity.Key]string, len(m))
	for name, key := range m {
		result[key] = name
	}
	return result
}
