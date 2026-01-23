package parser_ai_input

// inputActor represents an actor JSON file.
type inputActor struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Details    string `json:"details,omitempty"`
	UmlComment string `json:"uml_comment,omitempty"`
}
