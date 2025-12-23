package parser_json

// eventInOut is what triggers a transition between states.
type eventInOut struct {
	Key        string                `json:"key"`
	Name       string                `json:"name"`
	Details    string                `json:"details,omitempty"`
	Parameters []eventParameterInOut `json:"parameters,omitempty"`
}
