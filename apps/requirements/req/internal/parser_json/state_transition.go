package parser_json

// transition is a move between two states.
type transition struct {
	Key          string
	FromStateKey string
	EventKey     string
	GuardKey     string
	ActionKey    string
	ToStateKey   string
	UmlComment   string
}
