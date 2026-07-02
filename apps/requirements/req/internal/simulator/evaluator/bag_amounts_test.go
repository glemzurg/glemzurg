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

func TestEvalSumAmountsBagCountsDuplicates(t *testing.T) {
	amountsBagKey, err := identity.NewGlobalFunctionKey("_amountsbag")
	require.NoError(t, err)
	sumAmountsKey, err := identity.NewGlobalFunctionKey("_sumamounts")
	require.NoError(t, err)

	amountsBagSpec := `IF adjustments = {} THEN _Bags!SetToBag({}) ELSE LET c == CHOOSE x \in adjustments : TRUE IN _Bags!SetToBag({c.amount}) ⊕ _AmountsBag(adjustments \ {c})`
	sumAmountsSpec := `IF _Bags!BagToSet(amounts) = {} THEN 0 ELSE LET x == CHOOSE y \in _Bags!BagToSet(amounts) : _Bags!CopiesIn(y, amounts) > 0 IN x + _SumAmounts(amounts ⊖ _Bags!SetToBag({x}))`

	amountsBagAST, err := parser.ParseExpression(amountsBagSpec)
	require.NoError(t, err)
	sumAmountsAST, err := parser.ParseExpression(sumAmountsSpec)
	require.NoError(t, err)

	lowerCtx := &convert.LowerContext{
		GlobalFunctions: map[string]identity.Key{
			"_AmountsBag": amountsBagKey,
			"_SumAmounts": sumAmountsKey,
		},
	}

	amountsBagCtx := &convert.LowerContext{
		Parameters:      map[string]bool{"adjustments": true},
		GlobalFunctions: lowerCtx.GlobalFunctions,
	}
	sumAmountsCtx := &convert.LowerContext{
		Parameters:      map[string]bool{"amounts": true},
		GlobalFunctions: lowerCtx.GlobalFunctions,
	}

	amountsBagBody, err := convert.Lower(amountsBagAST, amountsBagCtx)
	require.NoError(t, err)
	sumAmountsBody, err := convert.Lower(sumAmountsAST, sumAmountsCtx)
	require.NoError(t, err)

	reg := registry.NewRegistry()
	_, err = reg.RegisterGlobalFunction("_AmountsBag", amountsBagBody, []registry.Parameter{{Name: "adjustments"}})
	require.NoError(t, err)
	_, err = reg.RegisterGlobalFunction("_SumAmounts", sumAmountsBody, []registry.Parameter{{Name: "amounts"}})
	require.NoError(t, err)

	adapter := registry.NewRuntimeAdapter(reg)
	evaluator.SetEvalContext(&evaluator.EvalContext{IRRegistry: adapter})

	// Two distinct adjustments with the same amount must both count toward the sum.
	adj1 := object.NewRecordFromFields(map[string]object.Object{
		"id":     object.NewInteger(1),
		"amount": object.NewInteger(5),
	})
	adj2 := object.NewRecordFromFields(map[string]object.Object{
		"id":     object.NewInteger(2),
		"amount": object.NewInteger(5),
	})
	adjustments := object.NewSetFromElements([]object.Object{adj1, adj2})

	bindings := evaluator.NewBindings()
	bindings.Set("adjustments", adjustments, evaluator.NamespaceLocal)

	amountsBagCallAST, err := parser.ParseExpression(`_AmountsBag(adjustments)`)
	require.NoError(t, err)
	amountsBagCall, err := convert.Lower(amountsBagCallAST, amountsBagCtx)
	require.NoError(t, err)

	bagResult := evaluator.Eval(amountsBagCall, bindings)
	require.False(t, bagResult.IsError())

	sumBindings := evaluator.NewBindings()
	sumBindings.Set("amounts", bagResult.Value, evaluator.NamespaceLocal)
	sumCallAST, err := parser.ParseExpression(`_SumAmounts(amounts)`)
	require.NoError(t, err)
	sumCall, err := convert.Lower(sumCallAST, sumAmountsCtx)
	require.NoError(t, err)

	sumResult := evaluator.Eval(sumCall, sumBindings)
	if sumResult.IsError() {
		t.Fatalf("sum error: %s", sumResult.Error.Message)
	}
	require.Equal(t, object.TypeBag, bagResult.Value.Type(), "bag: %s", bagResult.Value.Inspect())
	sum, ok := sumResult.Value.(*object.Number)
	require.True(t, ok, "sum type: %s", sumResult.Value.Inspect())
	require.Equal(t, 0, sum.Cmp(object.NewInteger(10)), "sum=%s bag=%s", sum.Inspect(), bagResult.Value.Inspect())
}
