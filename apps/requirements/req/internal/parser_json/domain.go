package parser_json

// domainInOut is a root category of the model.
type domainInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"` // Markdown.
	Realized   bool   `json:"realized"`
	UmlComment string `json:"uml_comment"`
	// Nested.
	Subdomains []subdomainInOut `json:"subdomains"`
}
