package parser_json

// classInOut is a thing in the system.
type classInOut struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Details         string `json:"details,omitempty"`           // Markdown.
	ActorKey        string `json:"actor_key,omitempty"`         // If this class is an Actor this is the key of that actor.
	SuperclassOfKey string `json:"superclass_of_key,omitempty"` // If this class is part of a generalization as the superclass.
	SubclassOfKey   string `json:"subclass_of_key,omitempty"`   // If this class is part of a generalization as a subclass.
	UmlComment      string `json:"uml_comment,omitempty"`
	// Nested.
	Attributes  []attributeInOut  `json:"attributes,omitempty"`
	States      []stateInOut      `json:"states,omitempty"`
	Events      []eventInOut      `json:"events,omitempty"`
	Guards      []guardInOut      `json:"guards,omitempty"`
	Actions     []actionInOut     `json:"actions,omitempty"`
	Transitions []transitionInOut `json:"transitions,omitempty"`
}
