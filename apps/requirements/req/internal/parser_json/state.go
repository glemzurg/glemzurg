package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/state"
)

// stateInOut is a particular set of values in a state, distinct from all other states in the state.
type stateInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"` // Markdown.
	UmlComment string `json:"uml_comment"`
	// Nested.
	Actions []stateActionInOut `json:"actions"`
}

// ToRequirements converts the stateInOut to state.State.
func (s stateInOut) ToRequirements() state.State {
	st := state.State{
		Key:        s.Key,
		Name:       s.Name,
		Details:    s.Details,
		UmlComment: s.UmlComment,
		Actions:    nil,
	}
	for _, a := range s.Actions {
		st.Actions = append(st.Actions, a.ToRequirements())
	}
	return st
}

// FromRequirements creates a stateInOut from state.State.
func FromRequirementsState(s state.State) stateInOut {
	st := stateInOut{
		Key:        s.Key,
		Name:       s.Name,
		Details:    s.Details,
		UmlComment: s.UmlComment,
		Actions:    nil,
	}
	for _, a := range s.Actions {
		st.Actions = append(st.Actions, FromRequirementsStateAction(a))
	}
	return st
}
