package parser_json

// useCase is a user story for the system.
type useCase struct {
	Key        string
	Name       string
	Details    string // Markdown.
	Level      string // How high cocept or tightly focused the user case is.
	ReadOnly   bool   // This is a user story that does not change the state of the system.
	UmlComment string
	// Nested.
	Actors    map[string]useCaseActor
	Scenarios []scenario
	// Helpful data.
	DomainKey string
}
