package parser_json

// useCaseInOut is a user story for the system.
type useCaseInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"` // Markdown.
	Level      string `json:"level"`   // How high cocept or tightly focused the user case is.
	ReadOnly   bool   `json:"read_only"`
	UmlComment string `json:"uml_comment"`
	// Nested.
	Actors    map[string]useCaseActorInOut `json:"actors"`
	Scenarios []scenarioInOut              `json:"scenarios"`
}
