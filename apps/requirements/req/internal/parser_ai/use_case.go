package parser_ai

// inputUseCaseActor represents an actor reference within a use case.
// The map key is the class key of the actor class.
type inputUseCaseActor struct {
	UmlComment string `json:"uml_comment,omitempty"`
}

// inputUseCase represents a use case JSON file.
// Use cases are user stories for the system at various levels (sky/sea/mud).
type inputUseCase struct {
	Name       string                       `json:"name"`
	Details    string                       `json:"details,omitempty"`
	Level      string                       `json:"level"`
	ReadOnly   bool                         `json:"read_only,omitempty"`
	UMLComment string                       `json:"uml_comment,omitempty"`
	Actors     map[string]*inputUseCaseActor `json:"actors,omitempty"`

	// Children (not from JSON, populated during directory traversal)
	Scenarios map[string]*inputScenario `json:"-"`
}
