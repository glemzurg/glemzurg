package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalStringLiteral evaluates a string literal.
func evalStringLiteral(node *ast.StringLiteral) *EvalResult {
	return NewEvalResult(object.NewString(node.Value))
}
