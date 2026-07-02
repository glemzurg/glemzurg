package evaluator_test

import (
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestEvalBagCardinalitySetAndBag(t *testing.T) {
	set := object.NewSetFromElements([]object.Object{object.NewInteger(1), object.NewInteger(2)})
	setCall := &me.BuiltinCall{Module: "_Bags", Function: "BagCardinality", Args: []me.Expression{
		&me.LocalVar{Name: "s"},
	}}
	bindings := evaluator.NewBindings()
	bindings.Set("s", set, evaluator.NamespaceLocal)
	setResult := evaluator.Eval(setCall, bindings)
	require.False(t, setResult.IsError(), setResult.Error)
	count, ok := setResult.Value.(*object.Number)
	require.True(t, ok)
	require.Equal(t, 0, count.Cmp(object.NewInteger(2)))

	bag := object.NewBag()
	bag.Add(object.NewInteger(5), 2)
	bag.Add(object.NewInteger(7), 1)
	bindings.Set("s", bag, evaluator.NamespaceLocal)
	bagResult := evaluator.Eval(setCall, bindings)
	require.False(t, bagResult.IsError(), bagResult.Error)
	total, ok := bagResult.Value.(*object.Number)
	require.True(t, ok)
	require.Equal(t, 0, total.Cmp(object.NewInteger(3)))
}
