package parser_json

// actionInOut is what happens in a transition between states.
type actionInOut struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Details    string   `json:"details,omitempty"`
	Requires   []string `json:"requires,omitempty"` // To enter this action.
	Guarantees []string `json:"guarantees,omitempty"`
}
