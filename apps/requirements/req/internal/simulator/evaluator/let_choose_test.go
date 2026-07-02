package evaluator_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/registry"
	"github.com/stretchr/testify/require"
)

func TestEvalRecursiveSumGlobalFunction(t *testing.T) {
	gfKey, err := identity.NewGlobalFunctionKey("_sumadjustmentamounts")
	require.NoError(t, err)

	spec := `IF amounts = {} THEN 0 ELSE LET x == CHOOSE y \in amounts : TRUE IN x + _SumAdjustmentAmounts(amounts \ {x})`
	astExpr, err := parser.ParseExpression(spec)
	require.NoError(t, err)

	ctx := &convert.LowerContext{
		Parameters: map[string]bool{"amounts": true},
		GlobalFunctions: map[string]identity.Key{
			"_SumAdjustmentAmounts": gfKey,
		},
	}
	body, err := convert.Lower(astExpr, ctx)
	require.NoError(t, err)

	reg := registry.NewRegistry()
	_, err = reg.RegisterGlobalFunction("_SumAdjustmentAmounts", body, []registry.Parameter{{Name: "amounts"}})
	require.NoError(t, err)

	adapter := registry.NewRuntimeAdapter(reg)
	evaluator.SetEvalContext(&evaluator.EvalContext{IRRegistry: adapter})

	amounts := object.NewSetFromElements([]object.Object{
		object.NewInteger(3),
		object.NewInteger(1),
		object.NewInteger(2),
	})
	bindings := evaluator.NewBindings()
	bindings.Set("amounts", amounts, evaluator.NamespaceLocal)

	callAST, err := parser.ParseExpression(`_SumAdjustmentAmounts(amounts)`)
	require.NoError(t, err)
	callCtx := &convert.LowerContext{
		Parameters: map[string]bool{"amounts": true},
		GlobalFunctions: map[string]identity.Key{
			"_SumAdjustmentAmounts": gfKey,
		},
	}
	callExpr, err := convert.Lower(callAST, callCtx)
	require.NoError(t, err)

	result := evaluator.Eval(callExpr, bindings)
	require.False(t, result.IsError())
	sum, ok := result.Value.(*object.Number)
	require.True(t, ok)
	require.Equal(t, 0, sum.Cmp(object.NewInteger(6)))
}
