package parser_json

// actionInOut is what happens in a transition between states.
type actionInOut struct {
	Key        string
	Name       string
	Details    string
	Requires   []string // To enter this action.
	Guarantees []string
}
