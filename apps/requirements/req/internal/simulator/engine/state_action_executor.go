package engine

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// StateActionExecutor executes entry/exit/do StateActions around transitions.
type StateActionExecutor struct {
	actionExecutor *actions.ActionExecutor
}

// NewStateActionExecutor creates a new StateActionExecutor.
func NewStateActionExecutor(actionExecutor *actions.ActionExecutor) *StateActionExecutor {
	return &StateActionExecutor{
		actionExecutor: actionExecutor,
	}
}

// ExecuteExitActions runs all exit StateActions for the state being left.
func (e *StateActionExecutor) ExecuteExitActions(
	class model_class.Class,
	fromStateKey identity.Key,
	instance *state.ClassInstance,
) (invariants.ViolationList, error) {
	return e.executeStateActions(class, fromStateKey, instance, "exit")
}

// ExecuteEntryActions runs all entry StateActions for the state being entered.
func (e *StateActionExecutor) ExecuteEntryActions(
	class model_class.Class,
	toStateKey identity.Key,
	instance *state.ClassInstance,
) (invariants.ViolationList, error) {
	return e.executeStateActions(class, toStateKey, instance, "entry")
}

func (e *StateActionExecutor) executeStateActions(
	class model_class.Class,
	stateKey identity.Key,
	instance *state.ClassInstance,
	when string,
) (invariants.ViolationList, error) {
	s, ok := class.States[stateKey]
	if !ok {
		return nil, fmt.Errorf("state %s not found in class %s", stateKey.String(), class.Name)
	}

	var allViolations invariants.ViolationList

	for _, sa := range s.Actions {
		if sa.When != when {
			continue
		}

		action, ok := class.Actions[sa.ActionKey]
		if !ok {
			return nil, fmt.Errorf("state action references non-existent action %s in class %s", sa.ActionKey.String(), class.Name)
		}

		result, err := e.actionExecutor.ExecuteAction(action, instance, nil)
		if err != nil {
			return allViolations, fmt.Errorf("state %s action %s error: %w", when, action.Name, err)
		}

		allViolations = append(allViolations, result.Violations...)
	}

	return allViolations, nil
}
