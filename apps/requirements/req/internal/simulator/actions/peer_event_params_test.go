package actions

import (
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestResolvePositionalEventCallParamsBindsByOrderNotName(t *testing.T) {
	bindings := evaluator.NewBindings()
	bindings.Set("MinimumBalance", object.NewInteger(100), evaluator.NamespaceLocal)
	bindings.Set("TopoffBalance", object.NewInteger(200), evaluator.NamespaceLocal)

	eventCall := &me.EventCall{
		Args: []me.Expression{
			&me.LocalVar{Name: "b"},
			&me.LocalVar{Name: "TopoffBalance"},
			&me.LocalVar{Name: "MinimumBalance"},
		},
	}

	params, err := resolvePositionalEventCallParams(
		"b",
		[]string{"MinimumBalance", "TopoffBalance"},
		eventCall,
		bindings,
	)
	require.NoError(t, err)
	require.Equal(t, "200", params["MinimumBalance"].Inspect())
	require.Equal(t, "100", params["TopoffBalance"].Inspect())
}

func TestResolvePositionalEventCallParamsRejectsArgCountMismatch(t *testing.T) {
	eventCall := &me.EventCall{
		Args: []me.Expression{&me.LocalVar{Name: "MinimumBalance"}},
	}
	_, err := resolvePositionalEventCallParams(
		"",
		[]string{"MinimumBalance", "TopoffBalance"},
		eventCall,
		evaluator.NewBindings(),
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "supplies 1 arguments")
}
