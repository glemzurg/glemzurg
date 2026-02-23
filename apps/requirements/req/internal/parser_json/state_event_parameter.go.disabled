package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

// eventParameterInOut is a parameter for events.
type eventParameterInOut struct {
	Name   string `json:"name"`
	Source string `json:"source"` // Where the values for this parameter are coming from.
}

// ToRequirements converts the eventParameterInOut to model_state.EventParameter.
func (e eventParameterInOut) ToRequirements() model_state.EventParameter {
	return model_state.EventParameter{
		Name:   e.Name,
		Source: e.Source,
	}
}

// FromRequirements creates a eventParameterInOut from model_state.EventParameter.
func FromRequirementsEventParameter(e model_state.EventParameter) eventParameterInOut {
	return eventParameterInOut{
		Name:   e.Name,
		Source: e.Source,
	}
}
