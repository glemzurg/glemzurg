package parser_json

// transitionInOut is a move between two states.
type transitionInOut struct {
	Key          string
	FromStateKey string
	EventKey     string
	GuardKey     string
	ActionKey    string
	ToStateKey   string
	UmlComment   string
}
