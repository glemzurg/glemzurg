package parser_json

// scenarioInOut is a documented scenario for a use case, such as a sequence diagram.
type scenarioInOut struct {
	Key     string
	Name    string
	Details string    // Markdown.
	Steps   nodeInOut // The "abstract syntax tree" of the scenario.
	// Nested.
	Objects []scenarioObjectInOut
}
