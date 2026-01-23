package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
)

// stateActionInOut is a action that triggers when a state is entered or exited or happens perpetually.
type stateActionInOut struct {
	Key       string `json:"key"`
	ActionKey string `json:"action_key"`
	When      string `json:"when"`
}

// ToRequirements converts the stateActionInOut to model_state.StateAction.
func (s stateActionInOut) ToRequirements() (model_state.StateAction, error) {
	key, err := identity.ParseKey(s.Key)
	if err != nil {
		return model_state.StateAction{}, err
	}

	actionKey, err := identity.ParseKey(s.ActionKey)
	if err != nil {
		return model_state.StateAction{}, err
	}

	return model_state.StateAction{
		Key:       key,
		ActionKey: actionKey,
		When:      s.When,
	}, nil
}

// FromRequirements creates a stateActionInOut from model_state.StateAction.
func FromRequirementsStateAction(s model_state.StateAction) stateActionInOut {
	return stateActionInOut{
		Key:       s.Key.String(),
		ActionKey: s.ActionKey.String(),
		When:      s.When,
	}
}
