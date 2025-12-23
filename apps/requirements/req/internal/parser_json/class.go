package parser_json

// class is a thing in the system.
type class struct {
	Key             string
	Name            string
	Details         string // Markdown.
	ActorKey        string // If this class is an Actor this is the key of that actor.
	SuperclassOfKey string // If this class is part of a generalization as the superclass.
	SubclassOfKey   string // If this class is part of a generalization as a subclass.
	UmlComment      string
	// Nested.
	Attributes  []attribute
	States      []state
	Events      []event
	Guards      []guard
	Actions     []action
	Transitions []transition
}
