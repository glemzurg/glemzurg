package parser_json

// actor is a external user of this system, either a person or another system.
type actor struct {
	Key        string
	Name       string
	Details    string // Markdown.
	Type       string // "person" or "system"
	UmlComment string
	// Helpful data.
	ClassKeys []string // Classes that implement this actor.
}
