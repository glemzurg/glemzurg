package parser_json

// state is a particular set of values in a state, distinct from all other states in the state.
type state struct {
	Key        string
	Name       string
	Details    string // Markdown.
	UmlComment string
	// Nested.
	Actions []stateAction
}
