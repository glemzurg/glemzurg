package parser_json

// generalization is how two or more things in the system build on each other (like a super type and sub type).
type generalization struct {
	Key        string
	Name       string
	Details    string // Markdown.
	IsComplete bool   // Are the specializations complete, or can an instantiation of this generalization exist without a specialization.
	IsStatic   bool   // Are the specializations static and unchanging or can they change during runtime.
	UmlComment string
}
