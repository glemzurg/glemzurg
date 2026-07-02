package actions

import (
	"fmt"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// resolvePositionalEventCallParams maps EventCall arguments to event parameter names
// by declaration order. The bound set-map variable is skipped; remaining args bind to
// paramNames[0], paramNames[1], … by position so reversed TLA+ argument order is honored.
func resolvePositionalEventCallParams(
	boundVar string,
	paramNames []string,
	eventCall *me.EventCall,
	bindings *evaluator.Bindings,
) (map[string]object.Object, error) {
	valueArgs, err := nonBoundEventCallArgs(boundVar, eventCall)
	if err != nil {
		return nil, err
	}
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

func nonBoundEventCallArgs(boundVar string, eventCall *me.EventCall) ([]me.Expression, error) {
	valueArgs := make([]me.Expression, 0, len(eventCall.Args))
	for i, arg := range eventCall.Args {
		name, ok := eventCallArgName(arg)
		if !ok {
			return nil, fmt.Errorf("event arg[%d]: expected parameter reference", i)
		}
		if name == boundVar {
			continue
		}
		valueArgs = append(valueArgs, arg)
	}
	return valueArgs, nil
}

func eventCallArgName(arg me.Expression) (string, bool) {
	switch a := arg.(type) {
	case *me.LocalVar:
		return a.Name, true
	default:
		return "", false
	}
}
