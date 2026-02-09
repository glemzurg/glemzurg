package evaluator

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/typechecker"
	"github.com/stretchr/testify/assert"
)

func TestEvalTypedTupleLiteral(t *testing.T) {
	node := &ast.TupleLiteral{
		Elements: []ast.Expression{
			ast.NewIntLiteral(1),
			ast.NewIntLiteral(2),
			ast.NewIntLiteral(3),
		},
	}

	tc := typechecker.NewTypeChecker()
	typed, err := tc.Check(node)

	assert.NoError(t, err)
	assert.NotNil(t, typed)
	assert.Len(t, typed.Children, 3)

	bindings := NewBindings()
	result := EvalTyped(typed, bindings)

	assert.False(t, result.IsError())
	tuple := result.Value.(*object.Tuple)
	assert.Equal(t, 3, tuple.Len())
}
