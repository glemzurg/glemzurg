package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
)

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

// ToRequirements converts the transitionInOut to model_state.Transition.
func (t transitionInOut) ToRequirements() (model_state.Transition, error) {
	key, err := identity.ParseKey(t.Key)
	if err != nil {
		return model_state.Transition{}, err
	}

	eventKey, err := identity.ParseKey(t.EventKey)
	if err != nil {
		return model_state.Transition{}, err
	}

	// Handle optional pointer fields - empty string means nil
	var fromStateKey *identity.Key
	if t.FromStateKey != "" {
		k, err := identity.ParseKey(t.FromStateKey)
		if err != nil {
			return model_state.Transition{}, err
		}
		fromStateKey = &k
	}

	var guardKey *identity.Key
	if t.GuardKey != "" {
		k, err := identity.ParseKey(t.GuardKey)
		if err != nil {
			return model_state.Transition{}, err
		}
		guardKey = &k
	}

	var actionKey *identity.Key
	if t.ActionKey != "" {
		k, err := identity.ParseKey(t.ActionKey)
		if err != nil {
			return model_state.Transition{}, err
		}
		actionKey = &k
	}

	var toStateKey *identity.Key
	if t.ToStateKey != "" {
		k, err := identity.ParseKey(t.ToStateKey)
		if err != nil {
			return model_state.Transition{}, err
		}
		toStateKey = &k
	}

	return model_state.Transition{
		Key:          key,
		FromStateKey: fromStateKey,
		EventKey:     eventKey,
		GuardKey:     guardKey,
		ActionKey:    actionKey,
		ToStateKey:   toStateKey,
		UmlComment:   t.UmlComment,
	}, nil
}

// FromRequirements creates a transitionInOut from model_state.Transition.
func FromRequirementsTransition(t model_state.Transition) transitionInOut {
	// Handle optional pointer fields - nil means empty string
	var fromStateKey, guardKey, actionKey, toStateKey string
	if t.FromStateKey != nil {
		fromStateKey = t.FromStateKey.String()
	}
	if t.GuardKey != nil {
		guardKey = t.GuardKey.String()
	}
	if t.ActionKey != nil {
		actionKey = t.ActionKey.String()
	}
	if t.ToStateKey != nil {
		toStateKey = t.ToStateKey.String()
	}

	return transitionInOut{
		Key:          t.Key.String(),
		FromStateKey: fromStateKey,
		EventKey:     t.EventKey.String(),
		GuardKey:     guardKey,
		ActionKey:    actionKey,
		ToStateKey:   toStateKey,
		UmlComment:   t.UmlComment,
	}
}
