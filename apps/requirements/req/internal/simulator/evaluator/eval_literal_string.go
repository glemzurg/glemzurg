package evaluator

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// evalStringLiteral evaluates a string literal.
func evalStringLiteral(node *ast.StringLiteral) *EvalResult {
	return NewEvalResult(object.NewString(node.Value))
}
