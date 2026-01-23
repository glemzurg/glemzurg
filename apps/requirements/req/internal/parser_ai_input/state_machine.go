package parser_ai_input

// inputStateAction represents an action reference in a state.
type inputStateAction struct {
	ActionKey string `json:"action_key"`
	When      string `json:"when"` // "entry", "exit", or "do"
}

// inputState represents a state within a state machine.
type inputState struct {
	Name       string             `json:"name"`
	Details    string             `json:"details,omitempty"`
	UmlComment string             `json:"uml_comment,omitempty"`
	Actions    []inputStateAction `json:"actions,omitempty"`
}

// inputEventParameter represents a parameter for an event.
type inputEventParameter struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

// inputEvent represents an event in a state machine.
type inputEvent struct {
	Name       string                `json:"name"`
	Details    string                `json:"details,omitempty"`
	Parameters []inputEventParameter `json:"parameters,omitempty"`
}

// inputGuard represents a guard condition in a state machine.
type inputGuard struct {
	Name    string `json:"name"`
	Details string `json:"details"`
}

// inputTransition represents a transition in a state machine.
type inputTransition struct {
	FromStateKey *string `json:"from_state_key,omitempty"`
	ToStateKey   *string `json:"to_state_key,omitempty"`
	EventKey     string  `json:"event_key"`
	GuardKey     *string `json:"guard_key,omitempty"`
	ActionKey    *string `json:"action_key,omitempty"`
	UmlComment   string  `json:"uml_comment,omitempty"`
}

// inputStateMachine represents a state_machine.json file.
type inputStateMachine struct {
	States      map[string]inputState `json:"states,omitempty"`
	Events      map[string]inputEvent `json:"events,omitempty"`
	Guards      map[string]inputGuard `json:"guards,omitempty"`
	Transitions []inputTransition     `json:"transitions,omitempty"`
}
