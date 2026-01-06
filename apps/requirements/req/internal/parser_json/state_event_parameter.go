package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// eventParameterInOut is a parameter for events.
type eventParameterInOut struct {
	Name   string `json:"name"`
	Source string `json:"source"` // Where the values for this parameter are coming from.
}

// ToRequirements converts the eventParameterInOut to requirements.EventParameter.
func (e eventParameterInOut) ToRequirements() requirements.EventParameter {
	return requirements.EventParameter{
		Name:   e.Name,
		Source: e.Source,
	}
}

// FromRequirements creates a eventParameterInOut from requirements.EventParameter.
func FromRequirementsEventParameter(e requirements.EventParameter) eventParameterInOut {
	return eventParameterInOut{
		Name:   e.Name,
		Source: e.Source,
	}
}
