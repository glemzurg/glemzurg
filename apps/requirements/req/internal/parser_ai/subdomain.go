package parser_ai

// inputSubdomain represents a subdomain.json file.
type inputSubdomain struct {
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"`
	UmlComment string `json:"uml_comment,omitempty"`
}
