package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// stateInOut is a particular set of values in a state, distinct from all other states in the state.
type stateInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"` // Markdown.
	UmlComment string `json:"uml_comment"`
	// Nested.
	Actions []stateActionInOut `json:"actions"`
}

// ToRequirements converts the stateInOut to requirements.State.
func (s stateInOut) ToRequirements() requirements.State {
	state := requirements.State{
		Key:        s.Key,
		Name:       s.Name,
		Details:    s.Details,
		UmlComment: s.UmlComment,
	}
	for _, a := range s.Actions {
		state.Actions = append(state.Actions, a.ToRequirements())
	}
	return state
}

// FromRequirements creates a stateInOut from requirements.State.
func FromRequirementsState(s requirements.State) stateInOut {
	state := stateInOut{
		Key:        s.Key,
		Name:       s.Name,
		Details:    s.Details,
		UmlComment: s.UmlComment,
	}
	for _, a := range s.Actions {
		state.Actions = append(state.Actions, FromRequirementsStateAction(a))
	}
	return state
}
