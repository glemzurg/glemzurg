package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"

// eventInOut is what triggers a transition between states.
type eventInOut struct {
	Key        string                `json:"key"`
	Name       string                `json:"name"`
	Details    string                `json:"details"`
	Parameters []eventParameterInOut `json:"parameters"`
}

// ToRequirements converts the eventInOut to model_state.Event.
func (e eventInOut) ToRequirements() model_state.Event {
	event := model_state.Event{
		Key:        e.Key,
		Name:       e.Name,
		Details:    e.Details,
		Parameters: nil,
	}
	for _, p := range e.Parameters {
		event.Parameters = append(event.Parameters, p.ToRequirements())
	}
	return event
}

// FromRequirements creates a eventInOut from model_state.Event.
func FromRequirementsEvent(e model_state.Event) eventInOut {
	event := eventInOut{
		Key:        e.Key,
		Name:       e.Name,
		Details:    e.Details,
		Parameters: nil,
	}
	for _, p := range e.Parameters {
		event.Parameters = append(event.Parameters, FromRequirementsEventParameter(p))
	}
	return event
}
