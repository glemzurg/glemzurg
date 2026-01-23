package parser_ai_input

// This file contains the Go structs for parsing JSON input from AI-generated model files.
// These structs are separate from req_model structs to allow for:
// 1. Optimized input shapes (e.g., superclass/subclass in generalization)
// 2. Distinct error handling with unique error numbers
// 3. Clear separation between input validation and canonical model representation

// InputModel represents the model.json file.
type InputModel struct {
	Name    string `json:"name"`
	Details string `json:"details,omitempty"`
}

// InputActor represents an actor JSON file.
type InputActor struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Details    string `json:"details,omitempty"`
	UmlComment string `json:"uml_comment,omitempty"`
}

// InputDomain represents a domain.json file.
type InputDomain struct {
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"`
	Realized   bool   `json:"realized,omitempty"`
	UmlComment string `json:"uml_comment,omitempty"`
}

// InputSubdomain represents a subdomain.json file.
type InputSubdomain struct {
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"`
	UmlComment string `json:"uml_comment,omitempty"`
}

// InputAttribute represents an attribute within a class.
type InputAttribute struct {
	Name             string `json:"name"`
	DataTypeRules    string `json:"data_type_rules,omitempty"`
	Details          string `json:"details,omitempty"`
	DerivationPolicy string `json:"derivation_policy,omitempty"`
	Nullable         bool   `json:"nullable,omitempty"`
	UmlComment       string `json:"uml_comment,omitempty"`
}

// InputClass represents a class.json file.
type InputClass struct {
	Name       string                    `json:"name"`
	Details    string                    `json:"details,omitempty"`
	ActorKey   string                    `json:"actor_key,omitempty"`
	UmlComment string                    `json:"uml_comment,omitempty"`
	Attributes map[string]InputAttribute `json:"attributes,omitempty"`
	Indexes    [][]string                `json:"indexes,omitempty"`
}

// InputAssociation represents an association JSON file.
type InputAssociation struct {
	Name                string  `json:"name"`
	Details             string  `json:"details,omitempty"`
	FromClassKey        string  `json:"from_class_key"`
	FromMultiplicity    string  `json:"from_multiplicity"`
	ToClassKey          string  `json:"to_class_key"`
	ToMultiplicity      string  `json:"to_multiplicity"`
	AssociationClassKey *string `json:"association_class_key,omitempty"`
	UmlComment          string  `json:"uml_comment,omitempty"`
}

// InputStateAction represents an action reference in a state.
type InputStateAction struct {
	ActionKey string `json:"action_key"`
	When      string `json:"when"` // "entry", "exit", or "do"
}

// InputState represents a state within a state machine.
type InputState struct {
	Name       string             `json:"name"`
	Details    string             `json:"details,omitempty"`
	UmlComment string             `json:"uml_comment,omitempty"`
	Actions    []InputStateAction `json:"actions,omitempty"`
}

// InputEventParameter represents a parameter for an event.
type InputEventParameter struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

// InputEvent represents an event in a state machine.
type InputEvent struct {
	Name       string                `json:"name"`
	Details    string                `json:"details,omitempty"`
	Parameters []InputEventParameter `json:"parameters,omitempty"`
}

// InputGuard represents a guard condition in a state machine.
type InputGuard struct {
	Name    string `json:"name"`
	Details string `json:"details"`
}

// InputTransition represents a transition in a state machine.
type InputTransition struct {
	FromStateKey *string `json:"from_state_key,omitempty"`
	ToStateKey   *string `json:"to_state_key,omitempty"`
	EventKey     string  `json:"event_key"`
	GuardKey     *string `json:"guard_key,omitempty"`
	ActionKey    *string `json:"action_key,omitempty"`
	UmlComment   string  `json:"uml_comment,omitempty"`
}

// InputStateMachine represents a state_machine.json file.
type InputStateMachine struct {
	States      map[string]InputState `json:"states,omitempty"`
	Events      map[string]InputEvent `json:"events,omitempty"`
	Guards      map[string]InputGuard `json:"guards,omitempty"`
	Transitions []InputTransition     `json:"transitions,omitempty"`
}

// InputAction represents an action JSON file.
type InputAction struct {
	Name       string   `json:"name"`
	Details    string   `json:"details,omitempty"`
	Requires   []string `json:"requires,omitempty"`
	Guarantees []string `json:"guarantees,omitempty"`
}

// InputQuery represents a query JSON file.
type InputQuery struct {
	Name       string   `json:"name"`
	Details    string   `json:"details,omitempty"`
	Requires   []string `json:"requires,omitempty"`
	Guarantees []string `json:"guarantees,omitempty"`
}

// InputGeneralization represents a generalization JSON file.
type InputGeneralization struct {
	Name          string   `json:"name"`
	Details       string   `json:"details,omitempty"`
	SuperclassKey string   `json:"superclass_key"`
	SubclassKeys  []string `json:"subclass_keys"`
	IsComplete    bool     `json:"is_complete,omitempty"`
	IsStatic      bool     `json:"is_static,omitempty"`
	UmlComment    string   `json:"uml_comment,omitempty"`
}
