package parser_json

// guardInOut is a constraint on an event in a state machine.
type guardInOut struct {
	Key     string
	Name    string // A simple unique name for a guard, for internal use.
	Details string // How the details of the guard are represented, what shows in the uml.
}
