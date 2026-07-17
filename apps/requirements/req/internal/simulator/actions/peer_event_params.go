package actions

import (
	"fmt"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// resolvePositionalEventCallParams maps EventCall arguments to event parameter names
// by declaration order. The bound set-map variable (LocalVar only) is skipped; remaining
// args are evaluated (LocalVar or FieldAccess etc.) and bound to paramNames by position.
func resolvePositionalEventCallParams(
	boundVar string,
	paramNames []string,
	eventCall *me.EventCall,
	bindings *evaluator.Bindings,
) (map[string]object.Object, error) {
	valueArgs := nonBoundEventCallArgs(boundVar, eventCall)
	if len(valueArgs) != len(paramNames) {
		return nil, fmt.Errorf(
			"event call supplies %d arguments but event declares %d parameters",
			len(valueArgs), len(paramNames),
		)
	}

	params := make(map[string]object.Object, len(paramNames))
	for i, paramName := range paramNames {
		result := evaluator.Eval(valueArgs[i], bindings)
		if result.IsError() {
			return nil, fmt.Errorf("event argument %d for parameter %q: %s", i, paramName, result.Error.Inspect())
		}
		params[paramName] = result.Value
	}
	return params, nil
}

// nonBoundEventCallArgs drops a single LocalVar matching boundVar (set-map row variable).
// All other expressions (including FieldAccess like r.amount) are kept for evaluation.
func nonBoundEventCallArgs(boundVar string, eventCall *me.EventCall) []me.Expression {
	if eventCall == nil {
		return nil
	}
	valueArgs := make([]me.Expression, 0, len(eventCall.Args))
	for _, arg := range eventCall.Args {
		if boundVar != "" {
			if lv, ok := arg.(*me.LocalVar); ok && lv.Name == boundVar {
				continue
			}
		}
		valueArgs = append(valueArgs, arg)
	}
	return valueArgs
}
