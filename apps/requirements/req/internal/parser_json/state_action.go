package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
)

// actionInOut is what happens in a transition between states.
type actionInOut struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Details    string   `json:"details"`
	Requires   []string `json:"requires"` // To enter this action.
	Guarantees []string `json:"guarantees"`
}

// ToRequirements converts the actionInOut to model_state.Action.
func (a actionInOut) ToRequirements() (model_state.Action, error) {
	key, err := identity.ParseKey(a.Key)
	if err != nil {
		return model_state.Action{}, err
	}

	return model_state.Action{
		Key:        key,
		Name:       a.Name,
		Details:    a.Details,
		Requires:   a.Requires,
		Guarantees: a.Guarantees,
	}, nil
}

// FromRequirements creates a actionInOut from model_state.Action.
func FromRequirementsAction(a model_state.Action) actionInOut {
	return actionInOut{
		Key:        a.Key.String(),
		Name:       a.Name,
		Details:    a.Details,
		Requires:   a.Requires,
		Guarantees: a.Guarantees,
	}
}
