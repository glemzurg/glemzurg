package parser_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	"github.com/stretchr/testify/require"
)

func TestParseLetExpr(t *testing.T) {
	expr, err := parser.ParseExpression(`LET x == 1 IN x + 1`)
	require.NoError(t, err)
	let, ok := expr.(*ast.LetExpr)
	require.True(t, ok)
	require.Equal(t, "x", let.Variable)
}

func TestParseChooseExpr(t *testing.T) {
	expr, err := parser.ParseExpression(`CHOOSE x \in {3, 1, 2} : TRUE`)
	require.NoError(t, err)
	_, ok := expr.(*ast.ChooseExpr)
	require.True(t, ok)
}

func TestParseRecursiveSumSpec(t *testing.T) {
	spec := `IF amounts = {} THEN 0 ELSE LET x == CHOOSE y \in amounts : TRUE IN x + _SumAdjustmentAmounts(amounts \ {x})`
	_, err := parser.ParseExpression(spec)
	require.NoError(t, err)
}
