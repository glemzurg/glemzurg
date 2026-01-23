package parser_ai_input

// inputQuery represents a query JSON file.
type inputQuery struct {
	Name       string   `json:"name"`
	Details    string   `json:"details,omitempty"`
	Requires   []string `json:"requires,omitempty"`
	Guarantees []string `json:"guarantees,omitempty"`
}
