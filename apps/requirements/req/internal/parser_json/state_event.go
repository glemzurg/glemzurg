package parser_json

// eventInOut is what triggers a transition between states.
type eventInOut struct {
	Key        string                `json:"key"`
	Name       string                `json:"name"`
	Details    string                `json:"details"`
	Parameters []eventParameterInOut `json:"parameters"`
}
