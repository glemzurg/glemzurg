package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// transitionInOut is a move between two states.
type transitionInOut struct {
	Key          string `json:"key"`
	FromStateKey string `json:"from_state_key"`
	EventKey     string `json:"event_key"`
	GuardKey     string `json:"guard_key"`
	ActionKey    string `json:"action_key"`
	ToStateKey   string `json:"to_state_key"`
	UmlComment   string `json:"uml_comment"`
}

// ToRequirements converts the transitionInOut to requirements.Transition.
func (t transitionInOut) ToRequirements() requirements.Transition {
	return requirements.Transition{
		Key:          t.Key,
		FromStateKey: t.FromStateKey,
		EventKey:     t.EventKey,
		GuardKey:     t.GuardKey,
		ActionKey:    t.ActionKey,
		ToStateKey:   t.ToStateKey,
		UmlComment:   t.UmlComment,
	}
}

// FromRequirements creates a transitionInOut from requirements.Transition.
func FromRequirementsTransition(t requirements.Transition) transitionInOut {
	return transitionInOut{
		Key:          t.Key,
		FromStateKey: t.FromStateKey,
		EventKey:     t.EventKey,
		GuardKey:     t.GuardKey,
		ActionKey:    t.ActionKey,
		ToStateKey:   t.ToStateKey,
		UmlComment:   t.UmlComment,
	}
}
