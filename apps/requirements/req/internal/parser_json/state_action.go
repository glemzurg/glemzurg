package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// actionInOut is what happens in a transition between states.
type actionInOut struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Details    string   `json:"details"`
	Requires   []string `json:"requires"` // To enter this action.
	Guarantees []string `json:"guarantees"`
}

// ToRequirements converts the actionInOut to requirements.Action.
func (a actionInOut) ToRequirements() requirements.Action {
	return requirements.Action{
		Key:             a.Key,
		Name:            a.Name,
		Details:         a.Details,
		Requires:        a.Requires,
		Guarantees:      a.Guarantees,
		FromTransitions: nil, // Not stored in JSON
		FromStates:      nil, // Not stored in JSON
	}
}

// FromRequirements creates a actionInOut from requirements.Action.
func FromRequirementsAction(a requirements.Action) actionInOut {
	return actionInOut{
		Key:        a.Key,
		Name:       a.Name,
		Details:    a.Details,
		Requires:   a.Requires,
		Guarantees: a.Guarantees,
	}
}
