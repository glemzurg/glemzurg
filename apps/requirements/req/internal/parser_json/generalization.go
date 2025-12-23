package parser_json

// generalizationInOut is how two or more things in the system build on each other (like a super type and sub type).
type generalizationInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"` // Markdown.
	IsComplete bool   `json:"is_complete"`       // Are the specializations complete, or can an instantiation of this generalization exist without a specialization.
	IsStatic   bool   `json:"is_static"`         // Are the specializations static and unchanging or can they change during runtime.
	UmlComment string `json:"uml_comment,omitempty"`
}
