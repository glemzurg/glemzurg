package parser_json

// actorInOut is a external user of this system, either a person or another system.
type actorInOut struct {
	Key        string
	Name       string
	Details    string // Markdown.
	Type       string // "person" or "system"
	UmlComment string
	// Helpful data.
	ClassKeys []string // Classes that implement this actor.
}
