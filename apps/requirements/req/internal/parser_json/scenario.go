package parser_json

// scenario is a documented scenario for a use case, such as a sequence diagram.
type scenario struct {
	Key     string
	Name    string
	Details string // Markdown.
	Steps   node   // The "abstract syntax tree" of the scenario.
	// Nested.
	Objects []scenarioObject
}
