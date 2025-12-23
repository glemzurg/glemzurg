package parser_json

// guard is a constraint on an event in a state machine.
type guard struct {
	Key     string
	Name    string // A simple unique name for a guard, for internal use.
	Details string // How the details of the guard are represented, what shows in the uml.
}
