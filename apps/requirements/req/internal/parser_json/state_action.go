package parser_json

// actionInOut is what happens in a transition between states.
type actionInOut struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Details    string   `json:"details"`
	Requires   []string `json:"requires"` // To enter this action.
	Guarantees []string `json:"guarantees"`
}
