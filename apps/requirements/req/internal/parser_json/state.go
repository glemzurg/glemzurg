package parser_json

// stateInOut is a particular set of values in a state, distinct from all other states in the state.
type stateInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"` // Markdown.
	UmlComment string `json:"uml_comment,omitempty"`
	// Nested.
	Actions []stateActionInOut `json:"actions,omitempty"`
}
