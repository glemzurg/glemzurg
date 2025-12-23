package parser_json

// actorInOut is a external user of this system, either a person or another system.
type actorInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"` // Markdown.
	Type       string `json:"type"`              // "person" or "system"
	UmlComment string `json:"uml_comment,omitempty"`
}
