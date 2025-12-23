package parser_json

// guardInOut is a constraint on an event in a state machine.
type guardInOut struct {
	Key     string `json:"key"`
	Name    string `json:"name"`    // A simple unique name for a guard, for internal use.
	Details string `json:"details"` // How the details of the guard are represented, what shows in the uml.
}
