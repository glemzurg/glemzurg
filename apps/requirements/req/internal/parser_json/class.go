package parser_json

// classInOut is a thing in the system.
type classInOut struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Details         string `json:"details"`           // Markdown.
	ActorKey        string `json:"actor_key"`         // If this class is an Actor this is the key of that actor.
	SuperclassOfKey string `json:"superclass_of_key"` // If this class is part of a generalization as the superclass.
	SubclassOfKey   string `json:"subclass_of_key"`   // If this class is part of a generalization as a subclass.
	UmlComment      string `json:"uml_comment"`
	// Nested.
	Attributes  []attributeInOut  `json:"attributes"`
	States      []stateInOut      `json:"states"`
	Events      []eventInOut      `json:"events"`
	Guards      []guardInOut      `json:"guards"`
	Actions     []actionInOut     `json:"actions"`
	Transitions []transitionInOut `json:"transitions"`
}
