package parser_json

// scenarioInOut is a documented scenario for a use case, such as a sequence diagram.
type scenarioInOut struct {
	Key     string    `json:"key"`
	Name    string    `json:"name"`
	Details string    `json:"details,omitempty"` // Markdown.
	Steps   nodeInOut `json:"steps"`             // The "abstract syntax tree" of the scenario.
	// Nested.
	Objects []scenarioObjectInOut `json:"objects,omitempty"`
}
