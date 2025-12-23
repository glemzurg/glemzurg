package parser_json

// transitionInOut is a move between two states.
type transitionInOut struct {
	Key          string `json:"key"`
	FromStateKey string `json:"from_state_key"`
	EventKey     string `json:"event_key"`
	GuardKey     string `json:"guard_key"`
	ActionKey    string `json:"action_key"`
	ToStateKey   string `json:"to_state_key"`
	UmlComment   string `json:"uml_comment"`
}
