package parser_json

// domainInOut is a root category of the model.
type domainInOut struct {
	Key        string
	Name       string
	Details    string // Markdown.
	Realized   bool   // If this domain has no semantic model because it is existing already, so only design in this domain.
	UmlComment string
	// Nested.
	Subdomains []subdomainInOut
}
