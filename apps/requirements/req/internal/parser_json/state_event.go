package parser_json

// eventInOut is what triggers a transition between states.
type eventInOut struct {
	Key        string
	Name       string
	Details    string
	Parameters []eventParameterInOut
}
