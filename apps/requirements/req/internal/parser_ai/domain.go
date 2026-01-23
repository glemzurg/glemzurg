package parser_ai

// inputDomain represents a domain.json file.
type inputDomain struct {
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"`
	Realized   bool   `json:"realized,omitempty"`
	UmlComment string `json:"uml_comment,omitempty"`
}
