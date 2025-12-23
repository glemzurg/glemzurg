package parser_json

// event is what triggers a transition between states.
type event struct {
	Key        string
	Name       string
	Details    string
	Parameters []eventParameter
}
