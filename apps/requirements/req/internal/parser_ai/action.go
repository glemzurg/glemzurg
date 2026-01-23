package parser_ai

// inputAction represents an action JSON file.
type inputAction struct {
	Name       string   `json:"name"`
	Details    string   `json:"details,omitempty"`
	Requires   []string `json:"requires,omitempty"`
	Guarantees []string `json:"guarantees,omitempty"`
}
