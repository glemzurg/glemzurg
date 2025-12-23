package parser_json

// useCaseInOut is a user story for the system.
type useCaseInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"` // Markdown.
	Level      string `json:"level,omitempty"`   // How high cocept or tightly focused the user case is.
	ReadOnly   bool   `json:"read_only"`
	UmlComment string `json:"uml_comment,omitempty"`
	// Nested.
	Actors    map[string]useCaseActorInOut `json:"actors,omitempty"`
	Scenarios []scenarioInOut              `json:"scenarios,omitempty"`
}
