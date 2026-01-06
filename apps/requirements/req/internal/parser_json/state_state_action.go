package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/state"

// stateActionInOut is a action that triggers when a state is entered or exited or happens perpetually.
type stateActionInOut struct {
	Key       string `json:"key"`
	ActionKey string `json:"action_key"`
	When      string `json:"when"`
}

// ToRequirements converts the stateActionInOut to state.StateAction.
func (s stateActionInOut) ToRequirements() state.StateAction {
	return state.StateAction{
		Key:       s.Key,
		ActionKey: s.ActionKey,
		When:      s.When,
	}
}

// FromRequirements creates a stateActionInOut from state.StateAction.
func FromRequirementsStateAction(s state.StateAction) stateActionInOut {
	return stateActionInOut{
		Key:       s.Key,
		ActionKey: s.ActionKey,
		When:      s.When,
	}
}
