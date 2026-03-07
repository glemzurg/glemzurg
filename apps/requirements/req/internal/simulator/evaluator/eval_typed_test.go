package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/typechecker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	require.NoError(t, err)
	assert.NotNil(t, typed)
	assert.Len(t, typed.Children, 3)

	bindings := NewBindings()
	result := EvalTyped(typed, bindings)

	assert.False(t, result.IsError())
	tuple := result.Value.(*object.Tuple)
	assert.Equal(t, 3, tuple.Len())
}
