package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/state"

// eventParameterInOut is a parameter for events.
type eventParameterInOut struct {
	Name   string `json:"name"`
	Source string `json:"source"` // Where the values for this parameter are coming from.
}

// ToRequirements converts the eventParameterInOut to state.EventParameter.
func (e eventParameterInOut) ToRequirements() state.EventParameter {
	return state.EventParameter{
		Name:   e.Name,
		Source: e.Source,
	}
}

// FromRequirements creates a eventParameterInOut from state.EventParameter.
func FromRequirementsEventParameter(e state.EventParameter) eventParameterInOut {
	return eventParameterInOut{
		Name:   e.Name,
		Source: e.Source,
	}
}
