package parser_json

// action is what happens in a transition between states.
type action struct {
	Key        string
	Name       string
	Details    string
	Requires   []string // To enter this action.
	Guarantees []string
}
