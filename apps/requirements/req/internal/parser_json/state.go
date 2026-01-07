package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"

// stateInOut is a particular set of values in a state, distinct from all other states in the state.
type stateInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"` // Markdown.
	UmlComment string `json:"uml_comment"`
	// Nested.
	Actions []stateActionInOut `json:"actions"`
}

// ToRequirements converts the stateInOut to model_state.State.
func (s stateInOut) ToRequirements() model_state.State {
	state := model_state.State{
		Key:        s.Key,
		Name:       s.Name,
		Details:    s.Details,
		UmlComment: s.UmlComment,
		Actions:    nil,
	}
	for _, a := range s.Actions {
		state.Actions = append(state.Actions, a.ToRequirements())
	}
	return state
}

// FromRequirements creates a stateInOut from model_state.State.
func FromRequirementsState(s model_state.State) stateInOut {
	state := stateInOut{
		Key:        s.Key,
		Name:       s.Name,
		Details:    s.Details,
		UmlComment: s.UmlComment,
		Actions:    nil,
	}
	for _, a := range s.Actions {
		state.Actions = append(state.Actions, FromRequirementsStateAction(a))
	}
	return state
}
