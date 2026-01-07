package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"

// stateActionInOut is a action that triggers when a state is entered or exited or happens perpetually.
type stateActionInOut struct {
	Key       string `json:"key"`
	ActionKey string `json:"action_key"`
	When      string `json:"when"`
}

// ToRequirements converts the stateActionInOut to model_state.StateAction.
func (s stateActionInOut) ToRequirements() model_state.StateAction {
	return model_state.StateAction{
		Key:       s.Key,
		ActionKey: s.ActionKey,
		When:      s.When,
	}
}

// FromRequirements creates a stateActionInOut from model_state.StateAction.
func FromRequirementsStateAction(s model_state.StateAction) stateActionInOut {
	return stateActionInOut{
		Key:       s.Key,
		ActionKey: s.ActionKey,
		When:      s.When,
	}
}
