package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"

// actionInOut is what happens in a transition between states.
type actionInOut struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Details    string   `json:"details"`
	Requires   []string `json:"requires"` // To enter this action.
	Guarantees []string `json:"guarantees"`
}

// ToRequirements converts the actionInOut to model_state.Action.
func (a actionInOut) ToRequirements() model_state.Action {
	return model_state.Action{
		Key:        a.Key,
		Name:       a.Name,
		Details:    a.Details,
		Requires:   a.Requires,
		Guarantees: a.Guarantees,
	}
}

// FromRequirements creates a actionInOut from model_state.Action.
func FromRequirementsAction(a model_state.Action) actionInOut {
	return actionInOut{
		Key:        a.Key,
		Name:       a.Name,
		Details:    a.Details,
		Requires:   a.Requires,
		Guarantees: a.Guarantees,
	}
}
