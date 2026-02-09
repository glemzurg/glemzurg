package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
)

// evalBooleanLiteral evaluates a boolean literal.
func evalBooleanLiteral(node *ast.BooleanLiteral) *EvalResult {
	return NewEvalResult(nativeBoolToBoolean(node.Value))
}
