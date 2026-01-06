package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/state"

// eventInOut is what triggers a transition between states.
type eventInOut struct {
	Key        string                `json:"key"`
	Name       string                `json:"name"`
	Details    string                `json:"details"`
	Parameters []eventParameterInOut `json:"parameters"`
}

// ToRequirements converts the eventInOut to state.Event.
func (e eventInOut) ToRequirements() state.Event {
	event := state.Event{
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

// FromRequirements creates a eventInOut from state.Event.
func FromRequirementsEvent(e state.Event) eventInOut {
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
